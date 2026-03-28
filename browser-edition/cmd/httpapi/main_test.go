package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"truco-tui/internal/appcore"
)

func postActionHTTP(t *testing.T, srv http.Handler, action, sessionID string, body map[string]interface{}) (int, map[string]interface{}) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(http.MethodPost, "/api/"+action, &buf)
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("X-Session-ID", sessionID)
	}
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response for %s: %v\nbody: %s", action, err, w.Body.String())
	}
	return w.Code, result
}

func postAction(t *testing.T, srv http.Handler, action, sessionID string, body map[string]interface{}) map[string]interface{} {
	t.Helper()
	_, result := postActionHTTP(t, srv, action, sessionID, body)
	return result
}

func createSession(t *testing.T, srv http.Handler) string {
	t.Helper()
	res := postAction(t, srv, "createSession", "", nil)
	if !res["ok"].(bool) {
		t.Fatalf("createSession failed: %v", res["error"])
	}
	sid, _ := res["sessionId"].(string)
	if sid == "" {
		t.Fatalf("missing sessionId: %v", res)
	}
	return sid
}

func parseBundle(t *testing.T, res map[string]interface{}) map[string]interface{} {
	t.Helper()
	bundle, ok := res["bundle"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected bundle in result, got %v", res["bundle"])
	}
	return bundle
}

func parseSnapshot(t *testing.T, res map[string]interface{}) map[string]interface{} {
	t.Helper()
	raw, ok := res["snapshot"].(string)
	if !ok || raw == "" {
		t.Fatalf("expected snapshot JSON string in result")
	}
	var snap map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	return snap
}

func eventKinds(t *testing.T, res map[string]interface{}) []string {
	t.Helper()
	raw, ok := res["events"].([]interface{})
	if !ok {
		t.Fatalf("expected events array, got %T", res["events"])
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		ev, ok := item.(map[string]interface{})
		if !ok {
			t.Fatalf("expected event object, got %T", item)
		}
		kind, _ := ev["kind"].(string)
		out = append(out, kind)
	}
	return out
}

func requireSessionNetwork(t *testing.T, res map[string]interface{}) map[string]interface{} {
	t.Helper()
	session, ok := res["session"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected session in response")
	}
	network, ok := session["network"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected session.network in response, got %v", session["network"])
	}
	return network
}

func containsKind(kinds []string, target string) bool {
	for _, kind := range kinds {
		if kind == target {
			return true
		}
	}
	return false
}

func createAndStart(t *testing.T, srv http.Handler) string {
	t.Helper()
	sid := createSession(t, srv)
	res := postAction(t, srv, "startGame", sid, map[string]interface{}{
		"numPlayers": 2,
		"name":       "Tester",
	})
	if !res["ok"].(bool) {
		t.Fatalf("startGame failed: %v", res["error"])
	}
	return sid
}

func advanceBrowserSessionUntilPlayableRound(t *testing.T, srv http.Handler, sid string, minRound int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		res := postAction(t, srv, "snapshot", sid, nil)
		snap := parseSnapshot(t, res)
		hand := snap["CurrentHand"].(map[string]interface{})
		round := int(hand["Round"].(float64))
		pending, _ := snap["PendingRaiseFor"].(float64)
		turn := int(snap["TurnPlayer"].(float64))
		roundCards, _ := hand["RoundCards"].([]interface{})

		if round >= minRound && turn == 0 && int(pending) == -1 && len(roundCards) == 0 {
			return
		}
		if int(pending) == 0 {
			res = postAction(t, srv, "accept", sid, nil)
			if !res["ok"].(bool) {
				t.Fatalf("accept pending truco failed while advancing browser round: %v", res["error"])
			}
			continue
		}
		if turn == 0 {
			res = postAction(t, srv, "play", sid, map[string]interface{}{"cardIndex": 0})
			if !res["ok"].(bool) {
				t.Fatalf("play while advancing browser round failed: %v", res["error"])
			}
			continue
		}
		_ = postAction(t, srv, "autoCpuLoopTick", sid, nil)
	}
	t.Fatalf("timeout advancing browser session to playable round %d", minRound)
}

func TestCreateSession(t *testing.T) {
	srv := newAPIServer()
	_ = createSession(t, srv)
}

func TestStartGameReturnsBundleSnapshotAndContractVersions(t *testing.T) {
	srv := newAPIServer()
	sid := createSession(t, srv)

	res := postAction(t, srv, "startGame", sid, map[string]interface{}{
		"numPlayers": 2,
		"name":       "Tester",
	})
	if !res["ok"].(bool) {
		t.Fatalf("startGame failed: %v", res["error"])
	}
	if res["mode"] != appcore.ModeOfflineMatch {
		t.Fatalf("mode = %v, want %q", res["mode"], appcore.ModeOfflineMatch)
	}

	bundle := parseBundle(t, res)
	versions := bundle["versions"].(map[string]interface{})
	if int(versions["core_api_version"].(float64)) != appcore.CoreAPIVersion {
		t.Fatalf("core_api_version = %v, want %d", versions["core_api_version"], appcore.CoreAPIVersion)
	}
	if int(versions["snapshot_schema_version"].(float64)) != appcore.SnapshotSchemaMajor {
		t.Fatalf("snapshot_schema_version = %v, want %d", versions["snapshot_schema_version"], appcore.SnapshotSchemaMajor)
	}

	snap := parseSnapshot(t, res)
	if int(snap["NumPlayers"].(float64)) != 2 {
		t.Fatalf("NumPlayers = %v, want 2", snap["NumPlayers"])
	}
}

