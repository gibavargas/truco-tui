#include "GameState.hpp"

namespace truco {

TableLead analyzeTable(const GameState& state) {
    TableLead lead{-1, -1, false};
    if (state.tableCards.empty()) return lead;

    lead.hasCards = true;
    int bestPower = -1;
    std::map<int, int> teamCounts;

    for (const auto& pc : state.tableCards) {
        int p = pc.card.power(state.manilha);
        if (p > bestPower) {
            bestPower = p;
            teamCounts.clear();
        }
        if (p == bestPower) {
            teamCounts[pc.team]++;
        }
    }

    lead.bestPower = bestPower;
    if (teamCounts.size() == 1) {
        lead.leadingTeam = teamCounts.begin()->first;
    } else {
        lead.leadingTeam = -1; // tie
    }
    return lead;
}

HandQuality evaluateHand(const std::vector<Card>& hand, Rank manilha) {
    HandQuality hq{0, 0, 0, 0, 0.0};
    for (const auto& c : hand) {
        int p = c.power(manilha);
        if (p >= 100) {
            hq.manilhas++;
            if (p >= 103) hq.score += 4.5;  // high manilha (Clubs, Hearts)
            else hq.score += 3.0;           // low manilha (Spades, Diamonds)
        } else if (p >= 8) {
            hq.highCards++;
            hq.score += 1.8;  // 3, 2
        } else if (p >= 5) {
            hq.mediumCards++;
            hq.score += 1.0;  // A, K
        } else if (p >= 5) {
            hq.mediumCards++;
            hq.score += 0.4;  // J, Q
        } else {
            hq.lowCards++;
        }
    }
    return hq;
}

} // namespace truco
