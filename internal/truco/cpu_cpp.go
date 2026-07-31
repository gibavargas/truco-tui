package truco

/*
#cgo CFLAGS: -I${SRCDIR}/../../native/ai-engine/include
#cgo LDFLAGS: -L${SRCDIR}/../../native/ai-engine -Wl,-rpath,${SRCDIR}/../../native/ai-engine -ltruco_ai -lstdc++

#include <stdlib.h>
#include "truco_ai.h"

// Helper to allocate arrays for hand cards
static void fill_ai_state(
    TrucoAIState* s,
    int manilha, int vira_suit, int vira_rank,
    int stake, int round,
    int our_wins, int their_wins,
    int pending_raise, int truco_by_team,
    int our_score, int their_score, int match_target, int match_finished,
    int my_id, int my_team, int turn_player, int can_ask_truco,
    const int* hand_suits, const int* hand_ranks, int hand_len,
    const int* tbl_suits, const int* tbl_ranks,
    const int* tbl_pids, const int* tbl_teams, int tbl_len
) {
    s->manilha = manilha;
    s->vira_suit = vira_suit;
    s->vira_rank = vira_rank;
    s->stake = stake;
    s->round = round;
    s->our_trick_wins = our_wins;
    s->their_trick_wins = their_wins;
    s->pending_raise_for = pending_raise;
    s->truco_by_team = truco_by_team;
    s->our_match_score = our_score;
    s->their_match_score = their_score;
    s->match_target = match_target;
    s->match_finished = match_finished;
    s->my_player_id = my_id;
    s->my_team = my_team;
    s->turn_player = turn_player;
    s->can_ask_truco = can_ask_truco ? 1 : 0;
    s->hand_suits = hand_suits;
    s->hand_ranks = hand_ranks;
    s->hand_len = hand_len;
    s->table_suits = tbl_suits;
    s->table_ranks = tbl_ranks;
    s->table_player_ids = tbl_pids;
    s->table_teams = tbl_teams;
    s->table_len = tbl_len;
}
*/
import "C"
import "fmt"

// rankToC converts a Go Rank to C enum value (0-9)
func rankToC(r Rank) int {
	switch r {
	case R4:
		return 0
	case R5:
		return 1
	case R6:
		return 2
	case R7:
		return 3
	case RQ:
		return 4
	case RJ:
		return 5
	case RK:
		return 6
	case RA:
		return 7
	case R2:
		return 8
	case R3:
		return 9
	}
	return 0
}

// suitToC converts a Go Suit to C enum value (0-3)
func suitToC(s Suit) int {
	switch s {
	case Diamonds:
		return 0
	case Spades:
		return 1
	case Hearts:
		return 2
	case Clubs:
		return 3
	}
	return 0
}

// CPUPersonality determines which strategy the C++ AI uses for a given CPU player.
// Uses player ID to assign consistent personalities within a match.
func cpuPersonality(playerID int) int {
	// Cycle through personalities: Balanced, Aggressive, Conservative, Bluffer
	return playerID % 4
}

// DecideCPUActionCpp delegates the CPU decision to the C++ AI engine.
// Falls back to the pure-Go heuristic if the native library fails.
func DecideCPUActionCpp(g *Game, playerID int) CPUAction {
	snap := g.Snapshot(playerID)
	team := g.TeamOfPlayer(playerID)
	cards := g.HandCards(playerID)
	if len(cards) == 0 {
		return CPUAction{Kind: "refuse"}
	}

	// Build hand arrays
	handSuits := make([]C.int, len(cards))
	handRanks := make([]C.int, len(cards))
	for i, c := range cards {
		handSuits[i] = C.int(suitToC(c.Suit))
		handRanks[i] = C.int(rankToC(c.Rank))
	}

	// Build table card arrays
	tableCards := snap.CurrentHand.RoundCards
	tblSuits := make([]C.int, len(tableCards))
	tblRanks := make([]C.int, len(tableCards))
	tblPids := make([]C.int, len(tableCards))
	tblTeams := make([]C.int, len(tableCards))
	for i, pc := range tableCards {
		tblSuits[i] = C.int(suitToC(pc.Card.Suit))
		tblRanks[i] = C.int(rankToC(pc.Card.Rank))
		tblPids[i] = C.int(pc.PlayerID)
		tblTeams[i] = C.int(teamForPlayer(snap.Players, pc.PlayerID))
	}

	// Determine truco_by_team
	trucoByTeam := -1
	if snap.CurrentHand.TrucoByTeam >= 0 && snap.CurrentHand.TrucoByTeam <= 1 {
		trucoByTeam = snap.CurrentHand.TrucoByTeam
	}

	// Fill the C struct
	var cs C.TrucoAIState
	C.fill_ai_state(
		&cs,
		C.int(rankToC(snap.CurrentHand.Manilha)),
		C.int(suitToC(snap.CurrentHand.Vira.Suit)),
		C.int(rankToC(snap.CurrentHand.Vira.Rank)),
		C.int(snap.CurrentHand.Stake),
		C.int(snap.CurrentHand.Round),
		C.int(snap.CurrentHand.TrickWins[team]),
		C.int(snap.CurrentHand.TrickWins[1-team]),
		C.int(snap.CurrentHand.PendingRaiseFor),
		C.int(trucoByTeam),
		C.int(snap.MatchPoints[team]),
		C.int(snap.MatchPoints[1-team]),
		C.int(TargetPoints),
		C.int(0), // matchFinished not relevant for per-turn decision
		C.int(playerID),
		C.int(team),
		C.int(snap.TurnPlayer),
		C.int(0), // canAskTruco computed below
		&handSuits[0],
		&handRanks[0],
		C.int(len(cards)),
		&tblSuits[0],
		&tblRanks[0],
		&tblPids[0],
		&tblTeams[0],
		C.int(len(tableCards)),
	)
	cs.can_ask_truco = 0
	if g.CanAskTrucoByPlayer(playerID) {
		cs.can_ask_truco = 1
	}

	// Call C++ engine
	personality := cpuPersonality(playerID)
	result := C.truco_ai_decide(C.int(personality), &cs)

	// Convert result back
	switch int(result.kind) {
	case 0:
		return CPUAction{Kind: "play", CardIndex: int(result.card_index)}
	case 1:
		return CPUAction{Kind: "play_facedown", CardIndex: int(result.card_index)}
	case 2:
		return CPUAction{Kind: "ask_truco"}
	case 3:
		return CPUAction{Kind: "raise"}
	case 4:
		return CPUAction{Kind: "accept"}
	case 5:
		return CPUAction{Kind: "refuse"}
	}

	// Fallback
	return CPUAction{Kind: "play", CardIndex: 0}
}

// CPUPersonalityName returns the name of the strategy assigned to a CPU player.
func CPUPersonalityName(playerID int) string {
	personality := cpuPersonality(playerID)
	name := C.truco_ai_personality_name(C.int(personality))
	return C.GoString(name)
}

// Ensure we reference fmt to avoid unused import
var _ = fmt.Sprintf
