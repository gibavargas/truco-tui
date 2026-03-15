package appcore

import (
	"encoding/json"
	"testing"

	"truco-tui/internal/truco"
)

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

func TestRuntimeSnapshotIncludesDerivedActionState(t *testing.T) {
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
	if state.UI.Actions.LocalPlayerID != 0 {
		t.Fatalf("LocalPlayerID = %d, want 0", state.UI.Actions.LocalPlayerID)
	}
	if state.UI.Actions.MustRespond {
		t.Fatal("did not expect pending truco response at start of offline match")
	}
	if state.Match.TurnPlayer == state.Match.CurrentPlayerIdx && !state.UI.Actions.CanPlayCard {
		t.Fatal("expected play action to be available when it is the local turn with no pending raise")
	}
	if !state.UI.Actions.CanCloseSession {
		t.Fatal("expected close session to be available outside idle mode")
	}
}

func TestRuntimeSnapshotIncludesLobbySlotParityState(t *testing.T) {
	game, err := truco.NewGame([]string{"Host", "Guest"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if !game.SetPlayerCPU(1, true, true) {
		t.Fatal("expected provisional CPU toggle to succeed")
	}

	rt := NewRuntime()
	rt.mode = "host_match"
	rt.game = game
	rt.localSeat = 0
	rt.lobby = &LobbySnapshot{
		InviteKey:      "abc",
		Slots:          []string{"Host", "Guest"},
		AssignedSeat:   0,
		NumPlayers:     2,
		Started:        true,
		HostSeat:       0,
		ConnectedSeats: map[int]bool{0: true, 1: false},
	}
	rt.match = ptrMatch(game.Snapshot(0))

	state := rt.SnapshotBundle()
	if len(state.UI.LobbySlots) != 2 {
		t.Fatalf("LobbySlots = %d, want 2", len(state.UI.LobbySlots))
	}
	target := state.UI.LobbySlots[1]
	if target.Status != "provisional_cpu" {
		t.Fatalf("slot status = %q, want provisional_cpu", target.Status)
	}
	if !target.CanVoteHost {
		t.Fatal("expected occupied remote seat to be votable")
	}
	if !target.CanRequestReplacement {
		t.Fatal("expected disconnected occupied seat to allow replacement invite")
	}
	if !target.IsProvisionalCPU {
		t.Fatal("expected provisional CPU marker on disconnected seat")
	}
}

func ptrMatch(s truco.Snapshot) *truco.Snapshot {
	return &s
}
