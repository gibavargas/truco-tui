#ifndef TRUCO_AIENGINE_HPP
#define TRUCO_AIENGINE_HPP

#include "Strategy.hpp"
#include <memory>

namespace truco {

// AIEngine manages strategy assignment and decision-making.
// Each CPU player gets a personality that stays consistent during a match.
class AIEngine {
public:
    // Personality types
    enum Personality { Balanced = 0, Aggressive = 1, Conservative = 2, Bluffer = 3 };

    // Get the singleton instance
    static AIEngine& instance();

    // Decide what action a CPU player should take
    // personalityId: 0=Balanced, 1=Aggressive, 2=Conservative, 3=Bluffer
    Decision decide(int personalityId, const GameState& state);

    // Get personality name
    static const char* personalityName(int personalityId);

private:
    AIEngine() = default;
    std::unique_ptr<Strategy> createStrategy(Personality p);
};

} // namespace truco

#endif
