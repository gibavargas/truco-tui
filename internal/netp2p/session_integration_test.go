package netp2p

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"truco-tui/internal/truco"
)

func waitUntil(t *testing.T, timeout time.Duration, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout waiting: %s", msg)
}

func waitEventContains(t *testing.T, ch <-chan string, timeout time.Duration, substr string) string {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case ev := <-ch:
			if strings.Contains(ev, substr) {
				return ev
			}
		case <-deadline:
			t.Fatalf("timeout waiting event containing %q", substr)
		}
	}
}

func waitAction(t *testing.T, ch <-chan ClientAction, timeout time.Duration) ClientAction {
	t.Helper()
	select {
	case a := <-ch:
		return a
	case <-time.After(timeout):
		t.Fatalf("timeout waiting host action")
		return ClientAction{}
	}
}

func waitState(t *testing.T, ch <-chan truco.Snapshot, timeout time.Duration) truco.Snapshot {
	t.Helper()
	select {
	case s := <-ch:
		return s
	case <-time.After(timeout):
		t.Fatalf("timeout waiting state update")
		return truco.Snapshot{}
	}
}

func TestHostClientGameFlow(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	cli, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	defer cli.Close()

	waitUntil(t, 2*time.Second, host.IsFull, "host lobby full")

	if err := host.StartGame(); err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	waitEventContains(t, cli.Events(), 2*time.Second, "Partida iniciada")

	g, err := truco.NewGame([]string{"Host", "Guest"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	s := g.Snapshot(1)
	host.SendGameStateToSeat(1, Message{Type: "game_state", State: &s})

	recv := waitState(t, cli.StateUpdates(), 2*time.Second)
	if recv.NumPlayers != 2 {
		t.Fatalf("recv.NumPlayers = %d, want 2", recv.NumPlayers)
	}

	if err := cli.SendGameAction("truco", 0); err != nil {
		t.Fatalf("SendGameAction: %v", err)
	}
	a := waitAction(t, host.Actions(), 2*time.Second)
	if a.Seat != 1 || a.Action != "truco" {
		t.Fatalf("unexpected host action: %+v", a)
	}

	host.SendHostChat("hello")
	waitEventContains(t, cli.Events(), 2*time.Second, "[chat] Host: hello")
}

func TestJoinSessionInvalidToken(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	inv, err := DecodeInviteKey(key)
	if err != nil {
		t.Fatalf("DecodeInviteKey: %v", err)
	}
	inv.Token = "invalid-token"
	badKey, err := EncodeInviteKey(inv)
	if err != nil {
		t.Fatalf("EncodeInviteKey: %v", err)
	}

	if _, err := JoinSession(badKey, "Guest", "auto"); err == nil {
		t.Fatalf("expected token validation error")
	} else if !strings.Contains(strings.ToLower(err.Error()), "token") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJoinSessionWithWildcardInviteFallsBackToLoopback(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	inv, err := DecodeInviteKey(key)
	if err != nil {
		t.Fatalf("DecodeInviteKey: %v", err)
	}
	_, port, err := net.SplitHostPort(inv.Addr)
	if err != nil {
		t.Fatalf("SplitHostPort(%q): %v", inv.Addr, err)
	}
	inv.Addr = net.JoinHostPort("::", port)

	wildKey, err := EncodeInviteKey(inv)
	if err != nil {
		t.Fatalf("EncodeInviteKey: %v", err)
	}
	cli, err := JoinSession(wildKey, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession with wildcard invite should fallback: %v", err)
	}
	defer cli.Close()
	waitUntil(t, 2*time.Second, host.IsFull, "host lobby full with wildcard fallback")
}

func TestJoinRejectsProtocolVersionMismatch(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	inv, err := DecodeInviteKey(key)
	if err != nil {
		t.Fatalf("DecodeInviteKey: %v", err)
	}
	conn, err := dialSessionConn(inv, 2*time.Second)
	if err != nil {
		t.Fatalf("dialSessionConn: %v", err)
	}
	defer closeConnWithLog(conn, "test protocol mismatch")
	reader := newConnReader(conn)

	if err := writeMessage(conn, Message{
		Type:            "join",
		ProtocolVersion: protocolVersion + 1,
		Token:           inv.Token,
		Name:            "Guest",
		DesiredRole:     "auto",
	}); err != nil {
		t.Fatalf("writeMessage(join mismatch): %v", err)
	}

	msg, err := readMessage(conn, reader)
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}
	if msg.Type != "error" {
		t.Fatalf("response type = %q, want error", msg.Type)
	}
	if !strings.Contains(strings.ToLower(msg.Error), "protocolo") {
		t.Fatalf("unexpected protocol mismatch error: %q", msg.Error)
	}
}

func TestJoinAcceptsPreviousProtocolVersion(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	inv, err := DecodeInviteKey(key)
	if err != nil {
		t.Fatalf("DecodeInviteKey: %v", err)
	}
	conn, err := dialSessionConn(inv, 2*time.Second)
	if err != nil {
		t.Fatalf("dialSessionConn: %v", err)
	}
	defer closeConnWithLog(conn, "test previous protocol version")
	reader := newConnReader(conn)

	if err := writeMessage(conn, Message{
		Type:            "join",
		ProtocolVersion: protocolVersion - 1,
		Token:           inv.Token,
		Name:            "Guest",
		DesiredRole:     "auto",
	}); err != nil {
		t.Fatalf("writeMessage(join previous): %v", err)
	}

	msg, err := readMessage(conn, reader)
	if err != nil {
		t.Fatalf("readMessage: %v", err)
	}
	if msg.Type != "join_ok" {
		t.Fatalf("response type = %q, want join_ok", msg.Type)
	}
	if msg.ProtocolVersion != protocolVersion-1 {
		t.Fatalf("protocol version = %d, want %d", msg.ProtocolVersion, protocolVersion-1)
	}
}

func TestJoinSessionFallsBackToPreviousProtocolVersion(t *testing.T) {
	tlsCfg, fingerprint, _, err := buildTLSConfig("fallback-protocol-seed")
	if err != nil {
		t.Fatalf("buildTLSConfig: %v", err)
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", tlsCfg)
	if err != nil {
		t.Fatalf("tls.Listen: %v", err)
	}
	defer ln.Close()

	seen := make(chan int, 4)
	handled := make(chan struct{})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer closeConnWithLog(conn, "fallback server conn")
				reader := newConnReader(conn)
				msg, err := readMessage(conn, reader)
				if err != nil {
					return
				}
				seen <- msg.ProtocolVersion
				if msg.ProtocolVersion == protocolVersion {
					_ = writeMessage(conn, Message{Type: "error", Error: fmt.Sprintf("versão de protocolo incompatível (cliente=%d, esperado=%d)", msg.ProtocolVersion, protocolVersion-1)})
					return
				}
				if msg.ProtocolVersion != protocolVersion-1 {
					_ = writeMessage(conn, Message{Type: "error", Error: fmt.Sprintf("versão inesperada: %d", msg.ProtocolVersion)})
					return
				}
				if err := writeMessage(conn, Message{
					Type:            "join_ok",
					ProtocolVersion: protocolVersion - 1,
					Assigned:        1,
					NumPlayers:      2,
					Slots:           []string{"Host", "Guest"},
					SessionID:       "session-legacy",
				}); err != nil {
					return
				}
				next, err := readMessage(conn, reader)
				if err != nil {
					return
				}
				seen <- next.ProtocolVersion
				close(handled)
				return
			}(conn)
		}
	}()

	inv := InviteKey{
		Addr:             ln.Addr().String(),
		Token:            "legacy-token",
		Fingerprint:      fingerprint,
		Transport:        "tcp_tls",
		TransportVersion: 1,
	}
	key, err := EncodeInviteKey(inv)
	if err != nil {
		t.Fatalf("EncodeInviteKey: %v", err)
	}

	cli, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession fallback: %v", err)
	}
	defer cli.Close()

	if err := cli.SendChat("hello"); err != nil {
		t.Fatalf("SendChat: %v", err)
	}

	select {
	case <-handled:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting fallback server")
	}

	var versions []int
	close(seen)
	for v := range seen {
		versions = append(versions, v)
	}
	if len(versions) < 3 {
		t.Fatalf("seen versions = %v, want at least join retry and chat", versions)
	}
	if versions[0] != protocolVersion || versions[1] != protocolVersion-1 {
		t.Fatalf("seen versions = %v, want retry from %d to %d", versions, protocolVersion, protocolVersion-1)
	}
	if versions[2] != protocolVersion-1 {
		t.Fatalf("client follow-up version = %d, want %d", versions[2], protocolVersion-1)
	}
}

