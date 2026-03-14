package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"truco-tui/internal/appcore"
)

type browserSession struct {
	rt *appcore.Runtime
}

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]*browserSession
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]*browserSession)}
}

func (s *sessionStore) get(id string) *browserSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessions[id]
}

func (s *sessionStore) create() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := randomKey()
	s.sessions[id] = &browserSession{rt: appcore.NewRuntime()}
	return id
}

type apiServer struct {
	store *sessionStore
}

func newAPIServer() *apiServer {
	return &apiServer{store: newSessionStore()}
}

func (srv *apiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errResult("only POST is allowed"))
		return
	}

	action := strings.TrimPrefix(r.URL.Path, "/api/")
	action = strings.TrimSuffix(action, "/")
	if action == "" {
		action = r.URL.Query().Get("action")
	}
	if action == "" {
		writeJSON(w, http.StatusBadRequest, errResult("missing action"))
		return
	}

	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("session_id")
	}

	var body map[string]interface{}
	if r.Body != nil {
		defer r.Body.Close()
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil && err.Error() != "EOF" {
			writeJSON(w, http.StatusBadRequest, errResult("invalid JSON body"))
			return
		}
	}
	if body == nil {
		body = map[string]interface{}{}
	}

	result := srv.dispatch(action, sessionID, body)
	writeJSON(w, http.StatusOK, result)
}

func (srv *apiServer) dispatch(action, sessionID string, body map[string]interface{}) map[string]interface{} {
	if action == "createSession" {
		id := srv.store.create()
		return map[string]interface{}{"ok": true, "sessionId": id}
	}

	bs := srv.store.get(sessionID)
	if bs == nil {
		return errResult("session not found")
	}

	switch action {
	case "setLocale":
		if err := dispatchIntent(bs.rt, "set_locale", appcore.SetLocalePayload{
			Locale: strVal(body, "locale", "pt-BR"),
		}); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, true)

	case "startGame":
		numPlayers := sanitizeNumPlayers(intVal(body, "numPlayers", 2))
		name := strVal(body, "name", "Voce")
		names, cpus := offlinePlayers(name, numPlayers)
		if err := dispatchIntent(bs.rt, "new_offline_game", appcore.NewOfflineGamePayload{
			PlayerNames: names,
			CPUFlags:    cpus,
		}); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, false)

	case "snapshot", "autoCpuLoopTick", "newHand":
		return runtimeResult(bs.rt, false)

	case "play":
		idx := intVal(body, "cardIndex", -1)
		if idx < 0 {
			return errResult("índice da carta ausente")
		}
		if err := dispatchIntent(bs.rt, "game_action", appcore.GameActionPayload{
			Action:    "play",
			CardIndex: idx,
		}); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, false)

	case "truco", "accept", "refuse":
		if err := dispatchIntent(bs.rt, "game_action", appcore.GameActionPayload{
			Action: action,
		}); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, false)

	case "reset", "leaveSession":
		if err := dispatchIntent(bs.rt, "close_session", nil); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, true)

	case "startOnlineHost":
		if err := dispatchIntent(bs.rt, "create_host_session", appcore.CreateHostPayload{
			HostName:   strings.TrimSpace(strVal(body, "name", "Host")),
			NumPlayers: sanitizeNumPlayers(intVal(body, "numPlayers", 2)),
			RelayURL:   strings.TrimSpace(strVal(body, "relay_url", "")),
		}); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, true)

	case "joinOnline":
		if err := dispatchIntent(bs.rt, "join_session", appcore.JoinSessionPayload{
			Key:         strings.TrimSpace(strVal(body, "key", "")),
			PlayerName:  strings.TrimSpace(strVal(body, "name", "Player")),
			DesiredRole: strings.TrimSpace(strVal(body, "role", "auto")),
		}); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, true)

	case "onlineState":
		return runtimeResult(bs.rt, false)

	case "startOnlineMatch":
		if err := dispatchIntent(bs.rt, "start_hosted_match", nil); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, true)

	case "sendChat":
		msg := strings.TrimSpace(strVal(body, "message", ""))
		if msg == "" {
			return errResult("chat message is empty")
		}
		if err := dispatchIntent(bs.rt, "send_chat", appcore.SendChatPayload{Text: msg}); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, true)

	case "sendHostVote":
		if err := dispatchIntent(bs.rt, "vote_host", appcore.HostVotePayload{
			CandidateSeat: intVal(body, "slot", 0),
		}); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, true)

	case "requestReplacementInvite":
		if err := dispatchIntent(bs.rt, "request_replacement_invite", appcore.ReplacementInvitePayload{
			TargetSeat: intVal(body, "slot", 0),
		}); err != nil {
			return errResult(err.Error())
		}
		return runtimeResult(bs.rt, true)

	case "pullOnlineEvents", "pollEvents":
		out := runtimeResult(bs.rt, true)
		out["events"] = drainEvents(bs.rt)
		return out

	default:
		return errResult("unknown action: " + action)
	}
}

