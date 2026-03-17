package main

import (
	"encoding/json"
	"testing"

	"truco-tui/internal/appcore"
)

func TestVersionsJSONMatchesRuntimeContract(t *testing.T) {
	var versions appcore.CoreVersions
	if err := json.Unmarshal([]byte(consumeCString(versionsJSON())), &versions); err != nil {
		t.Fatalf("unmarshal versions: %v", err)
	}
	if versions.CoreAPIVersion != appcore.CoreAPIVersion {
		t.Fatalf("core_api_version = %d, want %d", versions.CoreAPIVersion, appcore.CoreAPIVersion)
	}
	if versions.SnapshotSchema != appcore.SnapshotSchemaMajor {
		t.Fatalf("snapshot_schema_version = %d, want %d", versions.SnapshotSchema, appcore.SnapshotSchemaMajor)
	}
}

func TestDispatchIntentJSONRejectsInvalidPayload(t *testing.T) {
	handle := createRuntimeHandle()
	defer destroyRuntimeHandle(handle)

	raw := consumeCString(dispatchIntentJSON(handle, "{invalid"))
	var out map[string]string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("unmarshal invalid_json result: %v", err)
	}
	if out["code"] != "invalid_json" {
		t.Fatalf("code = %q, want invalid_json", out["code"])
	}
}

func TestFFIRuntimeLifecycleProducesSnapshotAndEvents(t *testing.T) {
	handle := createRuntimeHandle()
	defer destroyRuntimeHandle(handle)

	intent := map[string]any{
		"kind": appcore.IntentNewOfflineGame,
		"payload": map[string]any{
			"player_names": []string{"Ana", "CPU-2"},
			"cpu_flags":    []bool{false, true},
			"seed_lo":      7,
			"seed_hi":      9,
		},
	}
	b, err := json.Marshal(intent)
	if err != nil {
		t.Fatalf("marshal intent: %v", err)
	}
	if got := dispatchIntentJSON(handle, string(b)); got != nil {
		t.Fatalf("dispatch returned error: %s", consumeCString(got))
	}

	var bundle appcore.SnapshotBundle
	if err := json.Unmarshal([]byte(consumeCString(snapshotJSON(handle))), &bundle); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if bundle.Mode != appcore.ModeOfflineMatch {
		t.Fatalf("mode = %q, want %q", bundle.Mode, appcore.ModeOfflineMatch)
	}
	if bundle.Match == nil {
		t.Fatal("match snapshot is nil")
	}

	foundMatchUpdated := false
	foundSessionReady := false
	for {
		raw := consumeCString(pollEventJSON(handle))
		if raw == "" {
			break
		}
		var ev appcore.AppEvent
		if err := json.Unmarshal([]byte(raw), &ev); err != nil {
			t.Fatalf("unmarshal event: %v", err)
		}
		switch ev.Kind {
		case appcore.EventMatchUpdated:
			foundMatchUpdated = true
		case appcore.EventSessionReady:
			foundSessionReady = true
		}
	}
	if !foundMatchUpdated {
		t.Fatal("expected match_updated event")
	}
	if !foundSessionReady {
		t.Fatal("expected session_ready event")
	}
}
