package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"truco-tui/internal/truco"
)

// visualState concentra estados/efeitos temporários de UI usados por
// UI local e online para evitar duplicação de lógica.
type visualState struct {
	lastSeenTrickSeq int
	trickOverlayMsg  string
	trickOverlayID   int

	lastSeenRound      int
	lastSeenRoundSize  int
	playAnimID         int
	playAnimFrames     int
	playAnimCardIndex  int
	ghostCardSlot      int
	ghostFrames        int
	inputFlashSlot     int
	inputFlashFrames   int
	tableSweepFrames   int
	trucoFlashFrames   int
	dealFrames         int
	errClearID         int
	uiFrame            int
	trickOverlayFrames int
}

func newVisualState(s truco.Snapshot) visualState {
	return visualState{
		lastSeenTrickSeq:  s.LastTrickSeq,
		lastSeenRound:     s.CurrentHand.Round,
		lastSeenRoundSize: len(s.CurrentHand.RoundCards),
		playAnimCardIndex: -1,
		ghostCardSlot:     -1,
		inputFlashSlot:    -1,
		dealFrames:        dealAnimFrames,
	}
}

func (v *visualState) onUITick() {
	v.uiFrame++
	if v.ghostFrames > 0 {
		v.ghostFrames--
		if v.ghostFrames == 0 {
			v.ghostCardSlot = -1
		}
	}
	if v.inputFlashFrames > 0 {
		v.inputFlashFrames--
		if v.inputFlashFrames == 0 {
			v.inputFlashSlot = -1
		}
	}
	if v.tableSweepFrames > 0 {
		v.tableSweepFrames--
	}
	if v.trucoFlashFrames > 0 {
		v.trucoFlashFrames--
	}
	if v.dealFrames > 0 {
		v.dealFrames--
	}
	if v.trickOverlayFrames > 0 {
		v.trickOverlayFrames--
	}
}

func (v *visualState) onCardAccepted(idx int) {
	v.ghostCardSlot = idx
	v.ghostFrames = ghostAnimFrames
	v.inputFlashSlot = idx
	v.inputFlashFrames = inputConfirmFrames
}

func (v *visualState) applySnapshotVisualTransitions(prev, next truco.Snapshot) {
	if prev.PendingRaiseFor == -1 && next.PendingRaiseFor != -1 {
		v.trucoFlashFrames = trucoFlashAnimFrames
	}
	if next.LastTrickSeq > prev.LastTrickSeq {
		v.tableSweepFrames = trickSweepAnimFrames
	}
	if prev.CurrentHand.Dealer != next.CurrentHand.Dealer && next.CurrentHand.Round == 1 && len(next.CurrentHand.RoundCards) == 0 {
		v.dealFrames = dealAnimFrames
		v.ghostFrames = 0
		v.ghostCardSlot = -1
		v.inputFlashFrames = 0
		v.inputFlashSlot = -1
	}
}

func (v *visualState) updateTrickOverlay(s truco.Snapshot, localIdx int) tea.Cmd {
	return updateTrickOverlayState(
		s,
		localIdx,
		&v.lastSeenTrickSeq,
		&v.trickOverlayMsg,
		&v.trickOverlayID,
		&v.trickOverlayFrames,
	)
}

func (v *visualState) updatePlayAnimation(s truco.Snapshot) tea.Cmd {
	return updatePlayAnimationState(
		s,
		&v.lastSeenRound,
		&v.lastSeenRoundSize,
		&v.playAnimID,
		&v.playAnimFrames,
		&v.playAnimCardIndex,
	)
}

func (v *visualState) onPlayAnimTick(id int) tea.Cmd {
	if id == v.playAnimID && v.playAnimFrames > 0 {
		v.playAnimFrames--
		if v.playAnimFrames > 0 {
			return playAnimTickCmd(v.playAnimID)
		}
		v.playAnimCardIndex = -1
	}
	return nil
}

func (v *visualState) onClearTrickOverlay(id int) {
	if id == v.trickOverlayID {
		v.trickOverlayMsg = ""
	}
}

func (v *visualState) onClearError(id int, errPtr *error) {
	if id == v.errClearID && errPtr != nil {
		*errPtr = nil
	}
}