func TestHostNetworkStateTracksMixedProtocolSession(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 4)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	inv, err := DecodeInviteKey(key)
	if err != nil {
		t.Fatalf("DecodeInviteKey: %v", err)
	}
	legacyConn, err := dialSessionConn(inv, 2*time.Second)
	if err != nil {
		t.Fatalf("dialSessionConn: %v", err)
	}
	defer closeConnWithLog(legacyConn, "test mixed legacy client")
	legacyReader := newConnReader(legacyConn)

	if err := writeMessage(legacyConn, Message{
		Type:            "join",
		ProtocolVersion: protocolVersion - 1,
		Token:           inv.Token,
		Name:            "Legacy",
		DesiredRole:     "auto",
	}); err != nil {
		t.Fatalf("writeMessage(join legacy): %v", err)
	}
	joinOK, err := readMessage(legacyConn, legacyReader)
	if err != nil {
		t.Fatalf("readMessage(join_ok legacy): %v", err)
	}
	if joinOK.Type != "join_ok" {
		t.Fatalf("response type = %q, want join_ok", joinOK.Type)
	}

	modern, err := JoinSession(key, "Modern", "auto")
	if err != nil {
		t.Fatalf("JoinSession modern: %v", err)
	}
	defer modern.Close()

	waitUntil(t, 2*time.Second, func() bool {
		versions := host.SeatProtocolVersions()
		return versions[0] == protocolVersion &&
			versions[joinOK.Assigned] == protocolVersion-1 &&
			versions[modern.AssignedSeat()] == protocolVersion
	}, "host protocol map to include mixed client versions")

	if host.TransportMode() != "tcp_tls" {
		t.Fatalf("transport = %q, want tcp_tls", host.TransportMode())
	}
	versions := host.SeatProtocolVersions()
	if versions[0] != protocolVersion {
		t.Fatalf("seat 0 protocol = %d, want %d", versions[0], protocolVersion)
	}
	if versions[joinOK.Assigned] != protocolVersion-1 {
		t.Fatalf("legacy seat protocol = %d, want %d", versions[joinOK.Assigned], protocolVersion-1)
	}
	if versions[modern.AssignedSeat()] != protocolVersion {
		t.Fatalf("modern seat protocol = %d, want %d", versions[modern.AssignedSeat()], protocolVersion)
	}
	mutated := host.SeatProtocolVersions()
	mutated[0] = 99
	if host.SeatProtocolVersions()[0] != protocolVersion {
		t.Fatalf("SeatProtocolVersions should return a copy")
	}
}

