package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"truco-tui/internal/truco"
)

// ---------------------------------------------------------------------------
// Session types
// ---------------------------------------------------------------------------

type onlineEvent struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	ActorSeat  int    `json:"actorSeat,omitempty"`
	TargetSeat int    `json:"targetSeat,omitempty"`
	Timestamp  int64  `json:"timestamp"`
}

type onlineState struct {
	Mode         string        `json:"mode"`
	InviteKey    string        `json:"inviteKey"`
	NumPlayers   int           `json:"numPlayers"`
	DesiredRole  string        `json:"desiredRole"`
	AssignedSeat int           `json:"assignedSeat"`
	HostSeat     int           `json:"hostSeat"`
	Slots        []string      `json:"slots"`
	Connected    []bool        `json:"connected"`
	Events       []onlineEvent `json:"events"`
}

type gameSession struct {
	mu         sync.Mutex
	game       *truco.Game
	localSeat  int
	playerName string
	online     *onlineState
}

// ---------------------------------------------------------------------------
// Session store
// ---------------------------------------------------------------------------

type sessionStore struct {
	mu       sync.Mutex
	sessions map[string]*gameSession
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]*gameSession)}
}

func (s *sessionStore) get(id string) *gameSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessions[id]
}

func (s *sessionStore) set(id string, gs *gameSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = gs
}


// ---------------------------------------------------------------------------
// API server
// ---------------------------------------------------------------------------

type apiServer struct {
	store *sessionStore
}

func newAPIServer() *apiServer {
	return &apiServer{store: newSessionStore()}
}

