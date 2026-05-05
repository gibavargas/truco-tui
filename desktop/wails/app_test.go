//go:build wails

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"truco-tui/internal/appcore"
)

func TestStartOfflineGameEmitsRuntimeUpdate(t *testing.T) {
	app := NewApp()
	updates := make(chan RuntimeUpdate, 4)
	app.emit = func(eventName string, payload any) {
		if eventName != runtimeUpdateEvent {
			return
		}
		update, ok := payload.(RuntimeUpdate)
		if !ok {
			t.Fatalf("payload type = %T, want RuntimeUpdate", payload)
		}
		updates <- update
	}

	if err := app.StartOfflineGame("Mesa", 4); err != nil {
		t.Fatalf("StartOfflineGame: %v", err)
	}

	select {
	case update := <-updates:
		if update.Bundle.Mode != appcore.ModeOfflineMatch {
			t.Fatalf("mode = %q, want %q", update.Bundle.Mode, appcore.ModeOfflineMatch)
		}
		if len(update.Events) == 0 {
			t.Fatal("expected runtime update events after starting a match")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for runtime update")
	}
}

func TestRuntimePumpEmitsAsyncUpdates(t *testing.T) {
	app := NewApp()
	app.pumpInterval = 10 * time.Millisecond

	updates := make(chan RuntimeUpdate, 8)
	app.emit = func(eventName string, payload any) {
		if eventName != runtimeUpdateEvent {
			return
		}
		update, ok := payload.(RuntimeUpdate)
		if !ok {
			t.Fatalf("payload type = %T, want RuntimeUpdate", payload)
		}
		updates <- update
	}

	app.startup(context.Background())
	defer app.shutdown(context.Background())

	payload := appcore.NewOfflineGamePayload{
		PlayerNames: []string{"Mesa", "CPU-2"},
		CPUFlags:    []bool{false, true},
	}
	if err := app.runtime.DispatchIntent(appcore.AppIntent{
		Kind:    appcore.IntentNewOfflineGame,
		Payload: mustJSONPayload(t, payload),
	}); err != nil {
		t.Fatalf("DispatchIntent: %v", err)
	}

	select {
	case update := <-updates:
		if update.Bundle.Mode != appcore.ModeOfflineMatch {
			t.Fatalf("mode = %q, want %q", update.Bundle.Mode, appcore.ModeOfflineMatch)
		}
		if len(update.Events) == 0 {
			t.Fatal("expected async update to include drained runtime events")
		}
	case <-time.After(800 * time.Millisecond):
		t.Fatal("timed out waiting for async runtime update")
	}
}

func TestPlayFaceDownCardUsesFaceDownPath(t *testing.T) {
	app := NewApp()
	// Dispatch manually to use a fixed seed where it is local player's turn to play first
	if err := app.dispatch(appcore.IntentNewOfflineGame, appcore.NewOfflineGamePayload{
		PlayerNames: []string{"Mesa", "CPU-2"},
		CPUFlags:    []bool{false, true},
		SeedLo:      42,
		SeedHi:      42,
	}); err != nil {
		t.Fatalf("StartOfflineGame: %v", err)
	}
	waitForPlayableFirstTrick(t, app)

	err := app.PlayFaceDownCard(0)
	if err == nil {
		t.Fatal("expected first-trick face-down play to fail")
	}
	if err.Code != "dispatch_failed" {
		t.Fatalf("error code = %q, want dispatch_failed", err.Code)
	}
	if !strings.Contains(strings.ToLower(err.Message), "virada") {
		t.Fatalf("error message = %q, want face-down validation", err.Message)
	}
}

func TestResetAliasesCloseSession(t *testing.T) {
	app := NewApp()
	if err := app.StartOfflineGame("Mesa", 2); err != nil {
		t.Fatalf("StartOfflineGame: %v", err)
	}
	if err := app.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	snapshot := app.Snapshot()
	if snapshot.Mode != appcore.ModeIdle {
		t.Fatalf("mode = %q, want %q", snapshot.Mode, appcore.ModeIdle)
	}
	if snapshot.Match != nil {
		t.Fatal("expected match snapshot cleared on reset")
	}
	if snapshot.Lobby != nil {
		t.Fatal("expected lobby snapshot cleared on reset")
	}
}

func TestCreateHostSessionTransitionsToLobby(t *testing.T) {
	app := NewApp()
	if err := app.CreateHostSession("Mesa", 2, "127.0.0.1:0", "", "tcp_tls"); err != nil {
		t.Fatalf("CreateHostSession: %v", err)
	}
	defer func() { _ = app.CloseSession() }()

	waitForMode(t, app, appcore.ModeHostLobby)

	snapshot := app.Snapshot()
	if snapshot.Lobby == nil {
		t.Fatal("expected lobby snapshot after host creation")
	}
	if snapshot.Lobby.InviteKey == "" {
		t.Fatal("expected invite key after host creation")
	}
	if snapshot.Connection.LastEventSeq == 0 {
		t.Fatal("expected event sequence to advance after host creation")
	}
}

func TestJoinSessionTransitionsClientToLobby(t *testing.T) {
	host := NewApp()
	if err := host.CreateHostSession("Mesa", 2, "127.0.0.1:0", "", "tcp_tls"); err != nil {
		t.Fatalf("CreateHostSession: %v", err)
	}
	defer func() { _ = host.CloseSession() }()
	waitForMode(t, host, appcore.ModeHostLobby)

	key := host.Snapshot().Lobby.InviteKey
	if key == "" {
		t.Fatal("expected invite key for join test")
	}

	client := NewApp()
	defer func() { _ = client.CloseSession() }()
	if err := client.JoinSession(key, "Visitante", "auto"); err != nil {
		t.Fatalf("JoinSession: %v", err)
	}

	waitForMode(t, client, appcore.ModeClientLobby)
}

func TestStartHostedMatchTransitionsHostAndClientToMatch(t *testing.T) {
	host := NewApp()
	if err := host.CreateHostSession("Mesa", 2, "127.0.0.1:0", "", "tcp_tls"); err != nil {
		t.Fatalf("CreateHostSession: %v", err)
	}
	defer func() { _ = host.CloseSession() }()
	waitForMode(t, host, appcore.ModeHostLobby)

	key := host.Snapshot().Lobby.InviteKey
	client := NewApp()
	defer func() { _ = client.CloseSession() }()
	if err := client.JoinSession(key, "Visitante", "auto"); err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	waitForMode(t, client, appcore.ModeClientLobby)

	if err := host.StartHostedMatch(); err != nil {
		t.Fatalf("StartHostedMatch: %v", err)
	}

	waitForMode(t, host, appcore.ModeHostMatch)
	waitForMode(t, client, appcore.ModeClientMatch)

	hostSnapshot := host.Snapshot()
	clientSnapshot := client.Snapshot()
	if hostSnapshot.Match == nil {
		t.Fatal("expected host match snapshot after starting hosted match")
	}
	if clientSnapshot.Match == nil {
		t.Fatal("expected client match snapshot after starting hosted match")
	}
	if !clientSnapshot.Connection.IsOnline {
		t.Fatal("expected client connection to stay online during hosted match")
	}
}

func TestHostedLobbyChatArrivesAsSystemEvent(t *testing.T) {
	host := NewApp()
	if err := host.CreateHostSession("Mesa", 2, "127.0.0.1:0", "", "tcp_tls"); err != nil {
		t.Fatalf("CreateHostSession: %v", err)
	}
	defer func() { _ = host.CloseSession() }()
	waitForMode(t, host, appcore.ModeHostLobby)

	key := host.Snapshot().Lobby.InviteKey
	client := NewApp()
	defer func() { _ = client.CloseSession() }()
	if err := client.JoinSession(key, "Visitante", "auto"); err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	waitForMode(t, client, appcore.ModeClientLobby)

	if err := client.SendChat("oi da mesa"); err != nil {
		t.Fatalf("SendChat: %v", err)
	}

	waitForEventText(t, host, 3*time.Second, "[chat] Visitante: oi da mesa")
	waitForEventText(t, client, 3*time.Second, "[chat] Visitante: oi da mesa")
}

func TestVoteHostUpdatesLobbyHostSeat(t *testing.T) {
	host := NewApp()
	if err := host.CreateHostSession("Mesa", 2, "127.0.0.1:0", "", "tcp_tls"); err != nil {
		t.Fatalf("CreateHostSession: %v", err)
	}
	defer func() { _ = host.CloseSession() }()
	waitForMode(t, host, appcore.ModeHostLobby)

	key := host.Snapshot().Lobby.InviteKey
	client := NewApp()
	defer func() { _ = client.CloseSession() }()
	if err := client.JoinSession(key, "Visitante", "auto"); err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	waitForMode(t, client, appcore.ModeClientLobby)

	if err := host.VoteHost(1); err != nil {
		t.Fatalf("host VoteHost: %v", err)
	}
	if err := client.VoteHost(1); err != nil {
		t.Fatalf("client VoteHost: %v", err)
	}

	waitForLobbyHostSeat(t, host, 1)
	waitForLobbyHostSeat(t, client, 1)
}

func TestReplacementInviteFlowWorksAfterDisconnect(t *testing.T) {
	host := NewApp()
	if err := host.CreateHostSession("Mesa", 2, "127.0.0.1:0", "", "tcp_tls"); err != nil {
		t.Fatalf("CreateHostSession: %v", err)
	}
	defer func() { _ = host.CloseSession() }()
	waitForMode(t, host, appcore.ModeHostLobby)

	key := host.Snapshot().Lobby.InviteKey
	client := NewApp()
	if err := client.JoinSession(key, "Visitante", "auto"); err != nil {
		t.Fatalf("JoinSession: %v", err)
	}
	waitForMode(t, client, appcore.ModeClientLobby)

	if err := host.StartHostedMatch(); err != nil {
		t.Fatalf("StartHostedMatch: %v", err)
	}
	waitForMode(t, host, appcore.ModeHostMatch)
	waitForMode(t, client, appcore.ModeClientMatch)

	if err := client.CloseSession(); err != nil {
		t.Fatalf("CloseSession client: %v", err)
	}
	waitForDisconnectedSeat(t, host, 1)

	if err := host.RequestReplacementInvite(1); err != nil {
		t.Fatalf("RequestReplacementInvite: %v", err)
	}

	inviteEvent := waitForEventKind(t, host, 3*time.Second, appcore.EventReplacementInvite)
	inviteKey := payloadString(inviteEvent, "invite_key")
	if inviteKey == "" {
		t.Fatal("expected replacement invite key in event payload")
	}

	sub := NewApp()
	defer func() { _ = sub.CloseSession() }()
	if err := sub.JoinSession(inviteKey, "Sub", "auto"); err != nil {
		t.Fatalf("JoinSession replacement: %v", err)
	}
	waitForModeOneOf(t, sub, appcore.ModeClientLobby, appcore.ModeClientMatch)
	waitForAssignedSeat(t, sub, 1)
}

func TestCloseSessionReturnsIdleAfterOnlineLobby(t *testing.T) {
	host := NewApp()
	if err := host.CreateHostSession("Mesa", 2, "127.0.0.1:0", "", "tcp_tls"); err != nil {
		t.Fatalf("CreateHostSession: %v", err)
	}
	waitForMode(t, host, appcore.ModeHostLobby)

	if err := host.CloseSession(); err != nil {
		t.Fatalf("CloseSession: %v", err)
	}

	snapshot := host.Snapshot()
	if snapshot.Mode != appcore.ModeIdle {
		t.Fatalf("mode = %q, want %q", snapshot.Mode, appcore.ModeIdle)
	}
	if snapshot.Lobby != nil {
		t.Fatal("expected lobby snapshot cleared after online close")
	}
	if snapshot.Match != nil {
		t.Fatal("expected match snapshot cleared after online close")
	}
	if snapshot.Connection.IsOnline {
		t.Fatal("expected connection to be offline after online close")
	}
}

func mustJSONPayload(t *testing.T, payload any) []byte {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return raw
}

func waitForPlayableFirstTrick(t *testing.T, app *App) {
	t.Helper()

	for range 100 {
		time.Sleep(5 * time.Millisecond)
		snapshot := app.Snapshot()
		if snapshot.Match != nil && snapshot.Match.CurrentHand.Round == 1 && snapshot.UI.Actions.CanPlayCard {
			return
		}
		if err := app.Tick(12); err != nil {
			t.Fatalf("Tick: %v", err)
		}
	}

	snapshot := app.Snapshot()
	if snapshot.Match == nil {
		t.Fatal("match snapshot unavailable while waiting for playable first trick")
	}
	t.Fatalf("timed out waiting for playable first trick: round=%d canPlay=%v", snapshot.Match.CurrentHand.Round, snapshot.UI.Actions.CanPlayCard)
}

func waitForMode(t *testing.T, app *App, want string) {
	t.Helper()

	for range 40 {
		snapshot := app.Snapshot()
		if snapshot.Mode == want {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	snapshot := app.Snapshot()
	t.Fatalf("timed out waiting for mode %q: got %q", want, snapshot.Mode)
}

func waitForModeOneOf(t *testing.T, app *App, wants ...string) {
	t.Helper()

	for range 120 {
		snapshot := app.Snapshot()
		for _, want := range wants {
			if snapshot.Mode == want {
				return
			}
		}
		time.Sleep(25 * time.Millisecond)
	}

	snapshot := app.Snapshot()
	t.Fatalf("timed out waiting for modes %v: got %q", wants, snapshot.Mode)
}

func waitForLobbyHostSeat(t *testing.T, app *App, want int) {
	t.Helper()

	for range 120 {
		snapshot := app.Snapshot()
		if snapshot.Lobby != nil && snapshot.Lobby.HostSeat == want {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	snapshot := app.Snapshot()
	if snapshot.Lobby == nil {
		t.Fatalf("timed out waiting for host seat %d: lobby snapshot missing", want)
	}
	t.Fatalf("timed out waiting for host seat %d: got %d", want, snapshot.Lobby.HostSeat)
}

func waitForDisconnectedSeat(t *testing.T, app *App, seat int) {
	t.Helper()

	for range 160 {
		snapshot := app.Snapshot()
		if snapshot.Lobby != nil && seat < len(snapshot.UI.LobbySlots) && !snapshot.UI.LobbySlots[seat].IsConnected {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	snapshot := app.Snapshot()
	connected := true
	if seat < len(snapshot.UI.LobbySlots) {
		connected = snapshot.UI.LobbySlots[seat].IsConnected
	}
	t.Fatalf("timed out waiting for seat %d disconnect: connected=%v", seat, connected)
}

func waitForAssignedSeat(t *testing.T, app *App, want int) {
	t.Helper()

	for range 120 {
		snapshot := app.Snapshot()
		if snapshot.Lobby != nil && snapshot.Lobby.AssignedSeat == want {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	snapshot := app.Snapshot()
	if snapshot.Lobby == nil {
		t.Fatalf("timed out waiting for assigned seat %d: lobby snapshot missing", want)
	}
	t.Fatalf("timed out waiting for assigned seat %d: got %d", want, snapshot.Lobby.AssignedSeat)
}

func waitForEventKind(t *testing.T, app *App, timeout time.Duration, kind string) appcore.AppEvent {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, event := range app.PollEvents() {
			if event.Kind == kind {
				return event
			}
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for event %q", kind)
	return appcore.AppEvent{}
}

func waitForEventText(t *testing.T, app *App, timeout time.Duration, want string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, event := range app.PollEvents() {
			if strings.Contains(payloadText(event), want) {
				return
			}
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for event text %q", want)
}

func payloadText(event appcore.AppEvent) string {
	if payload, ok := event.Payload.(map[string]any); ok {
		if text, ok := payload["text"].(string); ok {
			return text
		}
	}
	return fmt.Sprint(event.Payload)
}

func payloadString(event appcore.AppEvent, key string) string {
	if payload, ok := event.Payload.(map[string]any); ok {
		if value, ok := payload[key].(string); ok {
			return value
		}
	}
	return ""
}
