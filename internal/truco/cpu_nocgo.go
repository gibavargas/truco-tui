//go:build !cgo

package truco

// DecideCPUActionCpp is the pure-Go fallback used when the binary is built
// without CGO (for example cross-compiled with CGO_ENABLED=0). The C++ AI
// engine is unavailable in that configuration, so the Go heuristic decides.
func DecideCPUActionCpp(g *Game, playerID int) CPUAction {
	return DecideCPUAction(g, playerID)
}

// CPUPersonalityName reports the strategy used for a CPU player. Without the
// C++ engine there is a single Go heuristic.
func CPUPersonalityName(playerID int) string {
	return "Go heuristic"
}