func TestClientNetworkStateGettersReflectNegotiatedProtocol(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	cli, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	defer cli.Close()

	if cli.TransportMode() != "tcp_tls" {
		t.Fatalf("transport = %q, want tcp_tls", cli.TransportMode())
	}
	if cli.WireProtocolVersion() != protocolVersion {
		t.Fatalf("wire protocol = %d, want %d", cli.WireProtocolVersion(), protocolVersion)
	}
	versions := SupportedProtocolVersions()
	if len(versions) < 2 || versions[0] != protocolVersion || versions[1] != protocolVersion-1 {
		t.Fatalf("supported versions = %v, want [%d %d]", versions, protocolVersion, protocolVersion-1)
	}
	versions[0] = 77
	if SupportedProtocolVersions()[0] != protocolVersion {
		t.Fatalf("SupportedProtocolVersions should return a copy")
	}
}

func TestClientReconnectPreservesNegotiatedProtocolVersion(t *testing.T) {
	tlsCfg, fingerprint, _, err := buildTLSConfig("reconnect-protocol-seed")
	if err != nil {
		t.Fatalf("buildTLSConfig: %v", err)
	}
	ln, err := tls.Listen("tcp", "127.0.0.1:0", tlsCfg)
	if err != nil {
		t.Fatalf("tls.Listen: %v", err)
	}
	defer ln.Close()

	serverErr := make(chan error, 4)
	initialConnReady := make(chan net.Conn, 1)
	reconnectJoinVersion := make(chan int, 1)
	reconnectChatVersion := make(chan int, 1)
	go func() {
		reportErr := func(err error) {
			select {
			case serverErr <- err:
			default:
			}
		}

		conn, err := ln.Accept()
		if err != nil {
			reportErr(err)
			return
		}
		reader := newConnReader(conn)
		first, err := readMessage(conn, reader)
		if err != nil {
			reportErr(err)
			return
		}
		if first.ProtocolVersion != protocolVersion {
			reportErr(fmt.Errorf("first attempt protocol=%d, want %d", first.ProtocolVersion, protocolVersion))
			return
		}
		if err := writeMessage(conn, Message{
			Type:  "error",
			Error: fmt.Sprintf("versão de protocolo incompatível (cliente=%d, esperado=%d)", first.ProtocolVersion, protocolVersion-1),
		}); err != nil {
			reportErr(err)
		}
		closeConnWithLog(conn, "legacy reconnect mismatch")

		conn, err = ln.Accept()
		if err != nil {
			reportErr(err)
			return
		}
		reader = newConnReader(conn)
		legacyJoin, err := readMessage(conn, reader)
		if err != nil {
			reportErr(err)
			return
		}
		if legacyJoin.ProtocolVersion != protocolVersion-1 {
			reportErr(fmt.Errorf("legacy join protocol=%d, want %d", legacyJoin.ProtocolVersion, protocolVersion-1))
			return
		}
		if err := writeMessage(conn, Message{
			Type:            "join_ok",
			ProtocolVersion: protocolVersion - 1,
			Assigned:        1,
			NumPlayers:      2,
			Slots:           []string{"Host", "Guest"},
			SessionID:       "legacy-reconnect-session",
		}); err != nil {
			reportErr(err)
			return
		}
		initialConnReady <- conn

		conn, err = ln.Accept()
		if err != nil {
			reportErr(err)
			return
		}
		reader = newConnReader(conn)
		rejoin, err := readMessage(conn, reader)
		if err != nil {
			reportErr(err)
			return
		}
		reconnectJoinVersion <- rejoin.ProtocolVersion
		if err := writeMessage(conn, Message{
			Type:            "join_ok",
			ProtocolVersion: protocolVersion - 1,
			Assigned:        1,
			NumPlayers:      2,
			Slots:           []string{"Host", "Guest"},
			SessionID:       "legacy-reconnect-session",
		}); err != nil {
			reportErr(err)
			return
		}
		for {
			msg, err := readMessage(conn, reader)
			if err != nil {
				reportErr(err)
				return
			}
			if msg.Type == "ping" {
				if err := writeMessage(conn, Message{Type: "pong", ProtocolVersion: protocolVersion - 1}); err != nil {
					reportErr(err)
					return
				}
				continue
			}
			if msg.Type == "chat" {
				reconnectChatVersion <- msg.ProtocolVersion
				return
			}
		}
	}()

	inv := InviteKey{
		Addr:             ln.Addr().String(),
		Token:            "legacy-reconnect-token",
		Fingerprint:      fingerprint,
		Transport:        "tcp_tls",
		TransportVersion: 1,
	}
	key, err := EncodeInviteKey(inv)
	if err != nil {
		t.Fatalf("EncodeInviteKey: %v", err)
	}

	cli, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	defer cli.Close()

	if cli.WireProtocolVersion() != protocolVersion-1 {
		t.Fatalf("initial wire protocol = %d, want %d", cli.WireProtocolVersion(), protocolVersion-1)
	}

	var initialConn net.Conn
	select {
	case initialConn = <-initialConnReady:
	case err := <-serverErr:
		t.Fatalf("server error before initial connection ready: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting initial legacy connection")
	}
	closeConnWithLog(initialConn, "force client reconnect")

	waitEventContains(t, cli.Events(), 8*time.Second, "Reconectado ao host")
	if cli.WireProtocolVersion() != protocolVersion-1 {
		t.Fatalf("wire protocol after reconnect = %d, want %d", cli.WireProtocolVersion(), protocolVersion-1)
	}

	if err := cli.SendChat("after reconnect"); err != nil {
		t.Fatalf("SendChat after reconnect: %v", err)
	}

	select {
	case version := <-reconnectJoinVersion:
		if version != protocolVersion-1 {
			t.Fatalf("reconnect join protocol = %d, want %d", version, protocolVersion-1)
		}
	case err := <-serverErr:
		t.Fatalf("server error during reconnect: %v", err)
	case <-time.After(8 * time.Second):
		t.Fatalf("timeout waiting reconnect join")
	}

	select {
	case version := <-reconnectChatVersion:
		if version != protocolVersion-1 {
			t.Fatalf("reconnect chat protocol = %d, want %d", version, protocolVersion-1)
		}
	case err := <-serverErr:
		t.Fatalf("server error during reconnect chat: %v", err)
	case <-time.After(8 * time.Second):
		t.Fatalf("timeout waiting reconnect chat")
	}
}

