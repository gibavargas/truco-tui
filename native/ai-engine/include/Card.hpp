#ifndef TRUCO_CARD_HPP
#define TRUCO_CARD_HPP

#include <string>

namespace truco {

// Mirrors Go's Suit type
enum class Suit { Diamonds, Spades, Hearts, Clubs };

// Mirrors Go's Rank type (numeric values for easy comparison)
enum class Rank { R4, R5, R6, R7, RQ, RJ, RK, RA, R2, R3 };

struct Card {
    Suit suit;
    Rank rank;

    // Compute power level given the manilha (trump) rank
    // Matches Go's CardPower(): manilhas get 100+, normal cards 1-10
    int power(Rank manilha) const;

    // Check if this card is a manilha
    bool isManilha(Rank manilha) const { return rank == manilha; }

    std::string toString() const;
};

// Suit power ordering for manilhas (Clubs > Hearts > Spades > Diamonds)
int manilhaSuitPower(Suit s);

} // namespace truco

#endif
