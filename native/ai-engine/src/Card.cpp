#include "Card.hpp"
#include <sstream>

namespace truco {

// Must match Go's normalPower map
// R3=10, R2=9, RA=8, RK=7, RJ=6, RQ=5, R7=4, R6=3, R5=2, R4=1
static const int normalPower[] = {
    1,  // R4
    2,  // R5
    3,  // R6
    4,  // R7
    5,  // RQ
    6,  // RJ
    7,  // RK
    8,  // RA
    9,  // R2
    10  // R3
};

int manilhaSuitPower(Suit s) {
    switch (s) {
        case Suit::Clubs:    return 4;
        case Suit::Hearts:   return 3;
        case Suit::Spades:   return 2;
        case Suit::Diamonds: return 1;
    }
    return 0;
}

int Card::power(Rank manilha) const {
    if (rank == manilha) {
        return 100 + manilhaSuitPower(suit);
    }
    return normalPower[static_cast<int>(rank)];
}

static const char* suitNames[] = {"Ouros", "Espadas", "Copas", "Paus"};
static const char* rankNames[] = {"4", "5", "6", "7", "Q", "J", "K", "A", "2", "3"};

std::string Card::toString() const {
    std::ostringstream oss;
    oss << rankNames[static_cast<int>(rank)] << " de " << suitNames[static_cast<int>(suit)];
    return oss.str();
}

} // namespace truco
