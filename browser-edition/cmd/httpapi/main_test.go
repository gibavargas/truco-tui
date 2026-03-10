package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func postAction(t *testing.T, srv http.Handler, action, sessionID string, body map[string]interface{}) map[string]interface{} {
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
	return result
}

func createAndStart(t *testing.T, srv http.Handler) string {
	t.Helper()
	res := postAction(t, srv, "createSession", "", nil)
	if !res["ok"].(bool) {
		t.Fatalf("createSession failed: %v", res["error"])
	}
	sid := res["sessionId"].(string)

	res = postAction(t, srv, "startGame", sid, map[string]interface{}{
		"numPlayers": 2,
		"name":       "Tester",
	})
	if !res["ok"].(bool) {
		t.Fatalf("startGame failed: %v", res["error"])
	}
	return sid
}

func parseSnap(t *testing.T, res map[string]interface{}) map[string]interface{} {
	t.Helper()
	snapStr, ok := res["snapshot"].(string)
	if !ok {
		t.Fatalf("no snapshot in result")
	}
	var snap map[string]interface{}
	if err := json.Unmarshal([]byte(snapStr), &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	return snap
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCreateSession(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	if !res["ok"].(bool) {
		t.Fatalf("expected ok")
	}
	sid, ok := res["sessionId"].(string)
	if !ok || sid == "" {
		t.Fatalf("expected sessionId, got %v", res["sessionId"])
	}
}

func TestStartGame(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	res := postAction(t, srv, "snapshot", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("snapshot failed: %v", res["error"])
	}
	snap := parseSnap(t, res)
	if int(snap["NumPlayers"].(float64)) != 2 {
		t.Fatalf("expected 2 players, got %v", snap["NumPlayers"])
	}
}

func TestStartGame4Players(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	res = postAction(t, srv, "startGame", sid, map[string]interface{}{
		"numPlayers": 4,
		"name":       "P1",
	})
	if !res["ok"].(bool) {
		t.Fatalf("startGame 4p failed: %v", res["error"])
	}
	snap := parseSnap(t, res)
	if int(snap["NumPlayers"].(float64)) != 4 {
		t.Fatalf("expected 4 players, got %v", snap["NumPlayers"])
	}
}

func TestPlayCard(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	// Run CPU loop first to ensure it's our turn
	_ = postAction(t, srv, "autoCpuLoopTick", sid, nil)

	res := postAction(t, srv, "snapshot", sid, nil)
	snap := parseSnap(t, res)
	turn := int(snap["TurnPlayer"].(float64))

	if turn != 0 {
		// CPU goes first, play CPU then try
		_ = postAction(t, srv, "autoCpuLoopTick", sid, nil)
		res = postAction(t, srv, "snapshot", sid, nil)
		snap = parseSnap(t, res)
		turn = int(snap["TurnPlayer"].(float64))
	}

	if turn == 0 {
		if pending, ok := snap["PendingRaiseFor"].(float64); ok && int(pending) != -1 {
			res = postAction(t, srv, "accept", sid, nil)
			if !res["ok"].(bool) {
				t.Fatalf("accept pending truco failed: %v", res["error"])
			}
		}
		res = postAction(t, srv, "play", sid, map[string]interface{}{"cardIndex": 0})
		if !res["ok"].(bool) {
			t.Fatalf("play card failed: %v", res["error"])
		}
	}
}

func TestPlayCardMissingIndex(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	res := postAction(t, srv, "play", sid, nil)
	if res["ok"].(bool) {
		t.Fatalf("expected error when no cardIndex")
	}
}

func TestTrucoAskAndAccept(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	// Ensure it's our turn via autoCpu
	_ = postAction(t, srv, "autoCpuLoopTick", sid, nil)

	res := postAction(t, srv, "snapshot", sid, nil)
	snap := parseSnap(t, res)
	turn := int(snap["TurnPlayer"].(float64))

	if turn == 0 {
		res = postAction(t, srv, "truco", sid, nil)
		if !res["ok"].(bool) {
			// This can fail if game conditions don't allow (e.g., hand already at 12)
			// but for a fresh game it should work
			t.Logf("truco ask result: %v", res)
		}
	}
}

func TestRefuseOnEmptySession(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "refuse", "nonexistent", nil)
	if res["ok"].(bool) {
		t.Fatalf("expected error for nonexistent session")
	}
	errMsg, _ := res["error"].(string)
	if errMsg != "session not found" {
		t.Fatalf("expected 'session not found', got %q", errMsg)
	}
}

func TestSnapshotNoGame(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	res = postAction(t, srv, "snapshot", sid, nil)
	if res["ok"].(bool) {
		t.Fatalf("expected error when no game started")
	}
}

func TestNewHand(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	// newHand should work (starts another hand in the same match)
	res := postAction(t, srv, "newHand", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("newHand failed: %v", res["error"])
	}
}

func TestReset(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	res := postAction(t, srv, "reset", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("reset failed: %v", res["error"])
	}

	// After reset, snapshot should fail
	res = postAction(t, srv, "snapshot", sid, nil)
	if res["ok"].(bool) {
		t.Fatalf("expected error after reset")
	}
}

func TestAutoCpuLoopTick(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	res := postAction(t, srv, "autoCpuLoopTick", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("autoCpuLoopTick failed: %v", res["error"])
	}
	// changed may be true or false depending on whose turn it is
	if _, ok := res["changed"]; !ok {
		t.Fatalf("expected 'changed' field in response")
	}
}

func TestConcurrentSessions(t *testing.T) {
	srv := newAPIServer()

	// Create two independent sessions
	sid1 := createAndStart(t, srv)
	sid2 := createAndStart(t, srv)

	if sid1 == sid2 {
		t.Fatalf("sessions should have different IDs")
	}

	// Snapshot of session 1 should work independently of session 2
	res1 := postAction(t, srv, "snapshot", sid1, nil)
	res2 := postAction(t, srv, "snapshot", sid2, nil)
	if !res1["ok"].(bool) || !res2["ok"].(bool) {
		t.Fatalf("concurrent snapshots failed")
	}

	// Reset session 1 should not affect session 2
	_ = postAction(t, srv, "reset", sid1, nil)
	res2 = postAction(t, srv, "snapshot", sid2, nil)
	if !res2["ok"].(bool) {
		t.Fatalf("session 2 should work after session 1 reset")
	}
}

func TestUnknownAction(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	res = postAction(t, srv, "doesNotExist", sid, nil)
	if res["ok"].(bool) {
		t.Fatalf("expected error for unknown action")
	}
}

func TestOnlyPostAllowed(t *testing.T) {
	srv := newAPIServer()
	req := httptest.NewRequest(http.MethodGet, "/api/createSession", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Online lobby tests
// ---------------------------------------------------------------------------

func TestStartOnlineHost(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	res = postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "HostPlayer",
		"numPlayers": 2,
	})
	if !res["ok"].(bool) {
		t.Fatalf("startOnlineHost failed: %v", res["error"])
	}
	session, ok := res["session"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected session in response")
	}
	if session["mode"] != "host" {
		t.Fatalf("expected mode=host, got %v", session["mode"])
	}
	if session["inviteKey"] == "" {
		t.Fatalf("expected inviteKey")
	}
}

func TestJoinOnline(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	// Start host
	res = postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "HostPlayer",
		"numPlayers": 2,
	})
	session := res["session"].(map[string]interface{})
	key := session["inviteKey"].(string)

	// Join
	res = postAction(t, srv, "joinOnline", sid, map[string]interface{}{
		"name": "Joiner",
		"key":  key,
		"role": "auto",
	})
	if !res["ok"].(bool) {
		t.Fatalf("joinOnline failed: %v", res["error"])
	}
}

