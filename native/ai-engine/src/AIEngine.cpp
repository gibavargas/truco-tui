#include "AIEngine.hpp"
#include <cmath>

namespace truco {

AIEngine& AIEngine::instance() {
    static AIEngine engine;
    return engine;
}

// Generate unique, deterministic personality weights from a player ID.
// Each CPU player gets a different blend so they all play differently.
// Uses simple hash-based generation.
static HybridStrategy makeHybridForPlayer(int playerID) {
    // Hash the playerID to produce 4 distinct weights
    unsigned int seed = static_cast<unsigned int>(playerID) * 2654435761u;
    auto hash = [&seed]() {
        // xorshift32
        seed ^= seed << 13;
        seed ^= seed >> 17;
        seed ^= seed << 5;
        return (seed % 100) / 100.0; // 0.00 - 0.99
    };

    double wB = 0.15 + hash() * 0.35;  // Balanced: 15-50%
    double wA = 0.10 + hash() * 0.35;  // Aggressive: 10-45%
    double wC = 0.05 + hash() * 0.30;  // Conservative: 5-35%
    double wBl = 0.05 + hash() * 0.30; // Bluffer: 5-35%

    return HybridStrategy(wB, wA, wC, wBl);
}

std::unique_ptr<Strategy> AIEngine::createStrategy(Personality p) {
    switch (p) {
        case Personality::Aggressive:   return std::make_unique<AggressiveStrategy>();
        case Personality::Conservative: return std::make_unique<ConservativeStrategy>();
        case Personality::Bluffer:      return std::make_unique<BlufferStrategy>();
        case Personality::Balanced:
        default: {
            // For "Balanced" (the default, personality ID 0), use Hybrid
            // with unique weights — actual playerID is passed via decide()
            return std::make_unique<BalancedStrategy>();
        }
    }
}

Decision AIEngine::decide(int personalityId, const GameState& state) {
    Personality p = static_cast<Personality>(personalityId & 3);

    // Personality 0 = Hybrid: every CPU gets a unique blend
    if (p == Personality::Balanced) {
        auto hybrid = makeHybridForPlayer(state.myPlayerId);
        return hybrid.decide(state);
    }

    auto strategy = createStrategy(p);
    return strategy->decide(state);
}

const char* AIEngine::personalityName(int personalityId) {
    switch (personalityId & 3) {
        case 0: return "Hybrid";
        case 1: return "Aggressive";
        case 2: return "Conservative";
        case 3: return "Bluffer";
        default: return "Hybrid";
    }
}

} // namespace truco
