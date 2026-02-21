package truco

// Regras de força das cartas no Truco Paulista.
// Ordem normal (mais forte para mais fraca): 3, 2, A, K, J, Q, 7, 6, 5, 4.
// Manilha é dinâmica: próxima carta após a vira.

var normalPower = map[Rank]int{
	R3: 10,
	R2: 9,
	RA: 8,
	RK: 7,
	RJ: 6,
	RQ: 5,
	R7: 4,
	R6: 3,
	R5: 2,
	R4: 1,
}

// Ordem de força entre naipes de manilha no Truco Paulista.
// Mais forte -> mais fraco.
var manilhaSuitPower = map[Suit]int{
	Clubs:    4, // Paus
	Hearts:   3, // Copas
	Spades:   2, // Espadas
	Diamonds: 1, // Ouros
}

var rankCycle = []Rank{R4, R5, R6, R7, RQ, RJ, RK, RA, R2, R3}

func NextRank(r Rank) Rank {
	for i, current := range rankCycle {
		if current == r {
			return rankCycle[(i+1)%len(rankCycle)]
		}
	}
	return R4
}

func CardPower(c Card, manilha Rank) int {
	if c.Rank == manilha {
		return 100 + manilhaSuitPower[c.Suit]
	}
	return normalPower[c.Rank]
}

func CompareCards(a, b Card, manilha Rank) int {
	pa := CardPower(a, manilha)
	pb := CardPower(b, manilha)
	if pa > pb {
		return 1
	}
	if pb > pa {
		return -1
	}
	return 0
}
