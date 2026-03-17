package netp2p

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"truco-tui/internal/truco"
)

// ClientSession participa de lobby/chat e da partida online.
type ClientSession struct {
	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	conn         net.Conn
	reader       *bufio.Reader
	invite       InviteKey
	sessionID    string
	desiredRole  string
	name         string
	assigned     int
	numPlayers   int
	slots        []string
	connected    map[int]bool
	events       chan string
	states       chan truco.Snapshot
	started      bool
	reconnecting bool
	closed       bool

	failoverHostSeat             int
	failoverPort                 int
	failoverPeers                map[int]string
	failoverSeatIDs              map[int]string
	failoverState                *truco.Snapshot
	failoverTLSSeed              string
	failoverEpoch                int
	failoverAuthorityFingerprint string
	failoverRouteHint            string
	failoverRelayHostAdminToken  string
}

type joinProtocolError struct {
	msg string
}

func (e joinProtocolError) Error() string {
	return e.msg
}

var (
	clientHeartbeatInterval  = 5 * time.Second
	clientConnectAttempts    = 3
	clientReconnectAttempts  = 8
	clientReconnectBaseDelay = 500 * time.Millisecond
)

const (
	ClientEventHostLostFailover = "__HOST_LOST_FAILOVER__"
)

type ClientFailoverState struct {
	Ready                bool
	HostSeat             int
	HandoffPort          int
	PeerHosts            map[int]string
	SeatSessionIDs       map[int]string
	FullState            *truco.Snapshot
	Slots                []string
	AssignedSeat         int
	NumPlayers           int
	Invite               InviteKey
	Name                 string
	DesiredRole          string
	SessionID            string
	TLSSeed              string
	Epoch                int
	AuthorityFingerprint string
	RouteHint            string
	RelayHostAdminToken  string
}

func JoinSession(key, playerName, desiredRole string) (*ClientSession, error) {
	inv, err := DecodeInviteKey(key)
	if err != nil {
		return nil, err
	}
	name, err := normalizeName(playerName)
	if err != nil {
		return nil, err
	}
	desiredRole = normalizeDesiredRole(desiredRole)

	conn, reader, first, err := dialAndJoin(inv, name, desiredRole, "", clientConnectAttempts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &ClientSession{
		ctx:              ctx,
		cancel:           cancel,
		conn:             conn,
		reader:           reader,
		invite:           inv,
		sessionID:        first.SessionID,
		desiredRole:      desiredRole,
		name:             name,
		assigned:         first.Assigned,
		numPlayers:       first.NumPlayers,
		slots:            append([]string{}, first.Slots...),
		connected:        cloneSeatBoolMap(first.ConnectedSeats),
		failoverHostSeat: first.HostSeat,
		events:           make(chan string, 128),
		states:           make(chan truco.Snapshot, 128),
	}
	go c.readLoop()
	go c.heartbeatLoop()
	return c, nil
}

func RejoinSession(inv InviteKey, playerName, desiredRole, sessionID string, attempts int) (*ClientSession, error) {
	name, err := normalizeName(playerName)
	if err != nil {
		return nil, err
	}
	desiredRole = normalizeDesiredRole(desiredRole)
	conn, reader, first, err := dialAndJoin(inv, name, desiredRole, sessionID, attempts)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := &ClientSession{
		ctx:              ctx,
		cancel:           cancel,
		conn:             conn,
		reader:           reader,
		invite:           inv,
		sessionID:        first.SessionID,
		desiredRole:      desiredRole,
		name:             name,
		assigned:         first.Assigned,
		numPlayers:       first.NumPlayers,
		slots:            append([]string{}, first.Slots...),
		connected:        cloneSeatBoolMap(first.ConnectedSeats),
		failoverHostSeat: first.HostSeat,
		events:           make(chan string, 128),
		states:           make(chan truco.Snapshot, 128),
	}
	go c.readLoop()
	go c.heartbeatLoop()
	return c, nil
}

func dialAndJoin(inv InviteKey, playerName, desiredRole, sessionID string, attempts int) (net.Conn, *bufio.Reader, Message, error) {
	if attempts < 1 {
		attempts = 1
	}
	addrs := inviteDialAddrs(inv.Addr)
	if len(addrs) == 0 {
		addrs = []string{inv.Addr}
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		for _, addr := range addrs {
			tryInv := inv
			tryInv.Addr = addr
			conn, reader, first, err := attemptDialJoin(tryInv, playerName, desiredRole, sessionID)
			if err == nil {
				return conn, reader, first, nil
			}
			var protocolErr joinProtocolError
			if errors.As(err, &protocolErr) {
				return nil, nil, Message{}, err
			}
			lastErr = fmt.Errorf("%s: %w", addr, err)
		}
		if attempt < attempts {
			time.Sleep(250 * time.Millisecond)
		}
	}

	if lastErr == nil {
		lastErr = errors.New("falha de conexão")
	}
	if attempts > 1 {
		return nil, nil, Message{}, fmt.Errorf("falha ao conectar após tentativas: %w", lastErr)
	}
	return nil, nil, Message{}, lastErr
}

func inviteDialAddrs(addr string) []string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return []string{addr}
	}
	host = strings.TrimSpace(strings.Trim(host, "[]"))
	out := make([]string, 0, 4)
	seen := map[string]bool{}
	add := func(h string) {
		h = strings.TrimSpace(h)
		if h == "" {
			return
		}
		a := net.JoinHostPort(h, port)
		if seen[a] {
			return
		}
		seen[a] = true
		out = append(out, a)
	}

	add(host)
	switch strings.ToLower(host) {
	case "", "0.0.0.0", "::":
		add("127.0.0.1")
		add("localhost")
		add("::1")
	case "localhost":
		add("127.0.0.1")
		add("::1")
	case "127.0.0.1":
		add("localhost")
		add("::1")
	case "::1":
		add("127.0.0.1")
		add("localhost")
	}
	return out
}

