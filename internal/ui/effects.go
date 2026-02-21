package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"truco-tui/internal/truco"
)

func updateTrickOverlayState(
	s truco.Snapshot,
	localIdx int,
	lastSeenTrickSeq *int,
	trickOverlayMsg *string,
	trickOverlayID *int,
	trickOverlayFrames *int,
) tea.Cmd {
	if s.LastTrickSeq <= *lastSeenTrickSeq {
		return nil
	}
	*lastSeenTrickSeq = s.LastTrickSeq
	team := 0
	if localIdx >= 0 && localIdx < len(s.Players) {
		team = s.Players[localIdx].Team
	}
	switch {
	case s.LastTrickTie:
		*trickOverlayMsg = fmt.Sprintf(tr("effect_trick_tie_format"), s.LastTrickRound)
	case s.LastTrickTeam == team:
		*trickOverlayMsg = fmt.Sprintf(tr("effect_trick_point_own_team_format"), s.LastTrickRound)
	default:
		*trickOverlayMsg = fmt.Sprintf(tr("effect_trick_point_enemy_team_format"), s.LastTrickRound)
	}
	(*trickOverlayID)++
	*trickOverlayFrames = trickOverlayAnimFrames
	return clearTrickOverlayCmd(*trickOverlayID)
}

func updatePlayAnimationState(
	s truco.Snapshot,
	lastSeenRound *int,
	lastSeenRoundSize *int,
	playAnimID *int,
	playAnimFrames *int,
	playAnimCardIndex *int,
) tea.Cmd {
	round := s.CurrentHand.Round
	roundSize := len(s.CurrentHand.RoundCards)

	if round != *lastSeenRound {
		*lastSeenRound = round
		*lastSeenRoundSize = roundSize
		*playAnimFrames = 0
		*playAnimCardIndex = -1
		return nil
	}

	if roundSize > *lastSeenRoundSize {
		*lastSeenRoundSize = roundSize
		(*playAnimID)++
		*playAnimFrames = playAnimMaxFrames
		*playAnimCardIndex = roundSize - 1
		return playAnimTickCmd(*playAnimID)
	}

	*lastSeenRoundSize = roundSize
	if roundSize == 0 {
		*playAnimFrames = 0
		*playAnimCardIndex = -1
	}
	return nil
}
