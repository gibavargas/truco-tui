//go:build js && wasm

package main

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"syscall/js"
	"time"

	"truco-tui/internal/truco"
)

type wasmApp struct {
	game       *truco.Game
	localSeat  int
	playerName string
	online     *onlineState
	funcs      []js.Func
}

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

func main() {
	app := &wasmApp{localSeat: 0, playerName: "Voce"}
	app.registerGlobal()
	select {}
}

func (a *wasmApp) registerGlobal() {
	api := map[string]any{
		"startGame":                js.FuncOf(a.startGame),
		"snapshot":                 js.FuncOf(a.snapshot),
		"play":                     js.FuncOf(a.play),
		"truco":                    js.FuncOf(a.truco),
		"accept":                   js.FuncOf(a.accept),
		"refuse":                   js.FuncOf(a.refuse),
		"cpuStep":                  js.FuncOf(a.cpuStep),
		"autoCpuLoopTick":          js.FuncOf(a.autoCpuLoopTick),
		"newHand":                  js.FuncOf(a.newHand),
		"reset":                    js.FuncOf(a.reset),
		"startOnlineHost":          js.FuncOf(a.startOnlineHost),
		"joinOnline":               js.FuncOf(a.joinOnline),
		"onlineState":              js.FuncOf(a.onlineStateSnapshot),
		"startOnlineMatch":         js.FuncOf(a.startOnlineMatch),
		"sendChat":                 js.FuncOf(a.sendChat),
		"sendHostVote":             js.FuncOf(a.sendHostVote),
		"requestReplacementInvite": js.FuncOf(a.requestReplacementInvite),
		"pullOnlineEvents":         js.FuncOf(a.pullOnlineEvents),
		"leaveSession":             js.FuncOf(a.leaveSession),
	}
	for _, v := range api {
		if fn, ok := v.(js.Func); ok {
			a.funcs = append(a.funcs, fn)
		}
	}
	js.Global().Set("TrucoWasm", js.ValueOf(api))
}

func (a *wasmApp) startGame(_ js.Value, args []js.Value) any {
	numPlayers := 2
	name := "Voce"
	if len(args) > 0 && args[0].Type() == js.TypeObject {
		cfg := args[0]
		if v := cfg.Get("numPlayers"); v.Type() == js.TypeNumber {
			n := v.Int()
			if n == 2 || n == 4 {
				numPlayers = n
			}
		}
		if v := cfg.Get("name"); v.Type() == js.TypeString {
			if s := v.String(); s != "" {
				name = s
			}
		}
	}

	names := make([]string, numPlayers)
	cpus := make([]bool, numPlayers)
	for i := 0; i < numPlayers; i++ {
		if i == 0 {
			names[i] = name
			cpus[i] = false
			continue
		}
		names[i] = fmt.Sprintf("CPU-%d", i+1)
		cpus[i] = true
	}

	g, err := truco.NewGame(names, cpus)
	if err != nil {
		return errResult(err)
	}
	a.game = g
	a.playerName = name
	a.online = nil
	return a.snapshotResult()
}

func (a *wasmApp) snapshot(_ js.Value, _ []js.Value) any {
	if a.game == nil {
		return errResult(fmt.Errorf("partida não iniciada"))
	}
	return a.snapshotResult()
}

func (a *wasmApp) snapshotResult() any {
	snap := a.game.Snapshot(a.localSeat)
	return okSnapshot(snap)
}

func (a *wasmApp) play(_ js.Value, args []js.Value) any {
	if a.game == nil {
		return errResult(fmt.Errorf("partida não iniciada"))
	}
	if len(args) == 0 {
		return errResult(fmt.Errorf("índice da carta ausente"))
	}
	if err := a.game.PlayCard(a.localSeat, args[0].Int()); err != nil {
		return errResult(err)
	}
	return a.snapshotResult()
}

func (a *wasmApp) truco(_ js.Value, _ []js.Value) any {
	if a.game == nil {
		return errResult(fmt.Errorf("partida não iniciada"))
	}
	pending := a.game.PendingTeamToRespond()
	if pending != -1 && a.game.TeamOfPlayer(a.localSeat) == pending {
		if err := a.game.RaiseTruco(a.localSeat); err != nil {
			return errResult(err)
		}
		return a.snapshotResult()
	}
	if err := a.game.AskTruco(a.localSeat); err != nil {
		return errResult(err)
	}
	return a.snapshotResult()
}

func (a *wasmApp) accept(_ js.Value, _ []js.Value) any {
	if a.game == nil {
		return errResult(fmt.Errorf("partida não iniciada"))
	}
	if err := a.game.RespondTruco(a.localSeat, true); err != nil {
		return errResult(err)
	}
	return a.snapshotResult()
}

func (a *wasmApp) refuse(_ js.Value, _ []js.Value) any {
	if a.game == nil {
		return errResult(fmt.Errorf("partida não iniciada"))
	}
	if err := a.game.RespondTruco(a.localSeat, false); err != nil {
		return errResult(err)
	}
	return a.snapshotResult()
}