func cloneSeatBoolMap(in map[int]bool) map[int]bool {
	if in == nil {
		return nil
	}
	out := make(map[int]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func attemptDialJoin(inv InviteKey, playerName, desiredRole, sessionID string) (net.Conn, *bufio.Reader, Message, error) {
	conn, err := dialSessionConnWithRelay(inv, 2*time.Second, playerName, desiredRole, sessionID)
	if err != nil {
		return nil, nil, Message{}, err
	}
	advertiseHost := ""
	if host, _, splitErr := net.SplitHostPort(conn.LocalAddr().String()); splitErr == nil {
		advertiseHost = host
	}
	reader := newConnReader(conn)
	req := Message{
		Type:            "join",
		ProtocolVersion: protocolVersion,
		Token:           inv.Token,
		Name:            playerName,
		DesiredRole:     desiredRole,
		SessionID:       sessionID,
		ReplaceToken:    inv.ReplaceToken,
		AdvertiseHost:   advertiseHost,
	}
	if err := writeMessage(conn, req); err != nil {
		closeConnWithLog(conn, "join send")
		return nil, nil, Message{}, err
	}

	first, err := readMessage(conn, reader)
	if err != nil {
		closeConnWithLog(conn, "join read")
		return nil, nil, Message{}, err
	}
	if first.Type == "error" {
		closeConnWithLog(conn, "join protocol error")
		return nil, nil, Message{}, joinProtocolError{msg: first.Error}
	}
	if first.Type != "join_ok" {
		closeConnWithLog(conn, "join unexpected response")
		return nil, nil, Message{}, joinProtocolError{msg: fmt.Sprintf("resposta inesperada: %s", first.Type)}
	}
	if first.ProtocolVersion != protocolVersion {
		closeConnWithLog(conn, "join protocol version mismatch")
		return nil, nil, Message{}, joinProtocolError{msg: fmt.Sprintf("versão de protocolo incompatível (host=%d, esperado=%d)", first.ProtocolVersion, protocolVersion)}
	}
	return conn, reader, first, nil
}

func (c *ClientSession) Events() <-chan string               { return c.events }
func (c *ClientSession) StateUpdates() <-chan truco.Snapshot { return c.states }
func (c *ClientSession) AssignedSeat() int                   { return c.assigned }
func (c *ClientSession) Name() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.name
}

func (c *ClientSession) Slots() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]string, len(c.slots))
	copy(out, c.slots)
	return out
}

func (c *ClientSession) ConnectedSeats() map[int]bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return cloneSeatBoolMap(c.connected)
}

func (c *ClientSession) CurrentHostSeat() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failoverHostSeat
}

func (c *ClientSession) DesiredRole() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.desiredRole
}

func (c *ClientSession) GameStarted() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.started
}

func (c *ClientSession) FailoverState() ClientFailoverState {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := ClientFailoverState{
		HostSeat:             c.failoverHostSeat,
		HandoffPort:          c.failoverPort,
		PeerHosts:            make(map[int]string, len(c.failoverPeers)),
		SeatSessionIDs:       make(map[int]string, len(c.failoverSeatIDs)),
		Slots:                append([]string{}, c.slots...),
		AssignedSeat:         c.assigned,
		NumPlayers:           c.numPlayers,
		Invite:               c.invite,
		Name:                 c.name,
		DesiredRole:          c.desiredRole,
		SessionID:            c.sessionID,
		TLSSeed:              c.failoverTLSSeed,
		Epoch:                c.failoverEpoch,
		AuthorityFingerprint: c.failoverAuthorityFingerprint,
		RouteHint:            c.failoverRouteHint,
		RelayHostAdminToken:  c.failoverRelayHostAdminToken,
	}
	for seat, host := range c.failoverPeers {
		out.PeerHosts[seat] = host
	}
	for seat, sid := range c.failoverSeatIDs {
		out.SeatSessionIDs[seat] = sid
	}
	if c.failoverState != nil {
		s := cloneSnapshot(*c.failoverState)
		out.FullState = &s
	}
	out.Ready = len(out.PeerHosts) > 0 && out.FullState != nil && len(out.Slots) == out.NumPlayers && strings.TrimSpace(out.TLSSeed) != ""
	return out
}