func TestClientReconnectsAndReceivesCachedState(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	cli, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	defer cli.Close()

	waitUntil(t, 2*time.Second, host.IsFull, "host lobby full")
	if err := host.StartGame(); err != nil {
		t.Fatalf("StartGame: %v", err)
	}

	g, err := truco.NewGame([]string{"Host", "Guest"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	s := g.Snapshot(1)
	s.Logs = append([]string{}, s.Logs...)
	s.Logs = append(s.Logs, "RECONNECT_CACHE_MARK")
	host.SendGameStateToSeat(1, Message{Type: "game_state", State: &s})
	_ = waitState(t, cli.StateUpdates(), 2*time.Second) // estado inicial

	cli.mu.Lock()
	_ = cli.conn.Close()
	cli.mu.Unlock()

	waitEventContains(t, cli.Events(), 8*time.Second, "Reconectado ao host")
	resynced := waitState(t, cli.StateUpdates(), 8*time.Second)
	found := false
	for _, line := range resynced.Logs {
		if strings.Contains(line, "RECONNECT_CACHE_MARK") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cached state marker after reconnect; logs=%v", resynced.Logs)
	}
}

func TestGameStateSnapshotIsTrimmedToWireLimit(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	cli, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	defer cli.Close()

	waitUntil(t, 2*time.Second, host.IsFull, "host lobby full")
	if err := host.StartGame(); err != nil {
		t.Fatalf("StartGame: %v", err)
	}

	g, err := truco.NewGame([]string{"Host", "Guest"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	s := g.Snapshot(1)
	s.Logs = nil
	for i := 0; i < 600; i++ {
		s.Logs = append(s.Logs, fmt.Sprintf("LOG-%03d-%s", i, strings.Repeat("x", 220)))
	}
	host.SendGameStateToSeat(1, Message{Type: "game_state", State: &s})

	recv := waitState(t, cli.StateUpdates(), 2*time.Second)
	if snapshotWireSize(recv) > scannerMaxBuffer {
		t.Fatalf("trimmed snapshot still exceeds wire size limit")
	}
	if len(recv.Logs) > maxOutboundSnapshotLogs {
		t.Fatalf("trimmed logs len = %d, want <= %d", len(recv.Logs), maxOutboundSnapshotLogs)
	}

	if err := cli.SendChat("still connected"); err != nil {
		t.Fatalf("SendChat after large snapshot: %v", err)
	}
	waitEventContains(t, cli.Events(), 2*time.Second, "[chat] Guest: still connected")
}

func TestFourPlayerRoleAssignment(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 4)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	partner, err := JoinSession(key, "Partner", "partner")
	if err != nil {
		t.Fatalf("JoinSession(partner): %v", err)
	}
	defer partner.Close()
	if got := partner.AssignedSeat(); got != 2 {
		t.Fatalf("partner seat=%d, want 2", got)
	}

	opp1, err := JoinSession(key, "Opp1", "opponent")
	if err != nil {
		t.Fatalf("JoinSession(opp1): %v", err)
	}
	defer opp1.Close()
	if got := opp1.AssignedSeat(); got != 1 {
		t.Fatalf("opp1 seat=%d, want 1", got)
	}

	opp2, err := JoinSession(key, "Opp2", "opponent")
	if err != nil {
		t.Fatalf("JoinSession(opp2): %v", err)
	}
	defer opp2.Close()
	if got := opp2.AssignedSeat(); got != 3 {
		t.Fatalf("opp2 seat=%d, want 3", got)
	}
}

func TestHeartbeatTimeoutEvictsIdleClient(t *testing.T) {
	host, key, err := NewHostSessionWithConfig("127.0.0.1:0", "Host", 2, HostConfig{
		HeartbeatInterval: 40 * time.Millisecond,
		HeartbeatTimeout:  120 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewHostSessionWithConfig: %v", err)
	}
	defer host.Close()

	inv, err := DecodeInviteKey(key)
	if err != nil {
		t.Fatalf("DecodeInviteKey: %v", err)
	}
	conn, _, _, err := dialAndJoin(inv, "IdleGuest", "auto", "", 1, supportedProtocolVersions())
	if err != nil {
		t.Fatalf("dialAndJoin: %v", err)
	}
	defer closeConnWithLog(conn, "test idle client")

	waitUntil(t, 2*time.Second, host.IsFull, "host should become full after join")
	waitUntil(t, 3*time.Second, func() bool { return !host.IsFull() }, "idle client should be evicted by heartbeat")
}

func TestMalformedInitialMessageDoesNotBreakHost(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	inv, err := DecodeInviteKey(key)
	if err != nil {
		t.Fatalf("DecodeInviteKey: %v", err)
	}
	conn, err := dialSessionConn(inv, 2*time.Second)
	if err != nil {
		t.Fatalf("dialSessionConn: %v", err)
	}
	if _, err := conn.Write([]byte("{not-json}\n")); err != nil {
		t.Fatalf("write malformed payload: %v", err)
	}
	closeConnWithLog(conn, "test malformed client")

	cli, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession after malformed payload: %v", err)
	}
	defer cli.Close()
	waitUntil(t, 2*time.Second, host.IsFull, "host should still accept valid join")
}

func TestChatBroadcastAcrossClients(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 4)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	c1, err := JoinSession(key, "GuestA", "opponent")
	if err != nil {
		t.Fatalf("JoinSession c1: %v", err)
	}
	defer c1.Close()

	c2, err := JoinSession(key, "GuestB", "partner")
	if err != nil {
		t.Fatalf("JoinSession c2: %v", err)
	}
	defer c2.Close()

	if err := c1.SendChat("hello everyone"); err != nil {
		t.Fatalf("SendChat: %v", err)
	}
	waitEventContains(t, c2.Events(), 2*time.Second, "[chat] GuestA: hello everyone")
}

func TestConcurrentJoinDisconnectRaces(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 4)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("G-%d", i)
			cli, err := JoinSession(key, name, "auto")
			if err != nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
			_ = cli.Close()
		}(i)
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting concurrent join/disconnect")
	}

	// Host should stay responsive after stress.
	cli, err := JoinSession(key, "FinalGuest", "auto")
	if err == nil {
		_ = cli.Close()
	}
}