func (a *wasmApp) cpuStep(_ js.Value, _ []js.Value) any {
	if a.game == nil {
		return errResult(fmt.Errorf("partida não iniciada"))
	}
	isCPU, pid := a.game.IsCPUTurn()
	if !isCPU {
		snap := a.game.Snapshot(a.localSeat)
		return map[string]any{
			"ok":       true,
			"changed":  false,
			"snapshot": mustJSON(snap),
		}
	}
	act := truco.DecideCPUAction(a.game, pid)
	if err := applyCPUAction(a.game, pid, act); err != nil {
		return errResult(err)
	}
	snap := a.game.Snapshot(a.localSeat)
	return map[string]any{
		"ok":       true,
		"changed":  true,
		"snapshot": mustJSON(snap),
	}
}

func (a *wasmApp) autoCpuLoopTick(_ js.Value, _ []js.Value) any {
	if a.game == nil {
		return errResult(fmt.Errorf("partida não iniciada"))
	}
	changed := false
	for i := 0; i < 6; i++ {
		isCPU, pid := a.game.IsCPUTurn()
		if !isCPU {
			break
		}
		act := truco.DecideCPUAction(a.game, pid)
		if err := applyCPUAction(a.game, pid, act); err != nil {
			return errResult(err)
		}
		changed = true
	}
	snap := a.game.Snapshot(a.localSeat)
	return map[string]any{
		"ok":       true,
		"changed":  changed,
		"snapshot": mustJSON(snap),
	}
}

func (a *wasmApp) newHand(_ js.Value, _ []js.Value) any {
	if a.game == nil {
		return errResult(fmt.Errorf("partida não iniciada"))
	}
	a.game.StartNewHand()
	return a.snapshotResult()
}

func (a *wasmApp) reset(_ js.Value, _ []js.Value) any {
	a.game = nil
	a.online = nil
	return map[string]any{"ok": true}
}

