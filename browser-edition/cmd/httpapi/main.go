package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

func (s *sessionStore) delete(id string) bool {
	s.mu.Lock()
	bs, ok := s.sessions[id]
	if ok {
		delete(s.sessions, id)
	}
	s.mu.Unlock()
	if ok && bs != nil && bs.rt != nil {
		_ = bs.rt.Close()
	}
	return ok
}

type apiServer struct {
	store *sessionStore
}

func newAPIServer() *apiServer {
	return &apiServer{store: newSessionStore()}
}

func (srv *apiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Security headers
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "1; mode=block")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errResult("method_not_allowed", "only POST is allowed"))
		return
	}

	action := strings.TrimPrefix(r.URL.Path, "/api/")
	action = strings.TrimSuffix(action, "/")
	if action == "" {
		action = r.URL.Query().Get("action")
	}
	if action == "" {
		writeJSON(w, http.StatusBadRequest, errResult("missing_action", "missing action"))
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
			writeJSON(w, http.StatusBadRequest, errResult("invalid_json_body", "invalid JSON body"))
			return
		}
	}
	if body == nil {
		body = map[string]interface{}{}
	}

	status, result := srv.dispatch(action, sessionID, body)
	writeJSON(w, status, result)
}

func (srv *apiServer) dispatch(action, sessionID string, body map[string]interface{}) (int, map[string]interface{}) {
	if action == "createSession" {
		id := srv.store.create()
		return http.StatusOK, map[string]interface{}{"ok": true, "sessionId": id}
	}

	bs := srv.store.get(sessionID)
	if bs == nil {
		return http.StatusNotFound, errResult("session_not_found", "session not found")
	}

	switch action {
	case "setLocale":
		if err := dispatchIntent(bs.rt, appcore.IntentSetLocale, appcore.SetLocalePayload{
			Locale: strVal(body, "locale", appcore.LocalePTBR),
		}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "set_locale_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, true)

	case "startGame":
		numPlayers := sanitizeNumPlayers(intVal(body, "numPlayers", 2))
		name := strVal(body, "name", "Voce")
		names, cpus := offlinePlayers(name, numPlayers)
		if err := dispatchIntent(bs.rt, appcore.IntentNewOfflineGame, appcore.NewOfflineGamePayload{
			PlayerNames: names,
			CPUFlags:    cpus,
		}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "start_game_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, false)

	case "snapshot":
		return http.StatusOK, runtimeResult(bs.rt, false)

	case "autoCpuLoopTick":
		if err := dispatchIntent(bs.rt, appcore.IntentTick, appcore.TickPayload{MaxSteps: 12}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "tick_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, false)

	case "newHand":
		if err := dispatchIntent(bs.rt, appcore.IntentNewHand, nil); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "new_hand_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, false)

	case "play":
		idx := intVal(body, "cardIndex", -1)
		if idx < 0 {
			return http.StatusBadRequest, errResult("missing_card_index", "índice da carta ausente")
		}
		if err := dispatchIntent(bs.rt, appcore.IntentGameAction, appcore.GameActionPayload{
			Action:    "play",
			CardIndex: idx,
			FaceDown:  boolVal(body, "faceDown", false),
		}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "game_action_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, false)

	case "truco", "accept", "refuse":
		if err := dispatchIntent(bs.rt, appcore.IntentGameAction, appcore.GameActionPayload{
			Action: action,
		}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "game_action_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, false)

	case "reset":
		if err := dispatchIntent(bs.rt, appcore.IntentCloseSession, nil); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "close_session_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, true)

	case "leaveSession", "closeSession":
		if !srv.store.delete(sessionID) {
			return http.StatusNotFound, errResult("session_not_found", "session not found")
		}
		return http.StatusOK, map[string]interface{}{"ok": true, "sessionClosed": true}

	case "startOnlineHost":
		if err := dispatchIntent(bs.rt, appcore.IntentCreateHostSession, appcore.CreateHostPayload{
			HostName:   strings.TrimSpace(strVal(body, "name", "Host")),
			NumPlayers: sanitizeNumPlayers(intVal(body, "numPlayers", 2)),
			RelayURL:   strings.TrimSpace(strVal(body, "relay_url", "")),
		}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "create_host_session_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, true)

	case "joinOnline":
		if err := dispatchIntent(bs.rt, appcore.IntentJoinSession, appcore.JoinSessionPayload{
			Key:         strings.TrimSpace(strVal(body, "key", "")),
			PlayerName:  strings.TrimSpace(strVal(body, "name", "Player")),
			DesiredRole: strings.TrimSpace(strVal(body, "role", appcore.DesiredRoleAuto)),
		}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "join_session_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, true)

	case "startOnlineMatch":
		if err := dispatchIntent(bs.rt, appcore.IntentStartHostedMatch, nil); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "start_hosted_match_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, true)

	case "sendChat":
		msg := strings.TrimSpace(strVal(body, "message", ""))
		if msg == "" {
			return http.StatusBadRequest, errResult("empty_chat_message", "chat message is empty")
		}
		if err := dispatchIntent(bs.rt, appcore.IntentSendChat, appcore.SendChatPayload{Text: msg}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "send_chat_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, true)

	case "sendHostVote":
		if err := dispatchIntent(bs.rt, appcore.IntentVoteHost, appcore.HostVotePayload{
			CandidateSeat: intVal(body, "slot", 0),
		}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "vote_host_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, true)

	case "requestReplacementInvite":
		if err := dispatchIntent(bs.rt, appcore.IntentRequestReplacementInvite, appcore.ReplacementInvitePayload{
			TargetSeat: intVal(body, "slot", 0),
		}); err != nil {
			return http.StatusUnprocessableEntity, runtimeErrResult(bs.rt, "request_replacement_invite_failed", err)
		}
		return http.StatusOK, runtimeResult(bs.rt, true)

	case "pullOnlineEvents", "pollEvents":
		return http.StatusOK, runtimeResult(bs.rt, true)

	default:
		return http.StatusBadRequest, errResult("unknown_action", "unknown action: "+action)
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
	normalizedBundle := normalizeBundleJSON(bundle)
	out := map[string]interface{}{
		"ok":      true,
		"bundle":  normalizedBundle,
		"mode":    bundle.Mode,
		"session": sessionFromBundle(bundle),
	}
	if matchSnapshot, ok := normalizedBundle["match"]; ok && matchSnapshot != nil {
		out["snapshot"] = mustJSON(matchSnapshot)
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
		"network":      bundle.Connection.Network,
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
			"kind":      appcore.EventError,
			"sequence":  ev.Sequence,
			"timestamp": ev.Timestamp,
			"payload":   map[string]interface{}{"text": err.Error()},
		}
	}
	var out interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]interface{}{
			"kind":      appcore.EventError,
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

func runtimeErrResult(rt *appcore.Runtime, fallbackCode string, err error) map[string]interface{} {
	bundle := rt.SnapshotBundle()
	normalizedBundle := normalizeBundleJSON(bundle)
	code := fallbackCode
	message := err.Error()
	if bundle.Connection.LastError != nil {
		if strings.TrimSpace(bundle.Connection.LastError.Code) != "" {
			code = bundle.Connection.LastError.Code
		}
		if strings.TrimSpace(bundle.Connection.LastError.Message) != "" {
			message = bundle.Connection.LastError.Message
		}
	}
	out := errResult(code, message)
	out["bundle"] = normalizedBundle
	out["mode"] = bundle.Mode
	out["session"] = sessionFromBundle(bundle)
	return out
}

func errResult(code, msg string) map[string]interface{} {
	return map[string]interface{}{
		"ok":         false,
		"error":      msg,
		"error_code": code,
	}
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func normalizeBundleJSON(bundle appcore.SnapshotBundle) map[string]interface{} {
	out := mustJSONObject(bundle)
	ensureArrayPath(out, "ui", "lobby_slots")
	ensureArrayPath(out, "diagnostics", "event_log")
	if _, ok := out["lobby"]; ok {
		ensureArrayPath(out, "lobby", "slots")
	}
	if _, ok := out["match"]; ok {
		ensureArrayPath(out, "match", "Players")
		ensureArrayPath(out, "match", "LastTrickCards")
		ensureArrayPath(out, "match", "TrickPiles")
		ensureArrayPath(out, "match", "Logs")
		ensureArrayPath(out, "match", "CurrentHand", "RoundCards")
		ensureArrayPath(out, "match", "CurrentHand", "TrickResults")
	}
	return out
}

func mustJSONObject(v interface{}) map[string]interface{} {
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]interface{}{}
	}
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]interface{}{}
	}
	return out
}

func ensureArrayPath(root map[string]interface{}, path ...string) {
	if len(path) == 0 {
		return
	}
	current := root
	for i, key := range path {
		if i == len(path)-1 {
			if raw, ok := current[key]; !ok || raw == nil {
				current[key] = []interface{}{}
			}
			return
		}
		next, ok := current[key].(map[string]interface{})
		if !ok {
			return
		}
		current = next
	}
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

func boolVal(body map[string]interface{}, key string, fallback bool) bool {
	if v, ok := body[key]; ok {
		switch b := v.(type) {
		case bool:
			return b
		case string:
			switch strings.ToLower(strings.TrimSpace(b)) {
			case "1", "true", "yes", "on":
				return true
			case "0", "false", "no", "off":
				return false
			}
		case float64:
			return b != 0
		}
	}
	return fallback
}

func randomKey() string {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("crypto/rand failed to provide entropy")
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
	staticRoot := resolveStaticRoot()

	mux := http.NewServeMux()
	mux.Handle("/api/", srv)
	mux.Handle("/", newStaticHandler(staticRoot))

	addr := net.JoinHostPort(host, port)
	log.Printf("Truco HTTP API listening on %s", addr)
	if staticRoot != "" {
		log.Printf("Serving browser assets from %s", staticRoot)
	}
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

func resolveStaticRoot() string {
	if envRoot := strings.TrimSpace(os.Getenv("TRUCO_WEB_ROOT")); envRoot != "" {
		if info, err := os.Stat(envRoot); err == nil && info.IsDir() {
			return envRoot
		}
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		if hasIndexHTML(exeDir) {
			return exeDir
		}
	}

	for _, candidate := range []string{
		filepath.Join("browser-edition", "dist"),
		"dist",
	} {
		if hasIndexHTML(candidate) {
			return candidate
		}
	}

	return ""
}

func hasIndexHTML(root string) bool {
	info, err := os.Stat(filepath.Join(root, "index.html"))
	return err == nil && !info.IsDir()
}

func newStaticHandler(root string) http.Handler {
	if root == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		})
	}

	fileServer := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(root, "index.html"))
			return
		}

		assetPath := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")))
		if info, err := os.Stat(assetPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(root, "index.html"))
	})
}