func TestSetLocaleDrainsRuntimeEventsExactlyOnce(t *testing.T) {
	srv := newAPIServer()
	sid := createSession(t, srv)

	res := postAction(t, srv, "setLocale", sid, map[string]interface{}{
		"locale": appcore.LocaleENUS,
	})
	if !res["ok"].(bool) {
		t.Fatalf("setLocale failed: %v", res["error"])
	}
	kinds := eventKinds(t, res)
	if !containsKind(kinds, appcore.EventLocaleChanged) {
		t.Fatalf("expected %q in events, got %v", appcore.EventLocaleChanged, kinds)
	}

	res = postAction(t, srv, "pullOnlineEvents", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("pullOnlineEvents failed: %v", res["error"])
	}
	kinds = eventKinds(t, res)
	if len(kinds) != 0 {
		t.Fatalf("expected drained queue to be empty, got %v", kinds)
	}
}

func TestPlayRequiresCardIndexErrorCode(t *testing.T) {
	srv := newAPIServer()
	sid := createSession(t, srv)
	_ = postAction(t, srv, "startGame", sid, map[string]interface{}{
		"numPlayers": 2,
		"name":       "Tester",
	})

	res := postAction(t, srv, "play", sid, nil)
	if res["ok"].(bool) {
		t.Fatalf("expected play without cardIndex to fail")
	}
	if res["error_code"] != "missing_card_index" {
		t.Fatalf("error_code = %v, want missing_card_index", res["error_code"])
	}
}

func TestPlayFaceDownMasksBrowserSnapshot(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)
	advanceBrowserSessionUntilPlayableRound(t, srv, sid, 2)

	res := postAction(t, srv, "play", sid, map[string]interface{}{
		"cardIndex": 0,
		"faceDown":  true,
	})
	if !res["ok"].(bool) {
		t.Fatalf("play face-down failed: %v", res["error"])
	}

	snap := parseSnapshot(t, res)
	roundCardsRaw := snap["CurrentHand"].(map[string]interface{})["RoundCards"]
	roundCards, ok := roundCardsRaw.([]interface{})
	if !ok || len(roundCards) != 1 {
		t.Fatalf("round cards = %v, want one played card", roundCardsRaw)
	}
	played := roundCards[0].(map[string]interface{})
	if played["FaceDown"] != true {
		t.Fatalf("expected FaceDown=true, got %v", played["FaceDown"])
	}
	card := played["Card"].(map[string]interface{})
	if rank, _ := card["Rank"].(string); rank != "" {
		t.Fatalf("masked browser snapshot leaked rank %q", rank)
	}
}

func TestRuntimeErrorsExposeNormalizedErrorCodeAndBundle(t *testing.T) {
	srv := newAPIServer()
	sid := createSession(t, srv)

	code, res := postActionHTTP(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
		"relay_url":  "://bad relay url",
	})
	if res["ok"].(bool) {
		t.Fatalf("expected invalid relay URL to fail")
	}
	if code == http.StatusOK {
		t.Fatalf("status = %d, want non-OK", code)
	}
	if res["error_code"] != "create_host_failed" {
		t.Fatalf("error_code = %v, want create_host_failed", res["error_code"])
	}
	bundle := parseBundle(t, res)
	connection := bundle["connection"].(map[string]interface{})
	lastError := connection["last_error"].(map[string]interface{})
	if lastError["code"] != "create_host_failed" {
		t.Fatalf("last_error.code = %v, want create_host_failed", lastError["code"])
	}
}

func TestStartOnlineHostIncludesNetworkSnapshot(t *testing.T) {
	srv := newAPIServer()
	sid := createSession(t, srv)

	res := postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
	})
	if !res["ok"].(bool) {
		t.Fatalf("startOnlineHost failed: %v", res["error"])
	}

	network := requireSessionNetwork(t, res)
	if network["transport"] != "tcp_tls" {
		t.Fatalf("transport = %v, want tcp_tls", network["transport"])
	}
	if mixed, _ := network["mixed_protocol_session"].(bool); mixed {
		t.Fatalf("mixed_protocol_session = %v, want false", mixed)
	}
	seatVersions, ok := network["seat_protocol_versions"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected seat_protocol_versions, got %v", network["seat_protocol_versions"])
	}
	if seatVersions["0"] != float64(2) {
		t.Fatalf("seat 0 protocol = %v, want 2", seatVersions["0"])
	}
}

