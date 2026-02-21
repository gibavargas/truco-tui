//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"truco-tui/internal/truco"
)

type wasmApp struct {
	game       *truco.Game
	localSeat  int
	playerName string
	funcs      []js.Func
}

func main() {
	app := &wasmApp{localSeat: 0, playerName: "Voce"}
	app.registerGlobal()
	select {}
}

func (a *wasmApp) registerGlobal() {
	api := map[string]any{
		"startGame":       js.FuncOf(a.startGame),
		"snapshot":        js.FuncOf(a.snapshot),
		"play":            js.FuncOf(a.play),
		"truco":           js.FuncOf(a.truco),
		"accept":          js.FuncOf(a.accept),
		"refuse":          js.FuncOf(a.refuse),
		"cpuStep":         js.FuncOf(a.cpuStep),
		"autoCpuLoopTick": js.FuncOf(a.autoCpuLoopTick),
		"newHand":         js.FuncOf(a.newHand),
		"reset":           js.FuncOf(a.reset),
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
	return map[string]any{"ok": true}
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
