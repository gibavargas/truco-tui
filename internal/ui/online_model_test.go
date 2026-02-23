package ui

import (
	"testing"

	"truco-tui/internal/netp2p"
)

func TestSelectFailoverSeatSkipsSeatZero(t *testing.T) {
	fs := netp2p.ClientFailoverState{
		HostSeat:    0,
		NumPlayers:  4,
		Slots:       []string{"Host", "A", "B", "C"},
		PeerHosts:   map[int]string{0: "10.0.0.1", 1: "10.0.0.2", 2: "10.0.0.3", 3: "10.0.0.4"},
		HandoffPort: 39001,
	}
	if got := selectFailoverSeat(fs); got != 1 {
		t.Fatalf("selectFailoverSeat() = %d, want 1 when host seat is 0", got)
	}
}

func TestSelectFailoverSeatPrefersLiveHostSeat(t *testing.T) {
	fs := netp2p.ClientFailoverState{
		HostSeat:    2,
		NumPlayers:  4,
		Slots:       []string{"Host", "A", "B", "C"},
		PeerHosts:   map[int]string{1: "10.0.0.2", 2: "10.0.0.3", 3: "10.0.0.4"},
		HandoffPort: 39001,
	}
	if got := selectFailoverSeat(fs); got != 2 {
		t.Fatalf("selectFailoverSeat() = %d, want 2", got)
	}
}