func TestJoinOnlineIncludesNegotiatedProtocolVersion(t *testing.T) {
	srv := newAPIServer()
	hostSID := createSession(t, srv)
	clientSID := createSession(t, srv)

	hostRes := postAction(t, srv, "startOnlineHost", hostSID, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
	})
	if !hostRes["ok"].(bool) {
		t.Fatalf("startOnlineHost failed: %v", hostRes["error"])
	}
	session := hostRes["session"].(map[string]interface{})
	key := session["inviteKey"].(string)

	res := postAction(t, srv, "joinOnline", clientSID, map[string]interface{}{
		"name": "Guest",
		"key":  key,
		"role": "auto",
	})
	if !res["ok"].(bool) {
		t.Fatalf("joinOnline failed: %v", res["error"])
	}

	network := requireSessionNetwork(t, res)
	if network["transport"] != "tcp_tls" {
		t.Fatalf("transport = %v, want tcp_tls", network["transport"])
	}
	if network["negotiated_protocol_version"] != float64(2) {
		t.Fatalf("negotiated_protocol_version = %v, want 2", network["negotiated_protocol_version"])
	}
}

func TestResetReturnsRuntimeToIdleAndEmitsSessionClosed(t *testing.T) {
	srv := newAPIServer()
	sid := createSession(t, srv)
	_ = postAction(t, srv, "startGame", sid, map[string]interface{}{
		"numPlayers": 2,
		"name":       "Tester",
	})

	code, res := postActionHTTP(t, srv, "reset", sid, nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d, want %d", code, http.StatusOK)
	}
	if !res["ok"].(bool) {
		t.Fatalf("reset failed: %v", res["error"])
	}
	if res["mode"] != appcore.ModeIdle {
		t.Fatalf("mode = %v, want %q", res["mode"], appcore.ModeIdle)
	}
	bundle := parseBundle(t, res)
	if bundle["lobby"] != nil {
		t.Fatal("expected lobby to be cleared after reset")
	}
	kinds := eventKinds(t, res)
	if !containsKind(kinds, appcore.EventSessionClosed) {
		t.Fatalf("expected %q in events, got %v", appcore.EventSessionClosed, kinds)
	}
}

func TestNewHandStartsAnotherHandInSameMatch(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	res := postAction(t, srv, "newHand", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("newHand failed: %v", res["error"])
	}
	if res["mode"] != appcore.ModeOfflineMatch {
		t.Fatalf("mode = %v, want %q", res["mode"], appcore.ModeOfflineMatch)
	}
}

func TestAutoCpuLoopTickReturnsBundle(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	res := postAction(t, srv, "autoCpuLoopTick", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("autoCpuLoopTick failed: %v", res["error"])
	}
	if res["bundle"] == nil {
		t.Fatalf("expected bundle field in response")
	}
}

func TestCloseSessionDeletesBrowserSession(t *testing.T) {
	srv := newAPIServer()
	sid := createSession(t, srv)

	code, res := postActionHTTP(t, srv, "closeSession", sid, nil)
	if code != http.StatusOK {
		t.Fatalf("status = %d, want %d", code, http.StatusOK)
	}
	if !res["ok"].(bool) {
		t.Fatalf("closeSession failed: %v", res["error"])
	}
	if closed, _ := res["sessionClosed"].(bool); !closed {
		t.Fatalf("sessionClosed = %v, want true", res["sessionClosed"])
	}

	code, res = postActionHTTP(t, srv, "snapshot", sid, nil)
	if code != http.StatusNotFound {
		t.Fatalf("status after close = %d, want %d", code, http.StatusNotFound)
	}
	if res["error_code"] != "session_not_found" {
		t.Fatalf("error_code after close = %v, want session_not_found", res["error_code"])
	}
}

func TestUnknownActionReturnsStructuredError(t *testing.T) {
	srv := newAPIServer()
	sid := createSession(t, srv)

	code, res := postActionHTTP(t, srv, "doesNotExist", sid, nil)
	if res["ok"].(bool) {
		t.Fatalf("expected unknown action to fail")
	}
	if code == http.StatusOK {
		t.Fatalf("status = %d, want non-OK", code)
	}
	if res["error_code"] != "unknown_action" {
		t.Fatalf("error_code = %v, want unknown_action", res["error_code"])
	}
}

func TestMissingSessionReturnsNotFound(t *testing.T) {
	srv := newAPIServer()
	code, res := postActionHTTP(t, srv, "snapshot", "missing-session", nil)
	if code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", code, http.StatusNotFound)
	}
	if res["error_code"] != "session_not_found" {
		t.Fatalf("error_code = %v, want session_not_found", res["error_code"])
	}
}

func TestOnlyPostAllowed(t *testing.T) {
	srv := newAPIServer()
	req := httptest.NewRequest(http.MethodGet, "/api/createSession", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
	var res map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res["error_code"] != "method_not_allowed" {
		t.Fatalf("error_code = %v, want method_not_allowed", res["error_code"])
	}
}
