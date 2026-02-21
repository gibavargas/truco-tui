package truco

import "testing"

func TestNextRankCycle(t *testing.T) {
	if got := NextRank(R4); got != R5 {
		t.Fatalf("NextRank(R4) = %s, want %s", got, R5)
	}
	if got := NextRank(R3); got != R4 {
		t.Fatalf("NextRank(R3) = %s, want %s", got, R4)
	}
}

func TestCompareCardsManilhaSuitOrder(t *testing.T) {
	manilha := R5
	cPaus := Card{Rank: manilha, Suit: Clubs}
	cCopas := Card{Rank: manilha, Suit: Hearts}
	cEspadas := Card{Rank: manilha, Suit: Spades}
	cOuros := Card{Rank: manilha, Suit: Diamonds}

	if CompareCards(cPaus, cCopas, manilha) <= 0 {
		t.Fatalf("paus deve vencer copas na manilha")
	}
	if CompareCards(cCopas, cEspadas, manilha) <= 0 {
		t.Fatalf("copas deve vencer espadas na manilha")
	}
	if CompareCards(cEspadas, cOuros, manilha) <= 0 {
		t.Fatalf("espadas deve vencer ouros na manilha")
	}
}

func TestNormalHierarchy(t *testing.T) {
	manilha := R6
	three := Card{Rank: R3, Suit: Clubs}
	two := Card{Rank: R2, Suit: Clubs}
	four := Card{Rank: R4, Suit: Clubs}

	if CompareCards(three, two, manilha) <= 0 {
		t.Fatalf("3 deveria vencer 2")
	}
	if CompareCards(two, four, manilha) <= 0 {
		t.Fatalf("2 deveria vencer 4")
	}
}
