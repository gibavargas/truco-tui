#include "truco_ai.h"
#include "AIEngine.hpp"
#include "GameState.hpp"
#include <vector>

// Convert C API types to C++ enums
static truco::Suit toSuit(int s) {
    return static_cast<truco::Suit>(s & 3);
}

static truco::Rank toRank(int r) {
    return static_cast<truco::Rank>(r % 10);
}

extern "C" TrucoDecision truco_ai_decide(int personality, const TrucoAIState* s) {
    truco::GameState state;

    // Hand state
    state.manilha = toRank(s->manilha);
    state.vira = {toSuit(s->vira_suit), toRank(s->vira_rank)};
    state.stake = s->stake;
    state.round = s->round;
    state.ourTrickWins = s->our_trick_wins;
    state.theirTrickWins = s->their_trick_wins;
    state.pendingRaiseFor = s->pending_raise_for;
    state.trucoByTeam0 = (s->truco_by_team == 0);
    state.trucoByTeam1 = (s->truco_by_team == 1);

    // Match state
    state.ourMatchScore = s->our_match_score;
    state.theirMatchScore = s->their_match_score;
    state.matchTarget = s->match_target;
    state.matchFinished = s->match_finished != 0;

    // Player state
    state.myPlayerId = s->my_player_id;
    state.myTeam = s->my_team;
    state.turnPlayer = s->turn_player;
    state.canAskTruco = s->can_ask_truco != 0;

    // My hand
    state.myHand.reserve(s->hand_len);
    for (int i = 0; i < s->hand_len; i++) {
        state.myHand.push_back({toSuit(s->hand_suits[i]), toRank(s->hand_ranks[i])});
    }

    // Table cards
    state.tableCards.reserve(s->table_len);
    for (int i = 0; i < s->table_len; i++) {
        truco::PlayedCardInfo pci;
        pci.card = {toSuit(s->table_suits[i]), toRank(s->table_ranks[i])};
        pci.playerId = s->table_player_ids[i];
        pci.team = s->table_teams[i];
        pci.faceDown = false;
        state.tableCards.push_back(pci);
    }

    // Decide via AI engine
    truco::Decision d = truco::AIEngine::instance().decide(personality, state);

    // Convert back to C struct
    TrucoDecision result;
    switch (d.kind) {
        case truco::Decision::PLAY:          result.kind = 0; break;
        case truco::Decision::PLAY_FACEDOWN: result.kind = 1; break;
        case truco::Decision::ASK_TRUCO:     result.kind = 2; break;
        case truco::Decision::RAISE:         result.kind = 3; break;
        case truco::Decision::ACCEPT:        result.kind = 4; break;
        case truco::Decision::REFUSE:        result.kind = 5; break;
        default:                             result.kind = 0; break;
    }
    result.card_index = d.cardIndex;
    return result;
}

extern "C" const char* truco_ai_personality_name(int personality) {
    return truco::AIEngine::personalityName(personality);
}
