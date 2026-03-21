package appcore

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"truco-tui/internal/netp2p"
)

func waitForRuntimeCondition(t *testing.T, timeout time.Duration, cond func() bool, msg string) {
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

func TestRuntimeOfflineSeededSnapshotAndActions(t *testing.T) {
	rt := NewRuntime()
	defer func() { _ = rt.Close() }()

	payload, err := json.Marshal(NewOfflineGamePayload{
		PlayerNames: []string{"Ana", "CPU-2"},
		CPUFlags:    []bool{false, true},
		SeedLo:      7,
		SeedHi:      9,
	})
	if err != nil {
		t.Fatalf("Marshal payload: %v", err)
	}
	if err := rt.DispatchIntent(AppIntent{Kind: "new_offline_game", Payload: payload}); err != nil {
		t.Fatalf("Dispatch new_offline_game: %v", err)
	}

	state := rt.SnapshotBundle()
	if state.Mode != "offline_match" {
		t.Fatalf("mode = %q, want offline_match", state.Mode)
	}
	if state.Match == nil {
		t.Fatal("match snapshot is nil")
	}
	if len(state.Match.Players) != 2 {
		t.Fatalf("players = %d, want 2", len(state.Match.Players))
	}

	foundMatchUpdate := false
	foundSessionReady := false
	for {
		ev, ok := rt.PollEvent()
		if !ok {
			break
		}
		if ev.Kind == "match_updated" {
			foundMatchUpdate = true
		}
		if ev.Kind == "session_ready" {
			foundSessionReady = true
		}
	}
	if !foundMatchUpdate {
		t.Fatal("expected at least one match_updated event")
	}
	if !foundSessionReady {
		t.Fatal("expected session_ready event")
	}
}

func TestRuntimeSnapshotJSONIncludesVersions(t *testing.T) {
	rt := NewRuntime()
	defer func() { _ = rt.Close() }()

	snapshotJSON, err := rt.SnapshotJSON()
	if err != nil {
		t.Fatalf("SnapshotJSON: %v", err)
	}

	var state SnapshotBundle
	if err := json.Unmarshal([]byte(snapshotJSON), &state); err != nil {
		t.Fatalf("Unmarshal snapshot JSON: %v", err)
	}
	if state.Versions.CoreAPIVersion != CoreAPIVersion {
		t.Fatalf("core_api_version = %d, want %d", state.Versions.CoreAPIVersion, CoreAPIVersion)
	}
	if state.Versions.SnapshotSchema != SnapshotSchemaMajor {
		t.Fatalf("snapshot_schema_version = %d, want %d", state.Versions.SnapshotSchema, SnapshotSchemaMajor)
	}
}

func TestRuntimeLocaleChange(t *testing.T) {
	rt := NewRuntime()
	defer func() { _ = rt.Close() }()

	payload, err := json.Marshal(SetLocalePayload{Locale: "en-US"})
	if err != nil {
		t.Fatalf("Marshal payload: %v", err)
	}
	if err := rt.DispatchIntent(AppIntent{Kind: "set_locale", Payload: payload}); err != nil {
		t.Fatalf("Dispatch set_locale: %v", err)
	}

	state := rt.SnapshotBundle()
	if state.Locale != "en-US" {
		t.Fatalf("locale = %q, want en-US", state.Locale)
	}
	ev, ok := rt.PollEvent()
	if !ok {
		t.Fatal("expected locale_changed event")
	}
	if ev.Kind != "locale_changed" {
		t.Fatalf("event kind = %q, want locale_changed", ev.Kind)
	}
}

func TestRuntimeNewHandAndReset(t *testing.T) {
	rt := NewRuntime()
	defer func() { _ = rt.Close() }()

	payload, err := json.Marshal(NewOfflineGamePayload{
		PlayerNames: []string{"Ana", "CPU-2"},
		CPUFlags:    []bool{false, true},
	})
	if err != nil {
		t.Fatalf("Marshal payload: %v", err)
	}
	if err := rt.DispatchIntent(AppIntent{Kind: "new_offline_game", Payload: payload}); err != nil {
		t.Fatalf("Dispatch new_offline_game: %v", err)
	}

	before := rt.SnapshotBundle()
	if before.Match == nil {
		t.Fatal("match snapshot is nil")
	}
	oldDealer := before.Match.CurrentHand.Dealer

	if err := rt.DispatchIntent(AppIntent{Kind: "new_hand"}); err != nil {
		t.Fatalf("Dispatch new_hand: %v", err)
	}
	after := rt.SnapshotBundle()
	if after.Mode != "offline_match" {
		t.Fatalf("mode = %q, want offline_match", after.Mode)
	}
	if after.Match == nil {
		t.Fatal("match snapshot after new_hand is nil")
	}
	if after.Match.CurrentHand.Dealer == oldDealer {
		t.Fatalf("dealer did not rotate: %d", after.Match.CurrentHand.Dealer)
	}

	if err := rt.DispatchIntent(AppIntent{Kind: "reset"}); err != nil {
		t.Fatalf("Dispatch reset: %v", err)
	}
	cleared := rt.SnapshotBundle()
	if cleared.Mode != "idle" {
		t.Fatalf("mode after reset = %q, want idle", cleared.Mode)
	}
	if cleared.Match != nil {
		t.Fatal("match should be nil after reset")
	}
	if cleared.Lobby != nil {
		t.Fatal("lobby should be nil after reset")
	}
}

func TestRuntimeHostLobbyNetworkSnapshot(t *testing.T) {
	rt := NewRuntime()
	defer func() { _ = rt.Close() }()

	payload, err := json.Marshal(CreateHostPayload{
		HostName:   "Host",
		NumPlayers: 2,
	})
	if err != nil {
		t.Fatalf("Marshal payload: %v", err)
	}
	if err := rt.DispatchIntent(AppIntent{Kind: "create_host_session", Payload: payload}); err != nil {
		t.Fatalf("Dispatch create_host_session: %v", err)
	}

	state := rt.SnapshotBundle()
	if state.Mode != "host_lobby" {
		t.Fatalf("mode = %q, want host_lobby", state.Mode)
	}
	if state.Connection.Network == nil {
		t.Fatal("connection.network is nil")
	}
	network := state.Connection.Network
	if network.Transport != "tcp_tls" {
		t.Fatalf("transport = %q, want tcp_tls", network.Transport)
	}
	if !reflect.DeepEqual(network.SupportedProtocolVersions, netp2p.SupportedProtocolVersions()) {
		t.Fatalf("supported versions = %v, want %v", network.SupportedProtocolVersions, netp2p.SupportedProtocolVersions())
	}
	if got := network.SeatProtocolVersions[0]; got != netp2p.ProtocolVersion {
		t.Fatalf("seat 0 protocol = %d, want %d", got, netp2p.ProtocolVersion)
	}
	if network.MixedProtocolSession {
		t.Fatal("mixed_protocol_session = true, want false")
	}
}

func TestRuntimeClientLobbyNetworkSnapshot(t *testing.T) {
	hostRT := NewRuntime()
	defer func() { _ = hostRT.Close() }()
	clientRT := NewRuntime()
	defer func() { _ = clientRT.Close() }()

	hostPayload, err := json.Marshal(CreateHostPayload{
		HostName:   "Host",
		NumPlayers: 2,
	})
	if err != nil {
		t.Fatalf("Marshal host payload: %v", err)
	}
	if err := hostRT.DispatchIntent(AppIntent{Kind: "create_host_session", Payload: hostPayload}); err != nil {
		t.Fatalf("Dispatch create_host_session: %v", err)
	}

	hostSnap := hostRT.SnapshotBundle()
	if hostSnap.Lobby == nil || hostSnap.Lobby.InviteKey == "" {
		t.Fatal("host invite key missing")
	}

	joinPayload, err := json.Marshal(JoinSessionPayload{
		Key:         hostSnap.Lobby.InviteKey,
		PlayerName:  "Guest",
		DesiredRole: "auto",
	})
	if err != nil {
		t.Fatalf("Marshal join payload: %v", err)
	}
	if err := clientRT.DispatchIntent(AppIntent{Kind: "join_session", Payload: joinPayload}); err != nil {
		t.Fatalf("Dispatch join_session: %v", err)
	}

	waitForRuntimeCondition(t, 2*time.Second, func() bool {
		snap := hostRT.SnapshotBundle()
		return snap.Connection.Network != nil &&
			snap.Connection.Network.SeatProtocolVersions[1] == netp2p.ProtocolVersion &&
			snap.Lobby != nil &&
			snap.Lobby.ConnectedSeats[1]
	}, "host lobby to include remote network state")

	clientState := clientRT.SnapshotBundle()
	if clientState.Mode != "client_lobby" {
		t.Fatalf("mode = %q, want client_lobby", clientState.Mode)
	}
	if clientState.Connection.Network == nil {
		t.Fatal("client connection.network is nil")
	}
	clientNetwork := clientState.Connection.Network
	if clientNetwork.Transport != "tcp_tls" {
		t.Fatalf("transport = %q, want tcp_tls", clientNetwork.Transport)
	}
	if clientNetwork.NegotiatedProtocolVersion != netp2p.ProtocolVersion {
		t.Fatalf("negotiated protocol = %d, want %d", clientNetwork.NegotiatedProtocolVersion, netp2p.ProtocolVersion)
	}
	if !reflect.DeepEqual(clientNetwork.SupportedProtocolVersions, netp2p.SupportedProtocolVersions()) {
		t.Fatalf("supported versions = %v, want %v", clientNetwork.SupportedProtocolVersions, netp2p.SupportedProtocolVersions())
	}
	if len(clientNetwork.SeatProtocolVersions) != 0 {
		t.Fatalf("client seat_protocol_versions = %v, want empty", clientNetwork.SeatProtocolVersions)
	}
}
