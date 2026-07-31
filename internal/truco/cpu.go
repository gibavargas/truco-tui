package truco

import (
	"math/rand"
)

// CPUAction descreve a ação escolhida pela IA.
type CPUAction struct {
	Kind      string // "play", "play_facedown", "ask_truco", "raise", "accept", "refuse"
	CardIndex int
}

// DecideCPUAction executa uma heurística avançada para o modo offline.
// A IA avalia qualidade da mão, estado do jogo, vazas ganhas e usa aleatoriedade
// para ser menos previsível.
func DecideCPUAction(g *Game, playerID int) CPUAction {
	snap := g.Snapshot(playerID)
	team := g.TeamOfPlayer(playerID)
	cards := g.HandCards(playerID)
	if len(cards) == 0 {
		return CPUAction{Kind: "refuse"}
	}

	strong, medium := countStrongCards(cards, snap.CurrentHand.Manilha)
	handQuality := evaluateHand(cards, snap.CurrentHand.Manilha)

	// === Responder a pedido de truco ===
	if snap.PendingRaiseFor == team {
		if shouldRaiseOnResponse(snap, team, handQuality, strong) {
			return CPUAction{Kind: "raise"}
		}
		if shouldAcceptRaise(snap, team, handQuality, strong, medium) {
			return CPUAction{Kind: "accept"}
		}
		return CPUAction{Kind: "refuse"}
	}

	// === Pedir truco ===
	if g.CanAskTrucoByPlayer(playerID) && shouldAskTruco(snap, team, handQuality, strong) {
		return CPUAction{Kind: "ask_truco"}
	}

	// === Jogar carta ===
	cardIdx, faceDown := chooseCardToPlay(snap, team, cards, g, playerID, handQuality)
	if faceDown {
		return CPUAction{Kind: "play_facedown", CardIndex: cardIdx}
	}
	return CPUAction{Kind: "play", CardIndex: cardIdx}
}

// countStrongCards returns counts of strong (manilhas, power >= 100) and medium (3, 2, A — power 8-10) cards.
func countStrongCards(cards []Card, manilha Rank) (strong, medium int) {
	for _, c := range cards {
		p := CardPower(c, manilha)
		if p >= 100 {
			strong++
		} else if p >= 8 {
			medium++
		}
	}
	return
}

// evaluateHand assigns a numeric quality score (0-10+) to the hand.
// Higher = better hand overall.
func evaluateHand(cards []Card, manilha Rank) float64 {
	var total float64
	for _, c := range cards {
		p := CardPower(c, manilha)
		switch {
		case p >= 103: // Zap (4♣), 7♥ — manilhas altas
			total += 4.5
		case p >= 100: // Outras manilhas (A♠, 7♦)
			total += 3.0
		case p >= 9: // 3, 2 — cartas normais mais fortes
			total += 1.8
		case p >= 7: // A, K
			total += 1.0
		case p >= 5: // J, Q
			total += 0.4
		}
	}
	return total
}

// === Decisões de Truco ===

func shouldAcceptRaise(s Snapshot, team int, handQuality float64, strong, medium int) bool {
	stake := s.CurrentHand.Stake
	ourWins := s.CurrentHand.TrickWins[team]
	theirWins := s.CurrentHand.TrickWins[1-team]

	// Sempre aceita com mão muito forte
	if strong >= 2 && handQuality >= 6 {
		return true
	}
	// Aceita com 1 manilha se a aposta ainda é razoável
	if strong >= 1 && stake <= 6 {
		return true
	}
	// Aceita se já está vencendo as vazas e a aposta não é absurda
	if ourWins > theirWins && stake <= 9 {
		return true
	}
	// Aceita aposta baixa com cartas médias no início
	if stake <= 3 && s.CurrentHand.Round <= 2 && (strong >= 1 || medium >= 2) {
		return true
	}
	// Bluff occasional: aceita aposta baixa mesmo com mão fraca
	if stake <= 3 && rand.Float64() < 0.15 {
		return true
	}
	return false
}

