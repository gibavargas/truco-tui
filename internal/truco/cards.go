package truco

import (
	"fmt"
	randv2 "math/rand/v2"
)

// Suit representa o naipe no baralho de Truco Paulista.
type Suit string

const (
	Diamonds Suit = "Ouros"
	Spades   Suit = "Espadas"
	Hearts   Suit = "Copas"
	Clubs    Suit = "Paus"
)

var allSuits = []Suit{Diamonds, Spades, Hearts, Clubs}

// Rank representa os valores usados no Truco Paulista.
type Rank string

const (
	R4 Rank = "4"
	R5 Rank = "5"
	R6 Rank = "6"
	R7 Rank = "7"
	RQ Rank = "Q"
	RJ Rank = "J"
	RK Rank = "K"
	RA Rank = "A"
	R2 Rank = "2"
	R3 Rank = "3"
)

var allRanks = []Rank{R4, R5, R6, R7, RQ, RJ, RK, RA, R2, R3}

// Card é uma carta individual.
type Card struct {
	Suit Suit
	Rank Rank
}

func (c Card) String() string {
	return fmt.Sprintf("%s de %s", c.Rank, c.Suit)
}

// Deck representa o baralho e permite embaralhar/distribuir.
type Deck struct {
	cards []Card
}

func NewDeck() *Deck {
	cards := make([]Card, 0, len(allSuits)*len(allRanks))
	for _, s := range allSuits {
		for _, r := range allRanks {
			cards = append(cards, Card{Suit: s, Rank: r})
		}
	}
	return &Deck{cards: cards}
}

func (d *Deck) Shuffle(rng *randv2.Rand) {
	for i := len(d.cards) - 1; i > 0; i-- {
		j := rng.IntN(i + 1)
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	}
}

func (d *Deck) Draw() (Card, bool) {
	if len(d.cards) == 0 {
		return Card{}, false
	}
	c := d.cards[0]
	d.cards = d.cards[1:]
	return c, true
}