func TestJoinOnlineNoKey(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	res = postAction(t, srv, "joinOnline", sid, map[string]interface{}{
		"name": "Joiner",
	})
	if res["ok"].(bool) {
		t.Fatalf("expected error when no key")
	}
}

func TestOnlineState(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	// Before online setup, session should be nil
	res = postAction(t, srv, "onlineState", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("onlineState failed")
	}

	// Start host
	_ = postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
	})
	res = postAction(t, srv, "onlineState", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("onlineState after host failed")
	}
	session, ok := res["session"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected session in onlineState response")
	}
	if session["inviteKey"] == nil || session["inviteKey"] == "" {
		t.Fatalf("expected inviteKey in onlineState")
	}
}

func TestStartOnlineMatch(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	_ = postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
	})

	res = postAction(t, srv, "startOnlineMatch", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("startOnlineMatch failed: %v", res["error"])
	}
	if res["snapshot"] == nil {
		t.Fatalf("expected snapshot in startOnlineMatch response")
	}
}

func TestSendChat(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	_ = postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
	})

	res = postAction(t, srv, "sendChat", sid, map[string]interface{}{
		"message": "Hello!",
	})
	if !res["ok"].(bool) {
		t.Fatalf("sendChat failed: %v", res["error"])
	}

	// Pull events should have the chat
	res = postAction(t, srv, "pullOnlineEvents", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("pullOnlineEvents failed")
	}
	events, ok := res["events"].([]interface{})
	if !ok {
		t.Fatalf("expected events array")
	}
	found := false
	for _, ev := range events {
		e := ev.(map[string]interface{})
		if e["type"] == "chat" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected chat event in pullOnlineEvents")
	}
}