func (a *wasmApp) startOnlineHost(_ js.Value, args []js.Value) any {
	cfg := parseConfig(args)
	name := strings.TrimSpace(cfg.name)
	if name == "" {
		name = "Host"
	}
	numPlayers := cfg.numPlayers
	if numPlayers != 2 && numPlayers != 4 {
		numPlayers = 2
	}
	key := randomKey()
	slots := make([]string, numPlayers)
	connected := make([]bool, numPlayers)
	slots[0] = name
	connected[0] = true

	a.online = &onlineState{
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
	a.pushOnlineEvent("session", "info", "Host lobby created.", 0, -1)
	return map[string]any{
		"ok":      true,
		"session": a.online,
	}
}

func (a *wasmApp) joinOnline(_ js.Value, args []js.Value) any {
	cfg := parseConfig(args)
	key := strings.TrimSpace(cfg.key)
	if key == "" {
		return errResult(fmt.Errorf("invite key is required"))
	}
	name := strings.TrimSpace(cfg.name)
	if name == "" {
		name = "Player"
	}
	role := strings.ToLower(strings.TrimSpace(cfg.role))
	if role == "" {
		role = "auto"
	}

	if a.online != nil && a.online.Mode == "host" && a.online.InviteKey == key {
		seat := pickSeatForRole(a.online.Slots, role)
		if seat < 0 {
			return errResult(fmt.Errorf("lobby is full"))
		}
		a.online.Slots[seat] = name
		a.online.Connected[seat] = true
		a.online.Mode = "client"
		a.online.DesiredRole = role
		a.online.AssignedSeat = seat
		a.pushOnlineEvent("join", "info", fmt.Sprintf("%s joined seat %d.", name, seat+1), seat, seat)
		return map[string]any{
			"ok":      true,
			"session": a.online,
		}
	}

	numPlayers := cfg.numPlayers
	if numPlayers != 2 && numPlayers != 4 {
		numPlayers = 2
	}
	slots := make([]string, numPlayers)
	connected := make([]bool, numPlayers)
	slots[0] = "Host"
	connected[0] = true
	seat := pickSeatForRole(slots, role)
	if seat < 0 {
		return errResult(fmt.Errorf("no seat available"))
	}
	slots[seat] = name
	connected[seat] = true

	a.online = &onlineState{
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
	a.pushOnlineEvent("join", "warning", "Joined in local alpha mode (single-runtime session).", seat, seat)
	return map[string]any{
		"ok":      true,
		"session": a.online,
	}
}

func (a *wasmApp) onlineStateSnapshot(_ js.Value, _ []js.Value) any {
	if a.online == nil {
		return map[string]any{"ok": true, "session": nil}
	}
	return map[string]any{"ok": true, "session": a.online}
}

func (a *wasmApp) startOnlineMatch(_ js.Value, _ []js.Value) any {
	if a.online == nil {
		return errResult(fmt.Errorf("online session not initialized"))
	}
	names := make([]string, a.online.NumPlayers)
	cpus := make([]bool, a.online.NumPlayers)
	for i := 0; i < a.online.NumPlayers; i++ {
		slotName := strings.TrimSpace(a.online.Slots[i])
		if slotName == "" {
			slotName = fmt.Sprintf("CPU-%d", i+1)
			cpus[i] = true
		}
		if i == a.online.AssignedSeat {
			cpus[i] = false
		} else if !a.online.Connected[i] {
			cpus[i] = true
		}
		names[i] = slotName
	}
	g, err := truco.NewGame(names, cpus)
	if err != nil {
		return errResult(err)
	}
	a.game = g
	a.localSeat = a.online.AssignedSeat
	a.playerName = names[a.localSeat]
	a.pushOnlineEvent("match_start", "info", "Online alpha match started.", a.localSeat, -1)
	return map[string]any{
		"ok":       true,
		"session":  a.online,
		"snapshot": mustJSON(a.game.Snapshot(a.localSeat)),
	}
}

func (a *wasmApp) sendChat(_ js.Value, args []js.Value) any {
	if a.online == nil {
		return errResult(fmt.Errorf("online session not initialized"))
	}
	if len(args) == 0 {
		return errResult(fmt.Errorf("chat message is required"))
	}
	msg := strings.TrimSpace(args[0].String())
	if msg == "" {
		return errResult(fmt.Errorf("chat message is empty"))
	}
	seat := a.online.AssignedSeat
	name := a.online.Slots[seat]
	a.pushOnlineEvent("chat", "info", fmt.Sprintf("%s: %s", name, msg), seat, -1)
	return map[string]any{"ok": true}
}

func (a *wasmApp) sendHostVote(_ js.Value, args []js.Value) any {
	if a.online == nil {
		return errResult(fmt.Errorf("online session not initialized"))
	}
	if len(args) == 0 {
		return errResult(fmt.Errorf("slot is required"))
	}
	target := args[0].Int() - 1
	if target < 0 || target >= a.online.NumPlayers {
		return errResult(fmt.Errorf("invalid slot"))
	}
	a.pushOnlineEvent(
		"host_vote",
		"info",
		fmt.Sprintf("Seat %d voted host transfer to seat %d.", a.online.AssignedSeat+1, target+1),
		a.online.AssignedSeat,
		target,
	)
	return map[string]any{"ok": true}
}

func (a *wasmApp) requestReplacementInvite(_ js.Value, args []js.Value) any {
	if a.online == nil {
		return errResult(fmt.Errorf("online session not initialized"))
	}
	if len(args) == 0 {
		return errResult(fmt.Errorf("slot is required"))
	}
	target := args[0].Int() - 1
	if target < 0 || target >= a.online.NumPlayers {
		return errResult(fmt.Errorf("invalid slot"))
	}
	invite := "REPL-" + randomKey()
	a.pushOnlineEvent(
		"replacement_invite",
		"warning",
		fmt.Sprintf("Replacement invite for seat %d: %s", target+1, invite),
		a.online.AssignedSeat,
		target,
	)
	return map[string]any{"ok": true, "inviteKey": invite}
}

func (a *wasmApp) pullOnlineEvents(_ js.Value, _ []js.Value) any {
	if a.online == nil {
		return map[string]any{"ok": true, "events": []onlineEvent{}}
	}
	out := append([]onlineEvent(nil), a.online.Events...)
	a.online.Events = a.online.Events[:0]
	return map[string]any{"ok": true, "events": out}
}

func (a *wasmApp) leaveSession(_ js.Value, _ []js.Value) any {
	a.game = nil
	a.online = nil
	return map[string]any{"ok": true}
}

func (a *wasmApp) pushOnlineEvent(evType, severity, msg string, actorSeat, targetSeat int) {
	if a.online == nil {
		return
	}
	a.online.Events = append(a.online.Events, onlineEvent{
		Type:       evType,
		Severity:   severity,
		Message:    msg,
		ActorSeat:  actorSeat,
		TargetSeat: targetSeat,
		Timestamp:  time.Now().UnixMilli(),
	})
}

func pickSeatForRole(slots []string, role string) int {
	candidate := []int{}
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

type appConfig struct {
	numPlayers int
	name       string
	key        string
	role       string
}

func parseConfig(args []js.Value) appConfig {
	cfg := appConfig{numPlayers: 2}
	if len(args) == 0 || args[0].Type() != js.TypeObject {
		return cfg
	}
	obj := args[0]
	if v := obj.Get("numPlayers"); v.Type() == js.TypeNumber {
		cfg.numPlayers = v.Int()
	}
	if v := obj.Get("name"); v.Type() == js.TypeString {
		cfg.name = v.String()
	}
	if v := obj.Get("key"); v.Type() == js.TypeString {
		cfg.key = v.String()
	}
	if v := obj.Get("role"); v.Type() == js.TypeString {
		cfg.role = v.String()
	}
	return cfg
}

func randomKey() string {
	var b [6]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		return fmt.Sprintf("K-%d", time.Now().UnixNano())
	}
	return strings.ToUpper(hex.EncodeToString(b[:]))
}

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
		return fmt.Errorf("ação CPU desconhecida: %s", a.Kind)
	}
}

func okSnapshot(s truco.Snapshot) map[string]any {
	return map[string]any{
		"ok":       true,
		"snapshot": mustJSON(s),
	}
}

func errResult(err error) map[string]any {
	return map[string]any{
		"ok":    false,
		"error": err.Error(),
	}
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
