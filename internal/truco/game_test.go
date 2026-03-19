package truco

import "testing"

func TestStakeProgressionByTrucoResponses(t *testing.T) {
	names := []string{"p1", "p2"}
	cpus := []bool{false, false}
	g, err := NewGame(names, cpus)
	if err != nil {
		t.Fatalf("NewGame error: %v", err)
	}

	firstTurn := g.Snapshot(0).TurnPlayer
	if err := g.AskTruco(firstTurn); err != nil {
		t.Fatalf("AskTruco #1: %v", err)
	}
	if err := g.RespondTruco(1-firstTurn, true); err != nil {
		t.Fatalf("RespondTruco #1: %v", err)
	}
	if got := g.Snapshot(0).CurrentHand.Stake; got != 3 {
		t.Fatalf("stake = %d, want 3", got)
	}

	// O próximo aumento deve vir do time oposto ao último que pediu.
	secondTurn := 1 - firstTurn
	g.mu.Lock()
	g.hand.Turn = secondTurn
	g.mu.Unlock()
	if err := g.AskTruco(secondTurn); err != nil {
		t.Fatalf("AskTruco #2: %v", err)
	}
	if err := g.RespondTruco(firstTurn, true); err != nil {
		t.Fatalf("RespondTruco #2: %v", err)
	}
	if got := g.Snapshot(0).CurrentHand.Stake; got != 6 {
		t.Fatalf("stake = %d, want 6", got)
	}
}

func TestAskTrucoMovesTurnToOpponentAndAcceptRestoresRequester(t *testing.T) {
	names := []string{"p1", "p2", "p3", "p4"}
	cpus := []bool{false, false, false, false}
	g, err := NewGame(names, cpus)
	if err != nil {
		t.Fatalf("NewGame error: %v", err)
	}

	// Força o turno para o jogador 4 (CPU-4 em cenários reais) para validar regressão.
	g.mu.Lock()
	g.hand.Turn = 3
	g.mu.Unlock()

	if err := g.AskTruco(3); err != nil {
		t.Fatalf("AskTruco: %v", err)
	}
	snap := g.Snapshot(0)
	if snap.PendingRaiseFor != 0 {
		t.Fatalf("PendingRaiseFor = %d, want 0", snap.PendingRaiseFor)
	}
	if snap.TurnPlayer != 0 {
		t.Fatalf("TurnPlayer after ask = %d, want 0", snap.TurnPlayer)
	}

	if err := g.RespondTruco(0, true); err != nil {
		t.Fatalf("RespondTruco accept: %v", err)
	}
	snap = g.Snapshot(0)
	if snap.CurrentHand.Stake != 3 {
		t.Fatalf("stake = %d, want 3", snap.CurrentHand.Stake)
	}
	if snap.PendingRaiseFor != -1 {
		t.Fatalf("PendingRaiseFor = %d, want -1", snap.PendingRaiseFor)
	}
	if snap.TurnPlayer != 3 {
		t.Fatalf("TurnPlayer after accept = %d, want 3", snap.TurnPlayer)
	}
}

func TestAskTrucoRejectsOutOfTurn(t *testing.T) {
	names := []string{"p1", "p2"}
	cpus := []bool{false, false}
	g, err := NewGame(names, cpus)
	if err != nil {
		t.Fatalf("NewGame error: %v", err)
	}
	turn := g.Snapshot(0).TurnPlayer
	other := 1 - turn
	if err := g.AskTruco(other); err == nil {
		t.Fatalf("esperava erro ao pedir truco fora do turno")
	}
}

