//go:build wails

package main

import (
	"context"
	"encoding/json"
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
	if err := app.StartOfflineGame("Mesa", 2); err != nil {
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

	for range 8 {
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