func offlinePlayers(name string, numPlayers int) ([]string, []bool) {
	names := make([]string, numPlayers)
	cpus := make([]bool, numPlayers)
	for i := 0; i < numPlayers; i++ {
		if i == 0 {
			names[i] = strings.TrimSpace(name)
			if names[i] == "" {
				names[i] = "Voce"
			}
			continue
		}
		names[i] = "CPU-" + strconv.Itoa(i+1)
		cpus[i] = true
	}
	return names, cpus
}

func dispatchIntent(rt *appcore.Runtime, kind string, payload interface{}) error {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		raw = b
	}
	return rt.DispatchIntent(appcore.AppIntent{Kind: kind, Payload: raw})
}

func runtimeResult(rt *appcore.Runtime, includeEvents bool) map[string]interface{} {
	bundle := rt.SnapshotBundle()
	out := map[string]interface{}{
		"ok":      true,
		"bundle":  bundle,
		"mode":    bundle.Mode,
		"session": sessionFromBundle(bundle),
	}
	if bundle.Match != nil {
		out["snapshot"] = mustJSON(bundle.Match)
	}
	if includeEvents {
		out["events"] = drainEvents(rt)
	}
	return out
}

func sessionFromBundle(bundle appcore.SnapshotBundle) map[string]interface{} {
	if bundle.Lobby == nil {
		return map[string]interface{}{}
	}
	mode := "client"
	if strings.Contains(bundle.Mode, "host") {
		mode = "host"
	}
	connected := make([]bool, bundle.Lobby.NumPlayers)
	for seat, isConnected := range bundle.Lobby.ConnectedSeats {
		if seat >= 0 && seat < len(connected) {
			connected[seat] = isConnected
		}
	}
	return map[string]interface{}{
		"mode":         mode,
		"inviteKey":    bundle.Lobby.InviteKey,
		"numPlayers":   bundle.Lobby.NumPlayers,
		"assignedSeat": bundle.Lobby.AssignedSeat,
		"hostSeat":     bundle.Lobby.HostSeat,
		"slots":        bundle.Lobby.Slots,
		"connected":    connected,
		"started":      bundle.Lobby.Started,
		"role":         bundle.Lobby.Role,
	}
}

func drainEvents(rt *appcore.Runtime) []interface{} {
	events := []interface{}{}
	for {
		ev, ok := rt.PollEvent()
		if !ok {
			break
		}
		events = append(events, normalizeEvent(ev))
	}
	return events
}

func normalizeEvent(ev appcore.AppEvent) interface{} {
	b, err := json.Marshal(ev)
	if err != nil {
		return map[string]interface{}{
			"kind":      "error",
			"sequence":  ev.Sequence,
			"timestamp": ev.Timestamp,
			"payload":   map[string]interface{}{"text": err.Error()},
		}
	}
	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]interface{}{
			"kind":      "error",
			"sequence":  ev.Sequence,
			"timestamp": ev.Timestamp,
			"payload":   map[string]interface{}{"text": err.Error()},
		}
	}
	return out
}

func sanitizeNumPlayers(v int) int {
	if v == 4 {
		return 4
	}
	return 2
}

func errResult(msg string) map[string]interface{} {
	return map[string]interface{}{
		"ok":    false,
		"error": msg,
	}
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func strVal(body map[string]interface{}, key, fallback string) string {
	if v, ok := body[key]; ok {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				return s
			}
		}
	}
	return fallback
}

func intVal(body map[string]interface{}, key string, fallback int) int {
	if v, ok := body[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case string:
			if i, err := strconv.Atoi(n); err == nil {
				return i
			}
		}
	}
	return fallback
}

func randomKey() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return strings.ToUpper(hex.EncodeToString(b[:]))
}

func main() {
	port := os.Getenv("TRUCO_API_PORT")
	if port == "" {
		port = "9090"
	}
	host := os.Getenv("TRUCO_API_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	srv := newAPIServer()

	mux := http.NewServeMux()
	mux.Handle("/api/", srv)

	addr := net.JoinHostPort(host, port)
	log.Printf("Truco HTTP API listening on %s", addr)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	if err := httpSrv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
