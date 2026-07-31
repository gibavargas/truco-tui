#ifndef TRUCO_GAMESTATE_HPP
#define TRUCO_GAMESTATE_HPP

#include "Card.hpp"
#include <vector>
#include <map>

namespace truco {

// A card played on the table
struct PlayedCardInfo {
    Card card;
    int playerId;
    bool faceDown;
    int team;
};

// Represents a player
struct PlayerInfo {
    int id;
    std::string name;
    bool isCpu;
    int team;
    std::vector<Card> hand;
};

// Full game state passed from Go — mirrors the Go Snapshot
struct GameState {
    // Hand info
    Rank manilha;
    Card vira;
    int stake;              // current truco stake (1, 3, 6, 9, 12)
    int round;              // 1, 2, or 3 (current trick)
    int ourTrickWins;       // tricks won by our team
    int theirTrickWins;     // tricks won by their team
    int pendingRaiseFor;    // team that needs to respond to truco, -1 if none
    bool trucoByTeam0;      // team 0 raised last
    bool trucoByTeam1;      // team 1 raised last

    // Match info
    int ourMatchScore;      // match points for our team
    int theirMatchScore;    // match points for their team
    int matchTarget;        // typically 12
    bool matchFinished;

    // Players
    std::vector<PlayerInfo> players;
    int myPlayerId;
    int myTeam;
    int turnPlayer;

    // Cards on the table this trick
    std::vector<PlayedCardInfo> tableCards;

    // My hand
    std::vector<Card> myHand;

    // Can I ask truco right now?
    bool canAskTruco;
};

// Analyze who is leading the current trick
struct TableLead {
    int bestPower;          // power of the best card on table
    int leadingTeam;        // team that's winning, -1 if tie
    bool hasCards;          // are there cards on the table?
};

TableLead analyzeTable(const GameState& state);

// Helper: how many cards of each power tier
struct HandQuality {
    int manilhas;           // power >= 100
    int highCards;          // 3, 2, A (power 8-10)
    int mediumCards;        // K, J, Q (power 5-7)
    int lowCards;           // 7, 6, 5, 4 (power 1-4)
    double score;           // overall 0-10+ quality score
};

HandQuality evaluateHand(const std::vector<Card>& hand, Rank manilha);

} // namespace truco

#endif