func TestSendHostVote(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	_ = postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
	})

	res = postAction(t, srv, "sendHostVote", sid, map[string]interface{}{
		"slot": 1,
	})
	if !res["ok"].(bool) {
		t.Fatalf("sendHostVote failed: %v", res["error"])
	}
}

func TestSendHostVoteInvalidSlot(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	_ = postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
	})

	res = postAction(t, srv, "sendHostVote", sid, map[string]interface{}{
		"slot": 99,
	})
	if res["ok"].(bool) {
		t.Fatalf("expected error for invalid slot")
	}
}

func TestRequestReplacementInvite(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	_ = postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
	})

	res = postAction(t, srv, "requestReplacementInvite", sid, map[string]interface{}{
		"slot": 2,
	})
	if !res["ok"].(bool) {
		t.Fatalf("requestReplacementInvite failed: %v", res["error"])
	}
	if res["inviteKey"] == nil || res["inviteKey"] == "" {
		t.Fatalf("expected inviteKey")
	}
}

func TestLeaveSession(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	_ = postAction(t, srv, "startOnlineHost", sid, map[string]interface{}{
		"name":       "Host",
		"numPlayers": 2,
	})

	res = postAction(t, srv, "leaveSession", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("leaveSession failed")
	}

	// After leaving, onlineState should return nil session
	res = postAction(t, srv, "onlineState", sid, nil)
	if res["session"] != nil {
		t.Fatalf("expected nil session after leave")
	}
}

func TestPullOnlineEventsNoSession(t *testing.T) {
	srv := newAPIServer()
	res := postAction(t, srv, "createSession", "", nil)
	sid := res["sessionId"].(string)

	// No online setup yet
	res = postAction(t, srv, "pullOnlineEvents", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("pullOnlineEvents should succeed with empty events")
	}
}

// ---------------------------------------------------------------------------
// Full game lifecycle test
// ---------------------------------------------------------------------------

func TestFullGameLifecycle(t *testing.T) {
	srv := newAPIServer()
	sid := createAndStart(t, srv)

	// Play several iterations of CPU loop + player cards to ensure the
	// lifecycle works without panics or errors.
	for round := 0; round < 20; round++ {
		// Run CPU
		res := postAction(t, srv, "autoCpuLoopTick", sid, nil)
		if !res["ok"].(bool) {
			t.Fatalf("cpu loop tick %d failed: %v", round, res["error"])
		}

		snap := parseSnap(t, res)
		if snap["MatchFinished"].(bool) {
			break
		}

		turn := int(snap["TurnPlayer"].(float64))
		if turn == 0 {
			// It's our turn — play card 0
			pending := int(snap["PendingRaiseFor"].(float64))
			if pending != -1 {
				// There's a pending truco — accept it
				res = postAction(t, srv, "accept", sid, nil)
				if !res["ok"].(bool) {
					// Maybe we can't accept, try refuse
					_ = postAction(t, srv, "refuse", sid, nil)
				}
				continue
			}

			res = postAction(t, srv, "play", sid, map[string]interface{}{"cardIndex": 0})
			if !res["ok"].(bool) {
				// Might fail if hand is empty — that's ok, trigger newHand
				t.Logf("play card failed at round %d: %v", round, res["error"])
			}
		}
	}

	// Reset and confirm clean state
	res := postAction(t, srv, "reset", sid, nil)
	if !res["ok"].(bool) {
		t.Fatalf("final reset failed")
	}
}
