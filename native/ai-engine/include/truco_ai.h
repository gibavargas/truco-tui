#ifndef TRUCO_AI_H
#define TRUCO_AI_H

#ifdef __cplusplus
extern "C" {
#endif

/*
 * C API for the Truco AI Engine (used via CGO from Go).
 * All structs are POD — no C++ features leak across the boundary.
 */

/* Result of a CPU decision */
typedef struct {
    int kind;        /* 0=play, 1=play_facedown, 2=ask_truco, 3=raise, 4=accept, 5=refuse */
    int card_index;  /* card index in hand, or -1 */
} TrucoDecision;

/*
 * State passed from Go. Mirrors the Go Snapshot.
 * Arrays use pointer + length (C-style).
 */
typedef struct {
    /* Hand state */
    int manilha;       /* Rank enum value 0-9 */
    int vira_suit;     /* Suit enum value 0-3 */
    int vira_rank;     /* Rank enum value 0-9 */
    int stake;
    int round;
    int our_trick_wins;
    int their_trick_wins;
    int pending_raise_for;  /* -1 if none */
    int truco_by_team;      /* -1 if none, 0 or 1 */

    /* Match state */
    int our_match_score;
    int their_match_score;
    int match_target;
    int match_finished;

    /* Player state */
    int my_player_id;
    int my_team;
    int turn_player;
    int can_ask_truco;

    /* My hand (up to 3 cards) */
    const int* hand_suits;   /* array of suit values, length = hand_len */
    const int* hand_ranks;   /* array of rank values */
    int hand_len;

    /* Table cards (this trick) */
    const int* table_suits;
    const int* table_ranks;
    const int* table_player_ids;
    const int* table_teams;
    int table_len;
} TrucoAIState;

/* Main entry point: decide what action the CPU should take */
TrucoDecision truco_ai_decide(int personality, const TrucoAIState* state);

/* Get personality name (for debugging) */
const char* truco_ai_personality_name(int personality);

#ifdef __cplusplus
}
#endif

#endif /* TRUCO_AI_H */