func TestDemocraticHostTransferByVotes(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 4)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	c1, err := JoinSession(key, "A", "opponent")
	if err != nil {
		t.Fatalf("JoinSession A: %v", err)
	}
	defer c1.Close()
	c2, err := JoinSession(key, "B", "partner")
	if err != nil {
		t.Fatalf("JoinSession B: %v", err)
	}
	defer c2.Close()

	if err := c1.SendHostVote(2); err != nil {
		t.Fatalf("SendHostVote c1: %v", err)
	}
	if err := c2.SendHostVote(2); err != nil {
		t.Fatalf("SendHostVote c2: %v", err)
	}
	waitUntil(t, 2*time.Second, func() bool { return host.CurrentHostSeat() == 2 }, "host seat should move to slot 3 by votes")
}

func TestHostTransferOnDisconnect(t *testing.T) {
	host, key, err := NewHostSessionWithConfig("127.0.0.1:0", "Host", 2, HostConfig{
		HeartbeatInterval: 50 * time.Millisecond,
		HeartbeatTimeout:  200 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewHostSessionWithConfig: %v", err)
	}
	defer host.Close()

	guest, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	defer guest.Close()

	if err := host.CastHostVote(0, 1); err != nil {
		t.Fatalf("CastHostVote local: %v", err)
	}
	if err := guest.SendHostVote(1); err != nil {
		t.Fatalf("SendHostVote guest: %v", err)
	}
	waitUntil(t, 2*time.Second, func() bool { return host.CurrentHostSeat() == 1 }, "host seat should become guest")

	_ = guest.Close()
	waitUntil(t, 4*time.Second, func() bool { return host.CurrentHostSeat() == 0 }, "host seat should transfer back after disconnect")
}

func TestReplacementInviteDuringStartedGame(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	guest, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession Guest: %v", err)
	}
	defer guest.Close()
	waitUntil(t, 2*time.Second, host.IsFull, "host lobby full")

	if err := host.StartGame(); err != nil {
		t.Fatalf("StartGame: %v", err)
	}
	_ = guest.Close()
	waitUntil(t, 2*time.Second, func() bool { return !host.IsSeatConnected(1) }, "seat 2 should disconnect")

	replKey, err := host.RequestReplacementInvite(0, 1)
	if err != nil {
		t.Fatalf("RequestReplacementInvite: %v", err)
	}
	sub, err := JoinSession(replKey, "Sub", "auto")
	if err != nil {
		t.Fatalf("JoinSession replacement: %v", err)
	}
	defer sub.Close()

	if got := sub.AssignedSeat(); got != 1 {
		t.Fatalf("replacement seat = %d, want 1", got)
	}
	slots := host.Slots()
	if slots[1] != "Sub" {
		t.Fatalf("slot 2 name = %q, want %q", slots[1], "Sub")
	}
}