func (srv *apiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

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
	var gs *gameSession
	if action != "createSession" {
		gs = srv.store.get(sessionID)
		if gs == nil {
			return errResult("session not found")
		}
		gs.mu.Lock()
		defer gs.mu.Unlock()
	}

	switch action {

	// ── session management ──────────────────────────────────────────────

	case "createSession":
		id := randomKey()
		srv.store.set(id, &gameSession{localSeat: 0, playerName: "Voce"})
		return map[string]interface{}{"ok": true, "sessionId": id}

	case "startGame":
		numPlayers := intVal(body, "numPlayers", 2)
		if numPlayers != 2 && numPlayers != 4 {
			numPlayers = 2
		}
		name := strVal(body, "name", "Voce")
		names := make([]string, numPlayers)
		cpus := make([]bool, numPlayers)
		for i := 0; i < numPlayers; i++ {
			if i == 0 {
				names[i] = name
				cpus[i] = false
			} else {
				names[i] = fmt.Sprintf("CPU-%d", i+1)
				cpus[i] = true
			}
		}
		g, err := truco.NewGame(names, cpus)
		if err != nil {
			return errResult(err.Error())
		}
		gs.game = g
		gs.playerName = name
		gs.online = nil
		return snapshotResult(gs)

	case "snapshot":
		if gs.game == nil {
			return errResult("partida não iniciada")
		}
		return snapshotResult(gs)

	case "play":
		if gs.game == nil {
			return errResult("partida não iniciada")
		}
		idx := intVal(body, "cardIndex", -1)
		if idx < 0 {
			return errResult("índice da carta ausente")
		}
		if err := gs.game.PlayCard(gs.localSeat, idx); err != nil {
			return errResult(err.Error())
		}
		return snapshotResult(gs)

	case "truco":
		if gs.game == nil {
			return errResult("partida não iniciada")
		}
		pending := gs.game.PendingTeamToRespond()
		if pending != -1 && gs.game.TeamOfPlayer(gs.localSeat) == pending {
			if err := gs.game.RaiseTruco(gs.localSeat); err != nil {
				return errResult(err.Error())
			}
			return snapshotResult(gs)
		}
		if err := gs.game.AskTruco(gs.localSeat); err != nil {
			return errResult(err.Error())
		}
		return snapshotResult(gs)

	case "accept":
		if gs.game == nil {
			return errResult("partida não iniciada")
		}
		if err := gs.game.RespondTruco(gs.localSeat, true); err != nil {
			return errResult(err.Error())
		}
		return snapshotResult(gs)

	case "refuse":
		if gs.game == nil {
			return errResult("partida não iniciada")
		}
		if err := gs.game.RespondTruco(gs.localSeat, false); err != nil {
			return errResult(err.Error())
		}
		return snapshotResult(gs)

	case "autoCpuLoopTick":
		if gs.game == nil {
			return errResult("partida não iniciada")
		}
		changed := false
		for i := 0; i < 6; i++ {
			isCPU, pid := gs.game.IsCPUTurn()
			if !isCPU {
				break
			}
			act := truco.DecideCPUAction(gs.game, pid)
			if err := applyCPUAction(gs.game, pid, act); err != nil {
				return errResult(err.Error())
			}
			changed = true
		}
		snap := gs.game.Snapshot(gs.localSeat)
		return map[string]interface{}{
			"ok":       true,
			"changed":  changed,
			"snapshot": mustJSON(snap),
		}

	case "newHand":
		if gs.game == nil {
			return errResult("partida não iniciada")
		}
		gs.game.StartNewHand()
		return snapshotResult(gs)

	case "reset":
		gs.game = nil
		gs.online = nil
		return map[string]interface{}{"ok": true}

	// ── online lobby ────────────────────────────────────────────────────

	case "startOnlineHost":
		name := strings.TrimSpace(strVal(body, "name", "Host"))
		if name == "" {
			name = "Host"
		}
		numPlayers := intVal(body, "numPlayers", 2)
		if numPlayers != 2 && numPlayers != 4 {
			numPlayers = 2
		}
		key := randomKey()
		slots := make([]string, numPlayers)
		connected := make([]bool, numPlayers)
		slots[0] = name
		connected[0] = true

		gs.online = &onlineState{
			Mode:         "host",
			InviteKey:    key,
			NumPlayers:   numPlayers,
			DesiredRole:  "auto",
			AssignedSeat: 0,
			HostSeat:     0,
			Slots:        slots,
			Connected:    connected,
			Events:       []onlineEvent{},
		}
		pushOnlineEvent(gs, "session", "info", "Host lobby created.", 0, -1)
		return map[string]interface{}{
			"ok":      true,
			"session": gs.online,
		}

	case "joinOnline":
		key := strings.TrimSpace(strVal(body, "key", ""))
		if key == "" {
			return errResult("invite key is required")
		}
		name := strings.TrimSpace(strVal(body, "name", "Player"))
		if name == "" {
			name = "Player"
		}
		role := strings.ToLower(strings.TrimSpace(strVal(body, "role", "auto")))
		if role == "" {
			role = "auto"
		}

		if gs.online != nil && gs.online.Mode == "host" && gs.online.InviteKey == key {
			seat := pickSeatForRole(gs.online.Slots, role)
			if seat < 0 {
				return errResult("lobby is full")
			}
			gs.online.Slots[seat] = name
			gs.online.Connected[seat] = true
			gs.online.Mode = "client"
			gs.online.DesiredRole = role
			gs.online.AssignedSeat = seat
			pushOnlineEvent(gs, "join", "info", fmt.Sprintf("%s joined seat %d.", name, seat+1), seat, seat)
			return map[string]interface{}{
				"ok":      true,
				"session": gs.online,
			}
		}

		numPlayers := intVal(body, "numPlayers", 2)
		if numPlayers != 2 && numPlayers != 4 {
			numPlayers = 2
		}
		slots := make([]string, numPlayers)
		connected := make([]bool, numPlayers)
		slots[0] = "Host"
		connected[0] = true
		seat := pickSeatForRole(slots, role)
		if seat < 0 {
			return errResult("no seat available")
		}
		slots[seat] = name
		connected[seat] = true

		gs.online = &onlineState{
			Mode:         "client",
			InviteKey:    key,
			NumPlayers:   numPlayers,
			DesiredRole:  role,
			AssignedSeat: seat,
			HostSeat:     0,
			Slots:        slots,
			Connected:    connected,
			Events:       []onlineEvent{},
		}
		pushOnlineEvent(gs, "join", "warning", "Joined in local alpha mode (single-runtime session).", seat, seat)
		return map[string]interface{}{
			"ok":      true,
			"session": gs.online,
		}

	case "onlineState":
		if gs.online == nil {
			return map[string]interface{}{"ok": true, "session": nil}
		}
		return map[string]interface{}{"ok": true, "session": gs.online}

	case "startOnlineMatch":
		if gs.online == nil {
			return errResult("online session not initialized")
		}
		names := make([]string, gs.online.NumPlayers)
		cpus := make([]bool, gs.online.NumPlayers)
		for i := 0; i < gs.online.NumPlayers; i++ {
			slotName := strings.TrimSpace(gs.online.Slots[i])
			if slotName == "" {
				slotName = fmt.Sprintf("CPU-%d", i+1)
				cpus[i] = true
			}
			if i == gs.online.AssignedSeat {
				cpus[i] = false
			} else if !gs.online.Connected[i] {
				cpus[i] = true
			}
			names[i] = slotName
		}
		g, err := truco.NewGame(names, cpus)
		if err != nil {
			return errResult(err.Error())
		}
		gs.game = g
		gs.localSeat = gs.online.AssignedSeat
		gs.playerName = names[gs.localSeat]
		pushOnlineEvent(gs, "match_start", "info", "Online alpha match started.", gs.localSeat, -1)
		return map[string]interface{}{
			"ok":       true,
			"session":  gs.online,
			"snapshot": mustJSON(gs.game.Snapshot(gs.localSeat)),
		}

	case "sendChat":
		if gs.online == nil {
			return errResult("online session not initialized")
		}
		msg := strings.TrimSpace(strVal(body, "message", ""))
		if msg == "" {
			return errResult("chat message is empty")
		}
		seat := gs.online.AssignedSeat
		name := gs.online.Slots[seat]
		pushOnlineEvent(gs, "chat", "info", fmt.Sprintf("%s: %s", name, msg), seat, -1)
		return map[string]interface{}{"ok": true}

	case "sendHostVote":
		if gs.online == nil {
			return errResult("online session not initialized")
		}
		target := intVal(body, "slot", 0) - 1
		if target < 0 || target >= gs.online.NumPlayers {
			return errResult("invalid slot")
		}
		pushOnlineEvent(gs, "host_vote", "info",
			fmt.Sprintf("Seat %d voted host transfer to seat %d.", gs.online.AssignedSeat+1, target+1),
			gs.online.AssignedSeat, target)
		return map[string]interface{}{"ok": true}

	case "requestReplacementInvite":
		if gs.online == nil {
			return errResult("online session not initialized")
		}
		target := intVal(body, "slot", 0) - 1
		if target < 0 || target >= gs.online.NumPlayers {
			return errResult("invalid slot")
		}
		invite := "REPL-" + randomKey()
		pushOnlineEvent(gs, "replacement_invite", "warning",
			fmt.Sprintf("Replacement invite for seat %d: %s", target+1, invite),
			gs.online.AssignedSeat, target)
		return map[string]interface{}{"ok": true, "inviteKey": invite}

	case "pullOnlineEvents":
		if gs.online == nil {
			return map[string]interface{}{"ok": true, "events": []onlineEvent{}}
		}
		out := append([]onlineEvent(nil), gs.online.Events...)
		gs.online.Events = gs.online.Events[:0]
		return map[string]interface{}{"ok": true, "events": out}

	case "leaveSession":
		gs.game = nil
		gs.online = nil
		return map[string]interface{}{"ok": true}

	default:
		return errResult(fmt.Sprintf("unknown action: %s", action))
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func applyCPUAction(g *truco.Game, pid int, a truco.CPUAction) error {
	switch a.Kind {
	case "ask_truco":
		return g.AskTruco(pid)
	case "raise":
		return g.RaiseTruco(pid)
	case "accept":
		return g.RespondTruco(pid, true)
	case "refuse":
		return g.RespondTruco(pid, false)
	case "play":
		return g.PlayCard(pid, a.CardIndex)
	default:
		return fmt.Errorf("unknown CPU action: %s", a.Kind)
	}
}

func snapshotResult(gs *gameSession) map[string]interface{} {
	snap := gs.game.Snapshot(gs.localSeat)
	return map[string]interface{}{
		"ok":       true,
		"snapshot": mustJSON(snap),
	}
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
		if s, ok := v.(string); ok && s != "" {
			return s
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
		return fmt.Sprintf("K-%d", time.Now().UnixNano())
	}
	return strings.ToUpper(hex.EncodeToString(b[:]))
}

func pushOnlineEvent(gs *gameSession, evType, severity, msg string, actorSeat, targetSeat int) {
	if gs.online == nil {
		return
	}
	gs.online.Events = append(gs.online.Events, onlineEvent{
		Type:       evType,
		Severity:   severity,
		Message:    msg,
		ActorSeat:  actorSeat,
		TargetSeat: targetSeat,
		Timestamp:  time.Now().UnixMilli(),
	})
}

func pickSeatForRole(slots []string, role string) int {
	var candidate []int
	if len(slots) == 2 {
		candidate = []int{1}
	} else {
		switch role {
		case "partner":
			candidate = []int{2, 1, 3}
		case "opponent":
			candidate = []int{1, 3, 2}
		default:
			candidate = []int{1, 2, 3}
		}
	}
	for _, seat := range candidate {
		if seat >= 0 && seat < len(slots) && strings.TrimSpace(slots[seat]) == "" {
			return seat
		}
	}
	for seat := range slots {
		if strings.TrimSpace(slots[seat]) == "" {
			return seat
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

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
