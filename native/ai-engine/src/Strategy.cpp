#include "Strategy.hpp"
#include <algorithm>
#include <memory>

namespace truco {

// === Shared Utilities ===

int Strategy::weakestCard(const std::vector<Card>& hand, Rank manilha) const {
    int idx = 0;
    int best = hand[0].power(manilha);
    for (int i = 1; i < (int)hand.size(); i++) {
        int p = hand[i].power(manilha);
        if (p < best) { best = p; idx = i; }
    }
    return idx;
}

int Strategy::strongestCard(const std::vector<Card>& hand, Rank manilha) const {
    int idx = 0;
    int best = hand[0].power(manilha);
    for (int i = 1; i < (int)hand.size(); i++) {
        int p = hand[i].power(manilha);
        if (p > best) { best = p; idx = i; }
    }
    return idx;
}

int Strategy::middleCard(const std::vector<Card>& hand, Rank manilha) const {
    std::vector<std::pair<int,int>> sorted; // (power, originalIdx)
    for (int i = 0; i < (int)hand.size(); i++) {
        sorted.push_back({hand[i].power(manilha), i});
    }
    std::sort(sorted.begin(), sorted.end());
    return sorted[sorted.size() / 2].second;
}

int Strategy::lowestWinningCard(const std::vector<Card>& hand, Rank manilha, int minPower) const {
    int idx = -1;
    int best = 1000;
    for (int i = 0; i < (int)hand.size(); i++) {
        int p = hand[i].power(manilha);
        if (p <= minPower) continue;
        if (p < best) { best = p; idx = i; }
    }
    return idx;
}

Decision Strategy::defaultPlay(const GameState& state) {
    TableLead lead = analyzeTable(state);
    int weak = weakestCard(state.myHand, state.manilha);

    // No cards on table — opening a trick
    if (!lead.hasCards) {
        if (state.round == 1 && state.myHand.size() >= 3) {
            double r = random();
            if (r < 0.10) return {Decision::PLAY, strongestCard(state.myHand, state.manilha)};
            if (r < 0.30) return {Decision::PLAY, weak};
            return {Decision::PLAY, middleCard(state.myHand, state.manilha)};
        }
        return {Decision::PLAY, weak};
    }

    // Our team is winning the trick — dump the weakest
    if (lead.leadingTeam == state.myTeam) {
        return {Decision::PLAY, weak};
    }

    // Opponent is winning — try to beat with lowest winning card
    int winner = lowestWinningCard(state.myHand, state.manilha, lead.bestPower);
    if (winner >= 0) {
        // Don't waste a manilha in round 1 unless we have to
        int p = state.myHand[winner].power(state.manilha);
        if (p >= 100 && state.round == 1 && state.myHand.size() > 1) {
            if (random() < 0.30) return {Decision::PLAY, weak};
        }
        return {Decision::PLAY, winner};
    }

    return {Decision::PLAY, weak};
}

// === Balanced ===

Decision BalancedStrategy::decide(const GameState& state) {
    HandQuality hq = evaluateHand(state.myHand, state.manilha);

    if (state.pendingRaiseFor == state.myTeam) {
        return respondToRaise(state, hq);
    }
    if (state.canAskTruco && shouldAskTruco(state, hq)) {
        return {Decision::ASK_TRUCO, -1};
    }

    // Face-down play decisions
    TableLead lead = analyzeTable(state);
    Decision play = defaultPlay(state);

    // If leading and have weak card, hide it 40% of the time
    if (lead.hasCards && lead.leadingTeam == state.myTeam && state.myHand.size() > 1) {
        int weak = weakestCard(state.myHand, state.manilha);
        if (state.myHand[weak].power(state.manilha) < 5 && random() < 0.40) {
            return {state.round >= 2 ? Decision::PLAY_FACEDOWN : Decision::PLAY, weak};
        }
    }
    // Strong hand round 1 — bluff face-down 25%
    if (!lead.hasCards && state.round == 1 && hq.score >= 6.0 && random() < 0.25) {
        return {state.round >= 2 ? Decision::PLAY_FACEDOWN : Decision::PLAY, play.cardIndex};
    }

    return play;
}

Decision BalancedStrategy::respondToRaise(const GameState& state, const HandQuality& hq) {
    int stake = state.stake;
    if (hq.manilhas >= 3 && stake <= 6) return {Decision::RAISE, -1};
    if (hq.manilhas >= 2 && stake <= 3) return {Decision::RAISE, -1};

    if (hq.manilhas >= 1 && stake <= 6) return {Decision::ACCEPT, -1};
    if (hq.highCards >= 2 && stake <= 3) return {Decision::ACCEPT, -1};
    if (state.ourTrickWins > state.theirTrickWins && stake <= 9) return {Decision::ACCEPT, -1};
    if (stake <= 3 && state.round <= 2 && (hq.manilhas >= 1 || hq.highCards >= 2)) return {Decision::ACCEPT, -1};
    if (stake <= 3 && random() < 0.15) return {Decision::ACCEPT, -1}; // bluff
    return {Decision::REFUSE, -1};
}

bool BalancedStrategy::shouldAskTruco(const GameState& state, const HandQuality& hq) {
    if (state.stake >= 12) return false;
    if (hq.manilhas >= 3) return true;
    if (hq.manilhas >= 2 && state.stake <= 3) return true;
    if (state.ourTrickWins > state.theirTrickWins && hq.score >= 5.0 && state.stake <= 3) return true;
    if (state.stake == 1 && state.round == 1 && random() < 0.10) return true; // bluff
    return false;
}

// === Aggressive ===

Decision AggressiveStrategy::decide(const GameState& state) {
    HandQuality hq = evaluateHand(state.myHand, state.manilha);

    if (state.pendingRaiseFor == state.myTeam) {
        return respondToRaise(state, hq);
    }
    if (state.canAskTruco && shouldAskTruco(state, hq)) {
        return {Decision::ASK_TRUCO, -1};
    }

    // Aggressive face-down: hides cards often
    TableLead lead = analyzeTable(state);
    Decision play = defaultPlay(state);

    if (lead.hasCards && lead.leadingTeam == state.myTeam && state.myHand.size() > 1) {
        int weak = weakestCard(state.myHand, state.manilha);
        if (random() < 0.60) return {state.round >= 2 ? Decision::PLAY_FACEDOWN : Decision::PLAY, weak};
    }
    if (!lead.hasCards && state.round == 1 && random() < 0.35) {
        return {state.round >= 2 ? Decision::PLAY_FACEDOWN : Decision::PLAY, play.cardIndex};
    }

    return play;
}

Decision AggressiveStrategy::respondToRaise(const GameState& state, const HandQuality& hq) {
    int stake = state.stake;
    if (hq.manilhas >= 2 && stake <= 9) return {Decision::RAISE, -1};
    if (hq.manilhas >= 1 && hq.highCards >= 1 && stake <= 6) return {Decision::RAISE, -1};

    if (hq.manilhas >= 1) return {Decision::ACCEPT, -1};
    if (hq.highCards >= 2 && stake <= 6) return {Decision::ACCEPT, -1};
    if (state.ourTrickWins > state.theirTrickWins) return {Decision::ACCEPT, -1};
    if (stake <= 6 && random() < 0.30) return {Decision::ACCEPT, -1}; // frequent bluff
    if (stake <= 3 && random() < 0.20) return {Decision::RAISE, -1};  // aggressive raise bluff
    return {Decision::REFUSE, -1};
}

bool AggressiveStrategy::shouldAskTruco(const GameState& state, const HandQuality& hq) {
    if (state.stake >= 12) return false;
    if (hq.manilhas >= 2) return true;
    if (hq.manilhas >= 1 && hq.highCards >= 1 && state.stake <= 3) return true;
    if (hq.highCards >= 2 && state.stake <= 3) return true;
    if (state.ourTrickWins > state.theirTrickWins && state.stake <= 6) return true;
    if (state.stake == 1 && state.round == 1 && random() < 0.20) return true; // frequent bluff
    return false;
}

// === Conservative ===

Decision ConservativeStrategy::decide(const GameState& state) {
    HandQuality hq = evaluateHand(state.myHand, state.manilha);

    if (state.pendingRaiseFor == state.myTeam) {
        return respondToRaise(state, hq);
    }
    if (state.canAskTruco && shouldAskTruco(state, hq)) {
        return {Decision::ASK_TRUCO, -1};
    }

    // Conservative: rarely plays face-down
    return defaultPlay(state);
}

Decision ConservativeStrategy::respondToRaise(const GameState& state, const HandQuality& hq) {
    int stake = state.stake;
    if (hq.manilhas >= 3 && stake <= 3) return {Decision::RAISE, -1};

    if (hq.manilhas >= 2) return {Decision::ACCEPT, -1};
    if (hq.manilhas >= 1 && hq.highCards >= 1) return {Decision::ACCEPT, -1};
    if (hq.manilhas >= 1 && stake <= 3) return {Decision::ACCEPT, -1};
    if (state.ourTrickWins > state.theirTrickWins && stake <= 6) return {Decision::ACCEPT, -1};
    if (stake <= 3 && hq.highCards >= 2) return {Decision::ACCEPT, -1};
    // No bluffing — conservative refuses with weak hands
    return {Decision::REFUSE, -1};
}

bool ConservativeStrategy::shouldAskTruco(const GameState& state, const HandQuality& hq) {
    if (state.stake >= 12) return false;
    if (hq.manilhas >= 3) return true;
    if (hq.manilhas >= 2 && hq.highCards >= 1) return true;
    // No bluffing
    return false;
}

// === Bluffer ===

Decision BlufferStrategy::decide(const GameState& state) {
    HandQuality hq = evaluateHand(state.myHand, state.manilha);

    if (state.pendingRaiseFor == state.myTeam) {
        return respondToRaise(state, hq);
    }
    if (state.canAskTruco && shouldAskTruco(state, hq)) {
        return {Decision::ASK_TRUCO, -1};
    }

    // Bluffer: heavy face-down usage
    TableLead lead = analyzeTable(state);
    Decision play = defaultPlay(state);

    if (lead.hasCards && lead.leadingTeam == state.myTeam && state.myHand.size() > 1) {
        if (random() < 0.50) {
            int weak = weakestCard(state.myHand, state.manilha);
            return {state.round >= 2 ? Decision::PLAY_FACEDOWN : Decision::PLAY, weak};
        }
    }
    if (!lead.hasCards && random() < 0.45) {
        return {state.round >= 2 ? Decision::PLAY_FACEDOWN : Decision::PLAY, play.cardIndex};
    }

    return play;
}

Decision BlufferStrategy::respondToRaise(const GameState& state, const HandQuality& hq) {
    int stake = state.stake;
    // Bluffer raises aggressively even with mediocre hands
    if (stake <= 6 && random() < 0.35) return {Decision::RAISE, -1};
    if (hq.manilhas >= 1 && stake <= 6) return {Decision::RAISE, -1};

    // Accepts almost anything with occasional refusal for mind games
    if (stake <= 6) return {Decision::ACCEPT, -1};
    if (hq.manilhas >= 1) return {Decision::ACCEPT, -1};
    if (random() < 0.40) return {Decision::ACCEPT, -1};
    return {Decision::REFUSE, -1};
}

bool BlufferStrategy::shouldAskTruco(const GameState& state, const HandQuality& hq) {
    if (state.stake >= 12) return false;
    // Bluffer asks truco very frequently
    if (state.stake == 1 && state.round == 1) {
        if (random() < 0.30) return true;
    }
    if (hq.manilhas >= 1) return true;
    if (hq.score >= 3.0 && state.stake <= 3) return true;
    if (random() < 0.20) return true; // pure bluff
    return false;
}

// === Hybrid ===

HybridStrategy::HybridStrategy(double wB, double wA, double wC, double wBl) {
    // Normalize so weights sum to 1.0
    double total = wB + wA + wC + wBl;
    if (total <= 0.0) total = 4.0;
    weights_[0] = wB / total;
    weights_[1] = wA / total;
    weights_[2] = wC / total;
    weights_[3] = wBl / total;
}

int HybridStrategy::rollStrategy() {
    double r = random();
    double cumulative = 0.0;
    for (int i = 0; i < 4; i++) {
        cumulative += weights_[i];
        if (r < cumulative) return i;
    }
    return 3; // fallback
}

std::unique_ptr<Strategy> HybridStrategy::getStrategy(int idx) {
    switch (idx) {
        case 0: return std::make_unique<BalancedStrategy>();
        case 1: return std::make_unique<AggressiveStrategy>();
        case 2: return std::make_unique<ConservativeStrategy>();
        case 3: return std::make_unique<BlufferStrategy>();
        default: return std::make_unique<BalancedStrategy>();
    }
}

Decision HybridStrategy::decide(const GameState& state) {
    // Roll once per turn — use the same strategy for the whole decision
    int idx = rollStrategy();
    auto strat = getStrategy(idx);
    return strat->decide(state);
}

// respondToRaise and shouldAskTruco are never called directly on Hybrid —
// decide() delegates everything to the rolled strategy.
// Keep minimal stubs to satisfy the abstract base contract.
Decision HybridStrategy::respondToRaise(const GameState&, const HandQuality&) {
    return {Decision::REFUSE, -1};
}
bool HybridStrategy::shouldAskTruco(const GameState&, const HandQuality&) {
    return false;
}

} // namespace truco