func TestTransferredHostCanIssueReplacementInvite(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 4)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	a, err := JoinSession(key, "A", "opponent")
	if err != nil {
		t.Fatalf("JoinSession A: %v", err)
	}
	defer a.Close()
	b, err := JoinSession(key, "B", "partner")
	if err != nil {
		t.Fatalf("JoinSession B: %v", err)
	}
	defer b.Close()
	c, err := JoinSession(key, "C", "opponent")
	if err != nil {
		t.Fatalf("JoinSession C: %v", err)
	}
	defer c.Close()

	waitUntil(t, 2*time.Second, host.IsFull, "lobby should be full")
	if err := host.StartGame(); err != nil {
		t.Fatalf("StartGame: %v", err)
	}

	// Transfer table host para slot 2 (player B).
	if err := host.CastHostVote(0, 2); err != nil {
		t.Fatalf("CastHostVote local: %v", err)
	}
	if err := a.SendHostVote(2); err != nil {
		t.Fatalf("SendHostVote A: %v", err)
	}
	if err := c.SendHostVote(2); err != nil {
		t.Fatalf("SendHostVote C: %v", err)
	}
	waitUntil(t, 2*time.Second, func() bool { return host.CurrentHostSeat() == 2 }, "host seat should be slot 3")

	// Derruba slot 4 para gerar condição de reposição.
	_ = c.Close()
	waitUntil(t, 2*time.Second, func() bool { return !host.IsSeatConnected(3) }, "slot 4 should disconnect")

	// Novo host remoto (B) pede convite de reposição para slot 4.
	if err := b.RequestReplacementInvite(3); err != nil {
		t.Fatalf("RequestReplacementInvite by transferred host: %v", err)
	}
	waitEventContains(t, b.Events(), 2*time.Second, "Convite de reposição (slot 4):")
}