func TestRaiseTrucoChainsToNextLevel(t *testing.T) {
	g, err := NewGame([]string{"p1", "p2"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame error: %v", err)
	}

	asker := g.Snapshot(0).TurnPlayer
	responder := 1 - asker

	if err := g.AskTruco(asker); err != nil {
		t.Fatalf("AskTruco: %v", err)
	}
	if err := g.RaiseTruco(responder); err != nil {
		t.Fatalf("RaiseTruco: %v", err)
	}

	s := g.Snapshot(0)
	if s.CurrentHand.Stake != 3 {
		t.Fatalf("stake after raise = %d, want 3", s.CurrentHand.Stake)
	}
	if s.PendingRaiseFor != g.TeamOfPlayer(asker) {
		t.Fatalf("pendingRaiseFor = %d, want %d", s.PendingRaiseFor, g.TeamOfPlayer(asker))
	}
	if s.PendingRaiseTo != 6 {
		t.Fatalf("pendingRaiseTo = %d, want 6", s.PendingRaiseTo)
	}
	if s.CurrentHand.RaiseRequester != responder {
		t.Fatalf("raiseRequester = %d, want %d", s.CurrentHand.RaiseRequester, responder)
	}

	if err := g.RespondTruco(asker, true); err != nil {
		t.Fatalf("RespondTruco accept after raise: %v", err)
	}
	s = g.Snapshot(0)
	if s.CurrentHand.Stake != 6 {
		t.Fatalf("stake after accepting raise = %d, want 6", s.CurrentHand.Stake)
	}
	if s.PendingRaiseFor != -1 {
		t.Fatalf("pendingRaiseFor = %d, want -1", s.PendingRaiseFor)
	}
	if s.TurnPlayer != responder {
		t.Fatalf("turn after accept should return to re-raiser: got %d, want %d", s.TurnPlayer, responder)
	}
}

func TestRaiseTrucoRejectsAtMaximum(t *testing.T) {
	g, err := NewGame([]string{"p1", "p2"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame error: %v", err)
	}

	asker := g.Snapshot(0).TurnPlayer
	responder := 1 - asker

	// Sobe até 9 e deixa pendente pedido para 12.
	for _, expectedStake := range []int{3, 6, 9} {
		g.mu.Lock()
		g.hand.Turn = asker
		g.mu.Unlock()
		if err := g.AskTruco(asker); err != nil {
			t.Fatalf("AskTruco (target %d): %v", expectedStake, err)
		}
		if err := g.RespondTruco(responder, true); err != nil {
			t.Fatalf("RespondTruco (target %d): %v", expectedStake, err)
		}
		if got := g.Snapshot(0).CurrentHand.Stake; got != expectedStake {
			t.Fatalf("stake = %d, want %d", got, expectedStake)
		}
		asker, responder = responder, asker
	}

	g.mu.Lock()
	g.hand.Turn = asker
	g.mu.Unlock()
	if err := g.AskTruco(asker); err != nil {
		t.Fatalf("AskTruco to 12 pending: %v", err)
	}
	if err := g.RaiseTruco(responder); err == nil {
		t.Fatalf("expected RaiseTruco error when pending call is already to 12")
	}
}

func TestCheckHandEndTieThenWinEndsHand(t *testing.T) {
	g := &Game{
		players: []Player{{ID: 0, Team: 0}, {ID: 1, Team: 1}},
		hand: HandState{
			TrickWins:    map[int]int{0: 1, 1: 0},
			TrickResults: []int{-1, 0},
			RoundStart:   0,
			WinnerTeam:   -1,
		},
	}
	g.checkHandEndLocked()

	if !g.hand.Finished {
		t.Fatalf("hand deveria terminar após empate na primeira e vitória na segunda")
	}
	if g.hand.WinnerTeam != 0 {
		t.Fatalf("winnerTeam = %d, want 0", g.hand.WinnerTeam)
	}
}

func TestCheckHandEndAllTieUsesRoundStarterTeam(t *testing.T) {
	g := &Game{
		players: []Player{{ID: 0, Team: 0}, {ID: 1, Team: 1}},
		hand: HandState{
			TrickWins:    map[int]int{0: 0, 1: 0},
			TrickResults: []int{-1, -1, -1},
			RoundStart:   1,
			WinnerTeam:   -1,
		},
	}
	g.checkHandEndLocked()

	if !g.hand.Finished {
		t.Fatalf("hand deveria terminar após três vazas")
	}
	if g.hand.WinnerTeam != 1 {
		t.Fatalf("winnerTeam = %d, want 1", g.hand.WinnerTeam)
	}
}

func TestTrickPilesPersistAcrossResolvedTricks(t *testing.T) {
	g := &Game{
		players: []Player{
			{ID: 0, Name: "p1", Team: 0},
			{ID: 1, Name: "p2", Team: 1},
		},
		numPlayers: 2,
		hand: HandState{
			Manilha:      "-",
			RoundStart:   0,
			TrickWins:    map[int]int{0: 0, 1: 0},
			TrickResults: []int{},
		},
		lastTrickWinner: -1,
		lastTrickTeam:   -1,
	}

	g.hand.RoundCards = []PlayedCard{
		{PlayerID: 0, Card: Card{Rank: "3", Suit: "Paus"}},
		{PlayerID: 1, Card: Card{Rank: "2", Suit: "Paus"}},
	}
	g.resolveTrickLocked()

	if got := len(g.trickPiles); got != 1 {
		t.Fatalf("trickPiles len after first trick = %d, want 1", got)
	}
	if got := g.trickPiles[0].Winner; got != 0 {
		t.Fatalf("first pile winner = %d, want 0", got)
	}
	if got := len(g.trickPiles[0].Cards); got != 2 {
		t.Fatalf("first pile card count = %d, want 2", got)
	}

	g.hand.RoundCards = []PlayedCard{
		{PlayerID: 1, Card: Card{Rank: "3", Suit: "Copas"}},
		{PlayerID: 0, Card: Card{Rank: "2", Suit: "Copas"}},
	}
	g.resolveTrickLocked()

	snap := g.Snapshot(0)
	if got := len(snap.TrickPiles); got != 2 {
		t.Fatalf("snapshot.TrickPiles len = %d, want 2", got)
	}
	if got := snap.TrickPiles[0].Winner; got != 0 {
		t.Fatalf("snapshot first pile winner = %d, want 0", got)
	}
	if got := snap.TrickPiles[1].Winner; got != 1 {
		t.Fatalf("snapshot second pile winner = %d, want 1", got)
	}
	if got := len(snap.TrickPiles[0].Cards); got != 2 {
		t.Fatalf("snapshot first pile card count = %d, want 2", got)
	}
}

func TestPlayCardRejectsOutOfTurn(t *testing.T) {
	names := []string{"p1", "p2"}
	cpus := []bool{false, false}
	g, err := NewGame(names, cpus)
	if err != nil {
		t.Fatalf("NewGame error: %v", err)
	}
	turn := g.Snapshot(0).TurnPlayer
	wrong := 1 - turn

	if err := g.PlayCard(wrong, 0); err == nil {
		t.Fatalf("esperava erro ao jogar fora do turno")
	}
}
