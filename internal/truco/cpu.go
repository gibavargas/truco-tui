package truco

// CPUAction descreve a ação escolhida pela IA.
type CPUAction struct {
	Kind      string // "play", "ask_truco", "raise", "accept", "refuse"
	CardIndex int
}

// DecideCPUAction executa uma heurística simples para manter o jogo fluido no modo offline.
func DecideCPUAction(g *Game, playerID int) CPUAction {
	snap := g.Snapshot(playerID)
	team := g.TeamOfPlayer(playerID)
	cards := g.HandCards(playerID)
	if len(cards) == 0 {
		return CPUAction{Kind: "refuse"}
	}

	strong := 0
	for _, c := range cards {
		if CardPower(c, snap.CurrentHand.Manilha) >= 9 {
			strong++
		}
	}

	if snap.PendingRaiseFor == team {
		if shouldRaiseOnResponse(snap, team, strong) {
			return CPUAction{Kind: "raise"}
		}
		if shouldAcceptRaise(snap, team, strong) {
			return CPUAction{Kind: "accept"}
		}
		return CPUAction{Kind: "refuse"}
	}

	// Só tenta pedir truco quando a ação é realmente válida para evitar
	// ficar em loop de erro ("seu time já aumentou a aposta por último").
	if g.CanAskTrucoByPlayer(playerID) && shouldAskTruco(snap, team, strong) {
		return CPUAction{Kind: "ask_truco"}
	}

	return CPUAction{Kind: "play", CardIndex: chooseCardToPlay(snap, team, cards)}
}

func shouldAcceptRaise(s Snapshot, team, strongCards int) bool {
	if strongCards >= 2 {
		return true
	}
	if strongCards >= 1 && s.CurrentHand.Stake <= 6 {
		return true
	}
	if s.CurrentHand.Stake <= 3 && s.CurrentHand.Round <= 2 {
		return true
	}
	if s.CurrentHand.TrickWins[team] > s.CurrentHand.TrickWins[1-team] && s.CurrentHand.Stake <= 9 {
		return true
	}
	return false
}

func shouldRaiseOnResponse(s Snapshot, team, strongCards int) bool {
	if s.CurrentHand.Stake >= 9 {
		return false
	}
	if strongCards >= 3 {
		return true
	}
	if strongCards >= 2 && s.CurrentHand.Stake <= 3 {
		return true
	}
	if strongCards >= 2 && s.CurrentHand.Round <= 2 &&
		s.CurrentHand.TrickWins[team] > s.CurrentHand.TrickWins[1-team] {
		return true
	}
	return false
}

func shouldAskTruco(s Snapshot, team, strongCards int) bool {
	if s.CurrentHand.Stake >= 12 {
		return false
	}
	if strongCards >= 3 {
		return true
	}
	if strongCards >= 2 && s.CurrentHand.Stake <= 3 && s.CurrentHand.Round <= 2 {
		return true
	}
	if strongCards >= 1 && s.CurrentHand.Stake == 1 && s.CurrentHand.TrickWins[team] > s.CurrentHand.TrickWins[1-team] {
		return true
	}
	return false
}

func chooseCardToPlay(s Snapshot, team int, cards []Card) int {
	manilha := s.CurrentHand.Manilha
	if len(cards) == 1 {
		return 0
	}

	tableBest, leadingTeam, hasTable := tableLeadState(s)
	weakest := weakestCardIndex(cards, manilha)
	if !hasTable {
		if s.CurrentHand.Round == 1 && len(cards) >= 3 {
			return middleCardIndex(cards, manilha)
		}
		return weakest
	}
	if leadingTeam == team {
		return weakest
	}
	if winner := lowestWinningCardIndex(cards, manilha, tableBest); winner >= 0 {
		return winner
	}
	return weakest
}

func tableLeadState(s Snapshot) (bestPower int, leadingTeam int, hasCards bool) {
	if len(s.CurrentHand.RoundCards) == 0 {
		return -1, -1, false
	}
	bestPower = -1
	teamCounts := map[int]int{}
	for _, pc := range s.CurrentHand.RoundCards {
		p := CardPower(pc.Card, s.CurrentHand.Manilha)
		if p > bestPower {
			bestPower = p
			teamCounts = map[int]int{}
		}
		if p == bestPower {
			t := teamForPlayer(s.Players, pc.PlayerID)
			teamCounts[t]++
		}
	}
	if len(teamCounts) == 1 {
		for team := range teamCounts {
			return bestPower, team, true
		}
	}
	return bestPower, -1, true
}

func teamForPlayer(players []Player, playerID int) int {
	for _, p := range players {
		if p.ID == playerID {
			return p.Team
		}
	}
	return -1
}

func weakestCardIndex(cards []Card, manilha Rank) int {
	idx := 0
	best := CardPower(cards[0], manilha)
	for i := 1; i < len(cards); i++ {
		p := CardPower(cards[i], manilha)
		if p < best {
			best = p
			idx = i
		}
	}
	return idx
}

func middleCardIndex(cards []Card, manilha Rank) int {
	type pair struct {
		idx   int
		power int
	}
	ordered := make([]pair, 0, len(cards))
	for i, c := range cards {
		ordered = append(ordered, pair{idx: i, power: CardPower(c, manilha)})
	}
	for i := 0; i < len(ordered)-1; i++ {
		for j := i + 1; j < len(ordered); j++ {
			if ordered[j].power < ordered[i].power {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	return ordered[len(ordered)/2].idx
}

func lowestWinningCardIndex(cards []Card, manilha Rank, minPower int) int {
	idx := -1
	best := 1000
	for i, c := range cards {
		p := CardPower(c, manilha)
		if p <= minPower {
			continue
		}
		if p < best {
			best = p
			idx = i
		}
	}
	return idx
}
