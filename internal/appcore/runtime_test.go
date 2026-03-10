package appcore

import (
	"encoding/json"
	"testing"
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