func (c *ClientSession) SendChat(text string) error {
	normalized, err := normalizeChatText(text)
	if err != nil {
		return err
	}
	return c.send(Message{Type: "chat", Text: normalized})
}

func (c *ClientSession) SendGameAction(action string, cardIndex int) error {
	if err := validateGameAction(action, cardIndex); err != nil {
		return err
	}
	return c.send(Message{Type: "game_action", Action: action, CardIndex: cardIndex})
}

func (c *ClientSession) SendHostVote(candidateSeat int) error {
	return c.send(Message{Type: "host_vote", HostCandidate: candidateSeat})
}

func (c *ClientSession) RequestReplacementInvite(targetSeat int) error {
	return c.send(Message{Type: "invite_request", TargetSeat: targetSeat})
}

func (c *ClientSession) send(msg Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("sessão encerrada")
	}
	if c.reconnecting {
		return errors.New("reconectando ao host")
	}
	if c.conn == nil {
		return errors.New("sem conexão ativa")
	}
	if err := writeMessage(c.conn, msg); err != nil {
		closeConnWithLog(c.conn, "client send")
		return err
	}
	return nil
}

func (c *ClientSession) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.cancel()
	var err error
	if c.conn != nil {
		err = c.conn.Close()
		if err != nil {
			logNetf("close conn (client close): %v", err)
		}
	}
	c.mu.Unlock()
	return err
}

func (c *ClientSession) safeEvent(text string) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	ch := c.events
	c.mu.Unlock()
	select {
	case ch <- text:
	default:
	}
}

func (c *ClientSession) safeState(s truco.Snapshot) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	ch := c.states
	c.mu.Unlock()
	select {
	case ch <- s:
	default:
		// Mantém sempre o estado mais recente caso o consumidor esteja atrasado.
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- s:
		default:
		}
	}
}

func (c *ClientSession) reconnectBackoff(attempt int) time.Duration {
	d := time.Duration(attempt) * clientReconnectBaseDelay
	if d > 4*time.Second {
		return 4 * time.Second
	}
	return d
}

func (c *ClientSession) tryReconnect() bool {
	if c.ctx.Err() != nil {
		return false
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return false
	}
	if c.reconnecting {
		c.mu.Unlock()
		return false
	}
	c.reconnecting = true
	inv := c.invite
	name := c.name
	role := c.desiredRole
	sessionID := c.sessionID
	oldConn := c.conn
	c.conn = nil
	c.reader = nil
	c.mu.Unlock()

	closeConnWithLog(oldConn, "client reconnect begin")
	c.safeEvent("Conexão perdida. Tentando reconectar...")

	for attempt := 1; attempt <= clientReconnectAttempts; attempt++ {
		if c.ctx.Err() != nil {
			break
		}
		conn, reader, first, err := dialAndJoin(inv, name, role, sessionID, 1)
		if err == nil {
			c.mu.Lock()
			if c.closed {
				c.reconnecting = false
				c.mu.Unlock()
				closeConnWithLog(conn, "client reconnect canceled")
				return false
			}
			c.conn = conn
			c.reader = reader
			c.assigned = first.Assigned
			c.numPlayers = first.NumPlayers
			if strings.TrimSpace(first.SessionID) != "" {
				c.sessionID = first.SessionID
			}
			if len(first.Slots) > 0 {
				c.slots = append([]string{}, first.Slots...)
			}
			c.reconnecting = false
			c.mu.Unlock()
			c.safeEvent("Reconectado ao host.")
			return true
		}
		if attempt < clientReconnectAttempts {
			time.Sleep(c.reconnectBackoff(attempt))
		}
	}

	c.mu.Lock()
	c.reconnecting = false
	c.mu.Unlock()
	return false
}

