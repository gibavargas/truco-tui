package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"truco-tui/internal/truco"
)

func newUIModelForTest(t *testing.T) UIModel {
	t.Helper()
	g, err := truco.NewGame([]string{"Jogador-1", "Jogador-2"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame() error: %v", err)
	}
	m := InitialModel(g)
	m.width = 100
	m.height = 35
	return m
}

func newUIModelForFourPlayers(t *testing.T) UIModel {
	t.Helper()
	g, err := truco.NewGame(
		[]string{"joao", "CPU-2", "CPU-3", "CPU-4"},
		[]bool{false, true, true, true},
	)
	if err != nil {
		t.Fatalf("NewGame(4p) error: %v", err)
	}
	m := InitialModel(g)
	m.width = 113
	m.height = 41
	return m
}

func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestTabCycle(t *testing.T) {
	m := newUIModelForTest(t)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(UIModel)
	if m.activeTab != "chat" {
		t.Fatalf("activeTab after 1st TAB = %q, want %q", m.activeTab, "chat")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(UIModel)
	if m.activeTab != "log" {
		t.Fatalf("activeTab after 2nd TAB = %q, want %q", m.activeTab, "log")
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(UIModel)
	if m.activeTab != "mesa" {
		t.Fatalf("activeTab after 3rd TAB = %q, want %q", m.activeTab, "mesa")
	}
}

func TestInvalidPlayShowsErrorInStatus(t *testing.T) {
	m := newUIModelForTest(t)

	// No estado inicial de NewGame, a mão começa no jogador 2 (seat 1),
	// então seat 0 jogar carta deve gerar erro de turno.
	updated, _ := m.Update(keyRune('1'))
	m = updated.(UIModel)

	if m.err == nil {
		t.Fatalf("expected m.err after invalid play, got nil")
	}
	if !strings.Contains(m.err.Error(), "não é a vez") {
		t.Fatalf("unexpected error message: %q", m.err.Error())
	}

	view := m.View()
	if !strings.Contains(view, "Erro:") {
		t.Fatalf("view should contain status error marker; view=%q", view)
	}
}

func TestViewReflectsSelectedTabPanel(t *testing.T) {
	m := newUIModelForTest(t)

	// chat
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(UIModel)
	view := m.View()
	if !strings.Contains(view, "CHAT (offline)") {
		t.Fatalf("chat panel not rendered; view=%q", view)
	}

	// log
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(UIModel)
	view = m.View()
	if !strings.Contains(view, "LOG DA PARTIDA") {
		t.Fatalf("log panel not rendered; view=%q", view)
	}
}

func TestChatInputCursorEditing(t *testing.T) {
	m := newUIModelForTest(t)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(UIModel)

	updated, _ = m.Update(keyRune('a'))
	m = updated.(UIModel)
	updated, _ = m.Update(keyRune('b'))
	m = updated.(UIModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(UIModel)
	updated, _ = m.Update(keyRune('X'))
	m = updated.(UIModel)

	if m.chatInput != "aXb" {
		t.Fatalf("chatInput = %q, want %q", m.chatInput, "aXb")
	}
	if m.chatCursor != 2 {
		t.Fatalf("chatCursor = %d, want %d", m.chatCursor, 2)
	}
}

func TestChatEnterAppendsLocalMessage(t *testing.T) {
	m := newUIModelForTest(t)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(UIModel)
	updated, _ = m.Update(keyRune('o'))
	m = updated.(UIModel)
	updated, _ = m.Update(keyRune('i'))
	m = updated.(UIModel)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(UIModel)

	if m.chatInput != "" {
		t.Fatalf("chatInput after Enter = %q, want empty", m.chatInput)
	}
	if len(m.chatLog) == 0 {
		t.Fatalf("expected chat log message after Enter")
	}
	last := m.chatLog[len(m.chatLog)-1]
	if !strings.Contains(last, "Jogador-1: oi") {
		t.Fatalf("unexpected chat message %q", last)
	}
}

func TestGameplayShortcutDoesNotPlayWhileTypingInChat(t *testing.T) {
	m := newUIModelForTest(t)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(UIModel)
	updated, _ = m.Update(keyRune('1'))
	m = updated.(UIModel)

	if m.chatInput != "1" {
		t.Fatalf("chatInput = %q, want %q", m.chatInput, "1")
	}
	if m.err != nil {
		t.Fatalf("did not expect gameplay error while in chat, got %v", m.err)
	}
}

func TestScoreHistoryFromLogsMatchesAllLocales(t *testing.T) {
	logs := []string{
		"abc",
		"Mão encerrada: time 1 ganhou 1 ponto(s).",
		"foo",
		"Hand ended: team 2 gained 3 points.",
		"Partida encerrada: time 1 venceu!",
		"Match ended: team 2 won!",
	}
	got := scoreHistoryFromLogs(logs, 3)
	if len(got) != 3 {
		t.Fatalf("scoreHistoryFromLogs len=%d, want 3 (got=%v)", len(got), got)
	}
}

func TestLeadingCardIndexes(t *testing.T) {
	manilha := truco.RA
	played := []truco.PlayedCard{
		{PlayerID: 0, Card: truco.Card{Rank: truco.R7, Suit: truco.Spades}},
		{PlayerID: 1, Card: truco.Card{Rank: truco.R2, Suit: truco.Hearts}},
		{PlayerID: 2, Card: truco.Card{Rank: truco.RA, Suit: truco.Clubs}}, // manilha, strongest
	}

	got := leadingCardIndexes(played, manilha)
	if !got[2] {
		t.Fatalf("expected index 2 to be leading")
	}
	if got[0] || got[1] {
		t.Fatalf("unexpected leading index map: %+v", got)
	}
}

func TestCompactLayoutRendersAt80x24(t *testing.T) {
	m := newUIModelForTest(t)
	m.width = 80
	m.height = 24
	view := m.View()
	if !strings.Contains(view, "TRUCO PAULISTA") {
		t.Fatalf("compact view missing header; view=%q", view)
	}
	if !strings.Contains(view, "Abas:") {
		t.Fatalf("compact view missing tabs; view=%q", view)
	}
}

func TestComputeLayoutProfiles(t *testing.T) {
	lpCompact := computeLayout(80, 24)
	if !lpCompact.compact {
		t.Fatalf("80x24 should be compact profile")
	}
	if lpCompact.panelLines != 1 {
		t.Fatalf("compact panelLines = %d, want 1", lpCompact.panelLines)
	}

	lpLarge := computeLayout(113, 41)
	if lpLarge.compact {
		t.Fatalf("113x41 should not be compact profile")
	}
	if lpLarge.panelLines != 3 {
		t.Fatalf("non-compact panelLines = %d, want 3", lpLarge.panelLines)
	}
}

func TestViewFitsBoundsAcrossTerminalSizes(t *testing.T) {
	m := newUIModelForFourPlayers(t)
	sizes := []struct {
		w int
		h int
	}{
		{80, 24},
		{90, 28},
		{113, 41},
		{140, 50},
	}

	for _, sz := range sizes {
		assertViewFitsBounds(t, &m, sz.w, sz.h)
	}
}

func TestViewShowsCurrentTurnClues(t *testing.T) {
	m := newUIModelForFourPlayers(t)
	view := m.View()
	if !strings.Contains(view, "Vez:") {
		t.Fatalf("turn badge is missing: view=%q", view)
	}
	if !strings.Contains(view, "▶") {
		t.Fatalf("active-player marker is missing: view=%q", view)
	}
}

func TestViewHighlightsLeadingRoundCard(t *testing.T) {
	m := newUIModelForFourPlayers(t)
	m.snapshot.CurrentHand.Manilha = truco.RA
	m.snapshot.CurrentHand.RoundCards = []truco.PlayedCard{
		{PlayerID: 0, Card: truco.Card{Rank: truco.R7, Suit: truco.Spades}},
		{PlayerID: 1, Card: truco.Card{Rank: truco.RA, Suit: truco.Clubs}}, // strongest
	}

	view := m.View()
	if !strings.Contains(view, "★") {
		t.Fatalf("leading-card marker is missing: view=%q", view)
	}
}

func TestShowsTrickEndOverlay(t *testing.T) {
	g, err := truco.NewGame([]string{"Jogador-1", "Jogador-2"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame() error: %v", err)
	}
	m := InitialModel(g)
	m.width = 100
	m.height = 35

	s := g.Snapshot(0)
	first := s.TurnPlayer
	second := 1 - first
	if err := g.PlayCard(first, 0); err != nil {
		t.Fatalf("PlayCard(first): %v", err)
	}
	if err := g.PlayCard(second, 0); err != nil {
		t.Fatalf("PlayCard(second): %v", err)
	}

	updated, _ := m.Update(syncMsg{snapshot: g.Snapshot(0)})
	m = updated.(UIModel)
	if m.trickOverlayMsg == "" {
		t.Fatalf("expected trick overlay active")
	}
	view := m.View()
	if !strings.Contains(view, "Fim da vaza") &&
		!strings.Contains(view, "Recolhendo vaza...") &&
		!strings.Contains(view, "Recolhendo vaza para") {
		t.Fatalf("expected trick end transition overlay; view=%q", view)
	}
	if strings.Contains(view, "Recolhendo vaza...") || strings.Contains(view, "Recolhendo vaza para") {
		return
	}

	if m.snapshot.LastTrickTie {
		if !strings.Contains(view, "empatou") {
			t.Fatalf("expected tie message; view=%q", view)
		}
		return
	}
	localTeam := m.snapshot.Players[m.localPlayerIdx].Team
	if m.snapshot.LastTrickTeam == localTeam {
		if !strings.Contains(view, "sua equipe") {
			t.Fatalf("expected own-team message; view=%q", view)
		}
		return
	}
	if !strings.Contains(view, "equipe adversaria") {
		t.Fatalf("expected opponent-team message; view=%q", view)
	}
}

func TestTrucoKeyRaisesWhenLocalTeamMustRespond(t *testing.T) {
	g, err := truco.NewGame([]string{"Jogador-1", "Jogador-2"}, []bool{false, false})
	if err != nil {
		t.Fatalf("NewGame() error: %v", err)
	}
	m := InitialModel(g)
	m.width = 100
	m.height = 35

	// Local (seat 0) precisa responder pedido do adversário e [t] deve subir.
	turn := g.Snapshot(0).TurnPlayer
	if turn != 1 {
		if err := g.PlayCard(turn, 0); err != nil {
			t.Fatalf("PlayCard(turn) to rotate turn: %v", err)
		}
	}
	if got := g.Snapshot(0).TurnPlayer; got != 1 {
		t.Fatalf("expected opponent turn before ask, got %d", got)
	}
	if err := g.AskTruco(1); err != nil {
		t.Fatalf("AskTruco(opponent): %v", err)
	}

	updated, _ := m.Update(syncMsg{snapshot: g.Snapshot(0)})
	m = updated.(UIModel)
	updated, _ = m.Update(keyRune('t'))
	m = updated.(UIModel)

	s := g.Snapshot(0)
	if s.CurrentHand.Stake != 3 {
		t.Fatalf("stake after raise = %d, want 3", s.CurrentHand.Stake)
	}
	if s.PendingRaiseFor != 1 {
		t.Fatalf("pending team after raise = %d, want 1", s.PendingRaiseFor)
	}
	if s.PendingRaiseTo != 6 {
		t.Fatalf("pending raise-to = %d, want 6", s.PendingRaiseTo)
	}
	if m.err != nil {
		t.Fatalf("did not expect error, got %v", m.err)
	}
}

func TestGameplayInputIsBlockedDuringTrickOverlay(t *testing.T) {
	m := newUIModelForTest(t)
	m.trickOverlayMsg = "Vaza 1: ponto da sua equipe."

	updated, _ := m.Update(keyRune('1'))
	m = updated.(UIModel)

	if m.err != nil {
		t.Fatalf("expected no gameplay error while overlay is active, got: %v", m.err)
	}
}

func TestViewFitsAtLeast100ResolutionCombinations(t *testing.T) {
	m := newUIModelForFourPlayers(t)
	widths := []int{80, 84, 88, 92, 96, 100, 104, 108, 112, 116}
	heights := []int{24, 26, 28, 30, 32, 34, 36, 38, 40, 42}

	count := 0
	for _, w := range widths {
		for _, h := range heights {
			assertViewFitsBounds(t, &m, w, h)
			count++
		}
	}
	if count < 100 {
		t.Fatalf("expected at least 100 resolution combinations, got %d", count)
	}
}

func assertViewFitsBounds(t *testing.T, m *UIModel, w, h int) {
	t.Helper()
	m.width = w
	m.height = h
	view := m.View()
	lines := strings.Split(view, "\n")

	if len(lines) != h {
		t.Fatalf("view height mismatch at %dx%d: got %d lines", w, h, len(lines))
	}
	for i, line := range lines {
		if got := lipgloss.Width(line); got != w {
			t.Fatalf("line %d width mismatch at %dx%d: width=%d", i+1, w, h, got)
		}
	}
}