func shouldRaiseOnResponse(s Snapshot, team int, handQuality float64, strong int) bool {
	stake := s.CurrentHand.Stake
	if stake >= 12 {
		return false
	}
	ourWins := s.CurrentHand.TrickWins[team]
	theirWins := s.CurrentHand.TrickWins[1-team]

	// Aumenta com manilhas fortes
	if strong >= 3 && stake <= 6 {
		return true
	}
	if strong >= 2 && stake <= 3 {
		return true
	}
	// Aumenta se está vencendo e tem mão decente
	if ourWins > theirWins && handQuality >= 4 && stake <= 6 {
		return true
	}
	return false
}

func shouldAskTruco(s Snapshot, team int, handQuality float64, strong int) bool {
	stake := s.CurrentHand.Stake
	if stake >= 12 {
		return false
	}
	ourWins := s.CurrentHand.TrickWins[team]
	theirWins := s.CurrentHand.TrickWins[1-team]

	// Pede truco com mão muito forte
	if strong >= 3 {
		return true
	}
	// Pede truco com 2 manilhas no início
	if strong >= 2 && stake <= 3 {
		return true
	}
	// Pede truco se está vencendo vazas com mão boa
	if ourWins > theirWins && handQuality >= 5 && stake <= 3 {
		return true
	}
	// Bluff: pede truco com mão fraca às vezes (10% chance)
	if stake == 1 && s.CurrentHand.Round == 1 && rand.Float64() < 0.10 {
		return true
	}
	return false
}

// === Decisão de Carta ===

// chooseCardToPlay decides which card to play and whether to play it face-down.
// Returns (cardIndex, faceDown).
func chooseCardToPlay(s Snapshot, team int, cards []Card, g *Game, playerID int, handQuality float64) (int, bool) {
	manilha := s.CurrentHand.Manilha
	if len(cards) == 1 {
		return 0, false
	}

	tableBest, leadingTeam, hasTable := tableLeadState(s)
	weakest := weakestCardIndex(cards, manilha)
	strongest := strongestCardIndex(cards, manilha)

	// === Jogo coberto (face-down) ===
	// 1. Se está vencendo a vaza com folga e tem carta fraca, joga coberta pra esconder
	if hasTable && leadingTeam == team && len(cards) > 1 {
		if CardPower(cards[weakest], manilha) < 5 && rand.Float64() < 0.40 {
			return weakest, true
		}
	}
	// 2. Primeira vaza, mão forte: joga coberta pra confundir
	if !hasTable && s.CurrentHand.Round == 1 && len(cards) >= 3 && handQuality >= 6 {
		if rand.Float64() < 0.25 {
			return weakest, true
		}
	}

	// === Sem cartas na mesa ===
	if !hasTable {
		if s.CurrentHand.Round == 1 && len(cards) >= 3 {
			// 70% meio, 20% mais fraca, 10% mais forte (surpresa)
			r := rand.Float64()
			if r < 0.10 {
				return strongest, false
			}
			if r < 0.30 {
				return weakest, false
			}
			return middleCardIndex(cards, manilha), false
		}
		return weakest, false
	}

	// === Nosso time está vencendo a vaza ===
	if leadingTeam == team {
		return weakest, false
	}

	// === Adversário vencendo: tentar ganhar com menor carta possível ===
	if winner := lowestWinningCardIndex(cards, manilha, tableBest); winner >= 0 {
		// Mas se a carta necessária é muito forte e é só rodada 1, pensar 2x
		power := CardPower(cards[winner], manilha)
		if power >= 10 && s.CurrentHand.Round == 1 && len(cards) > 1 {
			// 30% chance de não gastar manilha no início
			if rand.Float64() < 0.30 {
				return weakest, false
			}
		}
		return winner, false
	}

	// Não pode ganhar: descartar mais fraca
	return weakest, false
}

// === Helpers ===

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

func strongestCardIndex(cards []Card, manilha Rank) int {
	idx := 0
	best := CardPower(cards[0], manilha)
	for i := 1; i < len(cards); i++ {
		p := CardPower(cards[i], manilha)
		if p > best {
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