func (c *ClientSession) readLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		c.mu.Lock()
		conn := c.conn
		reader := c.reader
		reconnecting := c.reconnecting
		c.mu.Unlock()
		if reconnecting {
			time.Sleep(60 * time.Millisecond)
			continue
		}
		if reader == nil {
			if c.tryReconnect() {
				continue
			}
			c.safeEvent("Conexão encerrada. Não foi possível reconectar.")
			c.safeEvent(ClientEventHostLostFailover)
			_ = c.Close()
			return
		}

		msg, err := readMessage(conn, reader)
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			if c.tryReconnect() {
				continue
			}
			c.safeEvent("Conexão encerrada. Não foi possível reconectar.")
			c.safeEvent(ClientEventHostLostFailover)
			_ = c.Close()
			return
		}

		switch msg.Type {
		case "chat":
			c.safeEvent(fmt.Sprintf("[chat] %s: %s", msg.Name, msg.Text))
		case "lobby_update":
			c.mu.Lock()
			c.slots = append([]string{}, msg.Slots...)
			c.connected = cloneSeatBoolMap(msg.ConnectedSeats)
			if msg.HostSeat >= 0 {
				c.failoverHostSeat = msg.HostSeat
			}
			c.mu.Unlock()
			c.safeEvent("Lobby atualizado.")
		case "game_start":
			c.mu.Lock()
			c.started = true
			if len(msg.Slots) > 0 {
				c.slots = append([]string{}, msg.Slots...)
			}
			c.connected = cloneSeatBoolMap(msg.ConnectedSeats)
			if msg.HostSeat >= 0 {
				c.failoverHostSeat = msg.HostSeat
			}
			c.mu.Unlock()
			c.safeEvent("Partida iniciada pelo host.")
		case "game_state":
			want := normalizeFingerprint(c.invite.Fingerprint)
			fp := strings.TrimSpace(msg.AuthorityFingerprint)
			if fp == "" {
				c.safeEvent("[erro] estado ignorado: fingerprint da autoridade ausente")
				continue
			}
			if want != "" && normalizeFingerprint(fp) != want {
				c.safeEvent("[erro] estado ignorado: autoridade inválida")
				continue
			}
			c.mu.Lock()
			if msg.HandoffPort > 0 {
				c.failoverPort = msg.HandoffPort
			}
			if msg.HostSeat >= 0 {
				c.failoverHostSeat = msg.HostSeat
			}
			if len(msg.PeerHosts) > 0 {
				c.failoverPeers = make(map[int]string, len(msg.PeerHosts))
				for seat, host := range msg.PeerHosts {
					c.failoverPeers[seat] = host
				}
			}
			if len(msg.SeatSessionIDs) > 0 {
				c.failoverSeatIDs = make(map[int]string, len(msg.SeatSessionIDs))
				for seat, sid := range msg.SeatSessionIDs {
					c.failoverSeatIDs[seat] = sid
				}
			}
			if msg.FullState != nil {
				s := cloneSnapshot(*msg.FullState)
				c.failoverState = &s
			}
			if strings.TrimSpace(msg.TLSSeed) != "" {
				c.failoverTLSSeed = msg.TLSSeed
			}
			if msg.Epoch > 0 {
				c.failoverEpoch = msg.Epoch
			}
			if strings.TrimSpace(msg.AuthorityFingerprint) != "" {
				c.failoverAuthorityFingerprint = msg.AuthorityFingerprint
			}
			if strings.TrimSpace(msg.RouteHint) != "" {
				c.failoverRouteHint = msg.RouteHint
			}
			if strings.TrimSpace(msg.RelayHostAdminToken) != "" {
				c.failoverRelayHostAdminToken = msg.RelayHostAdminToken
			}
			c.mu.Unlock()
			if msg.State != nil {
				c.safeState(cloneSnapshot(*msg.State))
			}
		case "system":
			c.safeEvent("[system] " + msg.Text)
		case "error":
			c.safeEvent("[erro] " + msg.Error)
		case "shutdown":
			c.safeEvent("[system] " + msg.Text)
			_ = c.Close()
			return
		case "ping":
			if err := c.send(Message{Type: "pong", HeartbeatUnix: time.Now().Unix()}); err != nil {
				if c.tryReconnect() {
					continue
				}
				c.safeEvent("Conexão encerrada. Não foi possível reconectar.")
				c.safeEvent(ClientEventHostLostFailover)
				_ = c.Close()
				return
			}
		case "pong":
			// keepalive ack
		}
	}
}

func (c *ClientSession) heartbeatLoop() {
	tk := time.NewTicker(clientHeartbeatInterval)
	defer tk.Stop()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-tk.C:
			c.mu.Lock()
			reconnecting := c.reconnecting
			c.mu.Unlock()
			if reconnecting {
				continue
			}
			if err := c.send(Message{Type: "ping", HeartbeatUnix: time.Now().Unix()}); err != nil {
				if c.ctx.Err() != nil {
					return
				}
				if c.tryReconnect() {
					continue
				}
				c.safeEvent("Conexão encerrada. Não foi possível reconectar.")
				c.safeEvent(ClientEventHostLostFailover)
				_ = c.Close()
				return
			}
		}
	}
}