func TestRotateFailoverSnapshot(t *testing.T) {
	s := truco.Snapshot{
		NumPlayers: 4,
		Players: []truco.Player{
			{ID: 0, Name: "P0", Team: 0, Hand: []truco.Card{{Suit: truco.Clubs, Rank: truco.RA}}},
			{ID: 1, Name: "P1", Team: 1, Hand: []truco.Card{{Suit: truco.Hearts, Rank: truco.R2}}},
			{ID: 2, Name: "P2", Team: 0, Hand: []truco.Card{{Suit: truco.Spades, Rank: truco.R3}}},
			{ID: 3, Name: "P3", Team: 1, Hand: []truco.Card{{Suit: truco.Diamonds, Rank: truco.RK}}},
		},
		CurrentHand: truco.HandState{
			Dealer:         3,
			Turn:           1,
			RoundStart:     2,
			RaiseRequester: 0,
			RoundCards: []truco.PlayedCard{
				{PlayerID: 0, Card: truco.Card{Suit: truco.Clubs, Rank: truco.RA}},
				{PlayerID: 2, Card: truco.Card{Suit: truco.Spades, Rank: truco.R3}},
			},
		},
		TurnPlayer:       1,
		CurrentPlayerIdx: 3,
	}

	rot, err := RotateFailoverSnapshot(s, 2)
	if err != nil {
		t.Fatalf("RotateFailoverSnapshot: %v", err)
	}
	if rot.Players[0].Name != "P2" || rot.Players[1].Name != "P3" || rot.Players[2].Name != "P0" || rot.Players[3].Name != "P1" {
		t.Fatalf("unexpected rotated players order: %+v", []string{rot.Players[0].Name, rot.Players[1].Name, rot.Players[2].Name, rot.Players[3].Name})
	}
	if rot.CurrentHand.Turn != 3 || rot.CurrentHand.Dealer != 1 || rot.CurrentHand.RoundStart != 0 {
		t.Fatalf("unexpected rotated hand ids: turn=%d dealer=%d roundStart=%d", rot.CurrentHand.Turn, rot.CurrentHand.Dealer, rot.CurrentHand.RoundStart)
	}
	if len(rot.CurrentHand.RoundCards) != 2 || rot.CurrentHand.RoundCards[0].PlayerID != 2 || rot.CurrentHand.RoundCards[1].PlayerID != 0 {
		t.Fatalf("unexpected rotated round cards: %+v", rot.CurrentHand.RoundCards)
	}
}

