package truco

import "testing"

func TestDecideCPUActionDoesNotLoopOnIllegalTrucoRaise(t *testing.T) {
	g, err := NewGame([]string{"CPU-1", "CPU-2"}, []bool{true, true})
	if err != nil {
		t.Fatalf("NewGame() error: %v", err)
	}

	g.mu.Lock()
	// Força um estado em que o jogador 0 está na vez, mas o time dele já
	// foi o último a aumentar a aposta: pedir truco seria inválido.
	g.hand.Turn = 0
	g.hand.PendingRaiseFor = -1
	g.hand.Stake = 3
	g.hand.TrucoByTeam = g.players[0].Team
	g.hand.Manilha = R4
	g.players[0].Hand = []Card{
		{Rank: R3, Suit: Clubs},
		{Rank: R2, Suit: Hearts},
		{Rank: RA, Suit: Spades},
	}
	g.mu.Unlock()

	act := DecideCPUAction(g, 0)
	if act.Kind == "ask_truco" {
		t.Fatalf("CPU should not ask truco when own team raised last; action=%+v", act)
	}
	if act.Kind != "play" {
		t.Fatalf("expected fallback play action, got %+v", act)
	}
}

func TestCanAskTrucoByPlayerMatchesGameRules(t *testing.T) {
	g, err := NewGame([]string{"P1", "P2"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame() error: %v", err)
	}

	turn := g.Snapshot(0).TurnPlayer
	if !g.CanAskTrucoByPlayer(turn) {
		t.Fatalf("current turn player should be able to ask truco in initial state")
	}
	other := 1 - turn
	if g.CanAskTrucoByPlayer(other) {
		t.Fatalf("non-turn player should not be able to ask truco")
	}
}

func TestDecideCPUActionUsesLowestWinningCard(t *testing.T) {
	g, err := NewGame([]string{"CPU-1", "CPU-2"}, []bool{true, true})
	if err != nil {
		t.Fatalf("NewGame() error: %v", err)
	}

	g.mu.Lock()
	g.hand.Manilha = R4
	g.hand.Stake = 12 // bloqueia pedido de truco para focar na decisão de carta
	g.hand.Turn = 0
	g.hand.RoundCards = []PlayedCard{
		{PlayerID: 1, Card: Card{Rank: RQ, Suit: Clubs}},
	}
	g.players[0].Hand = []Card{
		{Rank: R5, Suit: Hearts},
		{Rank: R7, Suit: Spades},
		{Rank: R3, Suit: Diamonds},
	}
	g.mu.Unlock()

	oppPower := CardPower(Card{Rank: RQ, Suit: Clubs}, R4)
	expected := -1
	expectedPower := 1000
	for i, c := range g.HandCards(0) {
		p := CardPower(c, R4)
		if p > oppPower && p < expectedPower {
			expected = i
			expectedPower = p
		}
	}
	if expected < 0 {
		t.Fatalf("test setup invalid: expected at least one winning card")
	}

	act := DecideCPUAction(g, 0)
	if act.Kind != "play" {
		t.Fatalf("expected play action, got %+v", act)
	}
	if act.CardIndex != expected {
		t.Fatalf("expected lowest winning index %d, got %d", expected, act.CardIndex)
	}
}

func TestDecideCPUActionDiscardsWeakestWhenTeamLeading(t *testing.T) {
	g, err := NewGame([]string{"P1", "P2", "P3", "P4"}, []bool{true, true, true, true})
	if err != nil {
		t.Fatalf("NewGame() error: %v", err)
	}

	g.mu.Lock()
	g.hand.Manilha = R4
	g.hand.Stake = 12
	g.hand.Turn = 0
	g.hand.RoundCards = []PlayedCard{
		{PlayerID: 1, Card: Card{Rank: RQ, Suit: Clubs}},    // adversário
		{PlayerID: 2, Card: Card{Rank: RA, Suit: Diamonds}}, // parceiro lidera
	}
	g.players[0].Hand = []Card{
		{Rank: R2, Suit: Spades},
		{Rank: R5, Suit: Hearts},
		{Rank: RK, Suit: Clubs},
	}
	g.mu.Unlock()

	hand := g.HandCards(0)
	expected := 0
	expectedPower := CardPower(hand[0], R4)
	for i := 1; i < len(hand); i++ {
		p := CardPower(hand[i], R4)
		if p < expectedPower {
			expected = i
			expectedPower = p
		}
	}

	act := DecideCPUAction(g, 0)
	if act.Kind != "play" {
		t.Fatalf("expected play action, got %+v", act)
	}
	if act.CardIndex != expected {
		t.Fatalf("expected weakest card index %d, got %d", expected, act.CardIndex)
	}
}

func TestDecideCPUActionCanRaiseOnPendingResponse(t *testing.T) {
	g, err := NewGame([]string{"CPU-1", "CPU-2"}, []bool{true, true})
	if err != nil {
		t.Fatalf("NewGame() error: %v", err)
	}

	asker := g.Snapshot(0).TurnPlayer
	responder := 1 - asker
	if err := g.AskTruco(asker); err != nil {
		t.Fatalf("AskTruco: %v", err)
	}

	g.mu.Lock()
	g.hand.Turn = responder
	g.hand.Manilha = R4
	g.players[responder].Hand = []Card{
		{Rank: R4, Suit: Clubs},
		{Rank: R4, Suit: Hearts},
		{Rank: R3, Suit: Spades},
	}
	g.mu.Unlock()

	act := DecideCPUAction(g, responder)
	if act.Kind != "raise" {
		t.Fatalf("expected raise action, got %+v", act)
	}
}
