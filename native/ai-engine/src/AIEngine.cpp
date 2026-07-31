#include "AIEngine.hpp"

namespace truco {

AIEngine& AIEngine::instance() {
    static AIEngine engine;
    return engine;
}

std::unique_ptr<Strategy> AIEngine::createStrategy(Personality p) {
    switch (p) {
        case Personality::Aggressive:   return std::make_unique<AggressiveStrategy>();
        case Personality::Conservative: return std::make_unique<ConservativeStrategy>();
        case Personality::Bluffer:      return std::make_unique<BlufferStrategy>();
        case Personality::Balanced:
        default:                         return std::make_unique<BalancedStrategy>();
    }
}

Decision AIEngine::decide(int personalityId, const GameState& state) {
    Personality p = static_cast<Personality>(personalityId & 3);
    auto strategy = createStrategy(p);
    return strategy->decide(state);
}

const char* AIEngine::personalityName(int personalityId) {
    switch (personalityId & 3) {
        case 0: return "Balanced";
        case 1: return "Aggressive";
        case 2: return "Conservative";
        case 3: return "Bluffer";
        default: return "Balanced";
    }
}

} // namespace truco
