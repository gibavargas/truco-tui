#ifndef TRUCO_STRATEGY_HPP
#define TRUCO_STRATEGY_HPP

#include "GameState.hpp"
#include <random>

namespace truco {

// Decision returned by the AI
struct Decision {
    enum Kind {
        PLAY,
        PLAY_FACEDOWN,
        ASK_TRUCO,
        RAISE,
        ACCEPT,
        REFUSE
    };
    Kind kind;
    int cardIndex;
};

// Abstract base class — Strategy Pattern
// Each subclass implements a different personality
class Strategy {
public:
    virtual ~Strategy() = default;

    // Main entry point: given game state, decide what to do
    virtual Decision decide(const GameState& state) = 0;

    // Name of this strategy (for debugging)
    virtual const char* name() const = 0;

protected:
    // RNG shared across strategies
    std::mt19937& rng() {
        static std::mt19937 gen(std::random_device{}());
        return gen;
    }

    double random() {
        std::uniform_real_distribution<double> dist(0.0, 1.0);
        return dist(rng());
    }

    int randomInt(int exclusiveMax) {
        std::uniform_int_distribution<int> dist(0, exclusiveMax - 1);
        return dist(rng());
    }

    // === Shared utilities (used by all strategies) ===

    // Find the weakest card index in hand
    int weakestCard(const std::vector<Card>& hand, Rank manilha) const;

    // Find the strongest card index
    int strongestCard(const std::vector<Card>& hand, Rank manilha) const;

    // Find the middle card index (by power)
    int middleCard(const std::vector<Card>& hand, Rank manilha) const;

    // Find the lowest card that beats minPower, or -1
    int lowestWinningCard(const std::vector<Card>& hand, Rank manilha, int minPower) const;

    // Default play when no special condition applies
    Decision defaultPlay(const GameState& state);

    // Check if we should respond to truco (accept/refuse)
    // Returns the raise decision. Each strategy customizes thresholds.
    virtual Decision respondToRaise(const GameState& state, const HandQuality& hq) = 0;

    // Check if we should ask for truco
    virtual bool shouldAskTruco(const GameState& state, const HandQuality& hq) = 0;
};

// === Concrete Strategies ===

// Balanced: adaptable, medium aggression
class BalancedStrategy : public Strategy {
public:
    Decision decide(const GameState& state) override;
    const char* name() const override { return "Balanced"; }

protected:
    Decision respondToRaise(const GameState& state, const HandQuality& hq) override;
    bool shouldAskTruco(const GameState& state, const HandQuality& hq) override;
};

// Aggressive: bluffs often, raises frequently, plays face-down to confuse
class AggressiveStrategy : public Strategy {
public:
    Decision decide(const GameState& state) override;
    const char* name() const override { return "Aggressive"; }

protected:
    Decision respondToRaise(const GameState& state, const HandQuality& hq) override;
    bool shouldAskTruco(const GameState& state, const HandQuality& hq) override;
};

// Conservative: safe play, rarely bluffs, saves strong cards
class ConservativeStrategy : public Strategy {
public:
    Decision decide(const GameState& state) override;
    const char* name() const override { return "Conservative"; }

protected:
    Decision respondToRaise(const GameState& state, const HandQuality& hq) override;
    bool shouldAskTruco(const GameState& state, const HandQuality& hq) override;
};

// Bluffer: unpredictable, heavy face-down play, constant pressure
class BlufferStrategy : public Strategy {
public:
    Decision decide(const GameState& state) override;
    const char* name() const override { return "Bluffer"; }

protected:
    Decision respondToRaise(const GameState& state, const HandQuality& hq) override;
    bool shouldAskTruco(const GameState& state, const HandQuality& hq) override;
};

} // namespace truco

#endif