func TestRecoveredHostSessionReconnectBySessionID(t *testing.T) {
	host, key, err := NewHostSession("127.0.0.1:0", "Host", 2)
	if err != nil {
		t.Fatalf("NewHostSession: %v", err)
	}
	defer host.Close()

	guest, err := JoinSession(key, "Guest", "auto")
	if err != nil {
		t.Fatalf("JoinSession Guest: %v", err)
	}
	defer guest.Close()
	waitUntil(t, 2*time.Second, host.IsFull, "lobby full")
	if err := host.StartGame(); err != nil {
		t.Fatalf("StartGame: %v", err)
	}

	g, err := truco.NewGame([]string{"Host", "Guest"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	guestView := g.Snapshot(1)
	full := g.Snapshot(0)
	host.SendGameStateToSeat(1, Message{Type: "game_state", State: &guestView, FullState: &full})
	_ = waitState(t, guest.StateUpdates(), 2*time.Second)

	fs := guest.FailoverState()
	if !fs.Ready {
		t.Fatalf("expected ready failover state")
	}
	rotSnap, err := RotateFailoverSnapshot(*fs.FullState, 1)
	if err != nil {
		t.Fatalf("RotateFailoverSnapshot: %v", err)
	}
	_ = rotSnap
	rotSlots := RotateSeatSlice(fs.Slots, 1)
	rotPeers := RotateSeatMapString(fs.PeerHosts, 1, fs.NumPlayers)
	rotIDs := RotateSeatMapString(fs.SeatSessionIDs, 1, fs.NumPlayers)

	recovered, recoveredKey, err := NewRecoveredHostSession(
		"127.0.0.1:0",
		rotSlots[0],
		fs.NumPlayers,
		RecoveredHostState{
			Token:          fs.Invite.Token,
			TLSSeed:        fs.TLSSeed,
			Slots:          rotSlots,
			SeatSessionIDs: rotIDs,
			PeerHosts:      rotPeers,
			TableHostSeat:  0,
		},
		HostConfig{},
	)
	if err != nil {
		t.Fatalf("NewRecoveredHostSession: %v", err)
	}
	defer recovered.Close()

	inv, err := DecodeInviteKey(recoveredKey)
	if err != nil {
		t.Fatalf("DecodeInviteKey(recovered): %v", err)
	}
	sidOldHost := rotIDs[1]
	rejoin, err := RejoinSession(inv, rotSlots[1], "auto", sidOldHost, 8)
	if err != nil {
		t.Fatalf("RejoinSession to recovered host: %v", err)
	}
	defer rejoin.Close()
	if got := rejoin.AssignedSeat(); got != 1 {
		t.Fatalf("rejoin assigned seat=%d, want 1", got)
	}
}
