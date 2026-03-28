package netp2p

import (
	"testing"

	"truco-tui/internal/truco"
)

func TestRotateFailoverSnapshotRemapsLastTrickWinner(t *testing.T) {
	s := truco.Snapshot{
		NumPlayers: 4,
		Players: []truco.Player{
			{ID: 0, Name: "P0", Team: 0, Hand: []truco.Card{{Suit: truco.Clubs, Rank: truco.RA}}},
			{ID: 1, Name: "P1", Team: 1, Hand: []truco.Card{{Suit: truco.Hearts, Rank: truco.R2}}},
			{ID: 2, Name: "P2", Team: 0, Hand: []truco.Card{{Suit: truco.Spades, Rank: truco.R3}}},
			{ID: 3, Name: "P3", Team: 1, Hand: []truco.Card{{Suit: truco.Diamonds, Rank: truco.RK}}},
		},
		CurrentHand: truco.HandState{
			Dealer:         3,
			Turn:           1,
			RoundStart:     2,
			RaiseRequester: 0,
			RoundCards: []truco.PlayedCard{
				{PlayerID: 0, Card: truco.Card{Suit: truco.Clubs, Rank: truco.RA}},
				{PlayerID: 2, Card: truco.Card{Suit: truco.Spades, Rank: truco.R3}},
			},
		},
		TurnPlayer:       1,
		CurrentPlayerIdx: 3,
		LastTrickWinner:  3,
	}

	rot, err := RotateFailoverSnapshot(s, 2)
	if err != nil {
		t.Fatalf("RotateFailoverSnapshot: %v", err)
	}
	if rot.LastTrickWinner != 1 {
		t.Fatalf("unexpected rotated last trick winner: %d", rot.LastTrickWinner)
	}
}
