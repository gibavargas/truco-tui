package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"truco-tui/internal/truco"
)

func (m UIModel) buildStatusLine() string {
	s := m.snapshot
	localIdx := 0
	if s.CurrentPlayerIdx >= 0 {
		localIdx = s.CurrentPlayerIdx
	}
	if localIdx < 0 || localIdx >= len(s.Players) {
		localIdx = 0
	}
	localTeam := s.Players[localIdx].Team

	if m.err != nil {
		return m.renderAlert(tr("error_prefix")+m.err.Error()) + "  │  " + m.helpControls()
	}

	// Priority alerts
	if s.MatchFinished {
		return winnerStyle.Render(fmt.Sprintf(
			tr("status_match_over_format"), s.WinnerTeam+1))
	}

	if s.PendingRaiseFor == s.Players[localIdx].Team {
		raiseTo := s.PendingRaiseTo
		if raiseTo == 0 {
			raiseTo = nextStakeUI(s.CurrentHand.Stake)
		}
		raiseName := raiseLabelUI(raiseTo)
		raiseBy := safePlayerName(s.Players, s.CurrentHand.RaiseRequester)
		return m.renderAlert(fmt.Sprintf(
			tr("status_truco_response_owner_format"),
			strings.ToUpper(raiseName),
			raiseTo,
			raiseBy,
			tr("ui_role_you"),
		)) + "  │  " + m.helpControls()
	}

	if s.PendingRaiseFor != -1 {
		raiseTo := s.PendingRaiseTo
		if raiseTo == 0 {
			raiseTo = nextStakeUI(s.CurrentHand.Stake)
		}
		raiseBy := safePlayerName(s.Players, s.CurrentHand.RaiseRequester)
		waiting := tr("ui_role_opponent")
		if s.PendingRaiseFor == localTeam {
			waiting = tr("ui_role_partner")
		}
		return m.renderAlert(fmt.Sprintf(
			tr("status_truco_wait_owner_format"),
			strings.ToUpper(raiseLabelUI(raiseTo)),
			raiseTo,
			raiseBy,
			waiting,
		)) + "  │  " + m.helpControls()
	}

	turnName := s.Players[s.CurrentHand.Turn].Name
	turnInfo := fmt.Sprintf(tr("status_turn_format"), turnName)

	provisional := make([]string, 0, len(s.Players))
	for _, p := range s.Players {
		if p.ProvisionalCPU {
			provisional = append(provisional, clip(p.Name, 10))
		}
	}
	if len(provisional) > 0 {
		return turnInfo + "  │  " + tr("ui_role_cpu_prov") + ": " + strings.Join(provisional, ", ") + "  │  " + m.helpControls()
	}
	return turnInfo + "  │  " + m.helpControls()
}

func (m UIModel) helpControls() string {
	parts := []string{
		renderKeyHint("[1-3]", tr("help_play_cards_short")),
		renderKeyHint("[t]", tr("help_truco_short")),
		renderKeyHint("[a/r]", tr("help_answer_short")),
	}
	if m.activeTab == "chat" {
		parts = append(parts, renderKeyHint("[enter]", tr("help_send_short")))
	}
	if m.isOnline {
		parts = append(parts, renderKeyHint("[/host n]", tr("help_vote_host_short")))
		if m.isHost {
			parts = append(parts, renderKeyHint("[/invite n]", tr("help_invite_short")))
		} else {
			parts = append(parts, renderKeyHint("[/invite n]", tr("help_invite_request_short")))
		}
	}
	parts = append(parts, renderKeyHint("[tab]", tr("help_tab_short")))
	parts = append(parts, renderKeyHint("[q]", tr("help_quit_short")))
	return strings.Join(parts, "  ")
}

func (m UIModel) chatCommandsHint() string {
	if m.isOnline {
		if m.isHost {
			return tr("panel_chat_commands_online_host")
		}
		return tr("panel_chat_commands_online_client")
	}
	return tr("panel_chat_commands_offline")
}

func (m UIModel) renderRoleLane(width int) string {
	if width <= 0 || len(m.snapshot.Players) == 0 {
		return ""
	}
	localIdx := m.localPlayerIdx
	if m.snapshot.CurrentPlayerIdx >= 0 {
		localIdx = m.snapshot.CurrentPlayerIdx
	}
	if localIdx < 0 || localIdx >= len(m.snapshot.Players) {
		localIdx = 0
	}
	localID := m.snapshot.Players[localIdx].ID
	roles := deriveSeatRoles(m.snapshot, localIdx, m.isOnline)
	localInfo := roles[localID]
	badges := m.roleBadgeLabels(localInfo)
	if len(badges) == 0 {
		badges = []string{tr("ui_role_you")}
	}

	segments := []string{
		chipAccentStyle.Render(fmt.Sprintf("%s %s", tr("role_lane_you_prefix"), strings.Join(badges, " · "))),
	}
	if m.isOnline {
		segments = append(segments, chipStyle.Render(fmt.Sprintf("%s %s", tr("role_lane_host_prefix"), safePlayerName(m.snapshot.Players, 0))))
	}
	if m.snapshot.PendingRaiseFor != -1 {
		raiseBy := safePlayerName(m.snapshot.Players, m.snapshot.CurrentHand.RaiseRequester)
		raiseTo := m.snapshot.PendingRaiseTo
		if raiseTo == 0 {
			raiseTo = nextStakeUI(m.snapshot.CurrentHand.Stake)
		}
		segments = append(segments, alertStyle.Render(
			fmt.Sprintf(tr("role_lane_raise_pending_format"), raiseBy, strings.ToUpper(raiseLabelUI(raiseTo)), raiseTo),
		))
	}

	provisional := make([]string, 0, len(m.snapshot.Players))
	for _, p := range m.snapshot.Players {
		if p.ProvisionalCPU {
			provisional = append(provisional, clip(p.Name, 10))
		}
	}
	if len(provisional) > 0 {
		segments = append(segments, chipStyle.Render(fmt.Sprintf("%s %s", tr("ui_role_cpu_prov"), strings.Join(provisional, ","))))
	}
	return joinSegmentsWithinWidth(width, segments...)
}

func (m UIModel) roleBadgeLabels(info SeatRoleInfo) []string {
	keys := roleBadgeKeys(info)
	labels := make([]string, 0, len(keys))
	for _, k := range keys {
		labels = append(labels, strings.ToUpper(tr(k)))
	}
	return labels
}

func renderKeyHint(key, label string) string {
	return keyHintKeyStyle.Render(key) + " " + keyHintStyle.Render(label)
}

func nextStakeUI(s int) int {
	switch s {
	case 1:
		return 3
	case 3:
		return 6
	case 6:
		return 9
	case 9:
		return 12
	default:
		return s
	}
}

func raiseLabelUI(stake int) string {
	switch stake {
	case 3:
		return tr("truco_call_truco")
	case 6:
		return tr("truco_call_six")
	case 9:
		return tr("truco_call_nine")
	case 12:
		return tr("truco_call_twelve")
	default:
		return fmt.Sprintf("%d", stake)
	}
}

func stakeLadderLabel(current, pending int) string {
	steps := []int{1, 3, 6, 9, 12}
	parts := make([]string, 0, len(steps))
	for _, step := range steps {
		switch {
		case pending == step:
			parts = append(parts, fmt.Sprintf("{%d}", step))
		case current == step:
			parts = append(parts, fmt.Sprintf("[%d]", step))
		default:
			parts = append(parts, fmt.Sprintf("%d", step))
		}
	}
	return strings.Join(parts, ">")
}

func joinSegmentsWithinWidth(width int, segments ...string) string {
	if width <= 0 {
		return ""
	}
	out := ""
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		candidate := seg
		if out != "" {
			candidate = out + " " + seg
		}
		if lipgloss.Width(candidate) <= width {
			out = candidate
			continue
		}
		if out == "" {
			out = ansi.Truncate(seg, width, "")
		}
		break
	}
	return fitSingleLine(out, width)
}

func joinHorizontalWithSpacer(pos lipgloss.Position, spacer string, parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out = lipgloss.JoinHorizontal(pos, out, spacer, parts[i])
	}
	return out
}

func fitSingleLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	line = strings.ReplaceAll(line, "\n", " ")
	line = ansi.Truncate(line, width, "")
	lineW := lipgloss.Width(line)
	if lineW < width {
		line += strings.Repeat(" ", width-lineW)
	}
	return line
}

func paintBlockBackground(block string, bg lipgloss.Color) string {
	lines := strings.Split(strings.ReplaceAll(block, "\r\n", "\n"), "\n")
	lineStyle := lipgloss.NewStyle().Background(bg)
	for i := range lines {
		lines[i] = lineStyle.Render(lines[i])
	}
	return strings.Join(lines, "\n")
}

func (m UIModel) renderFlightRow(width int, compact bool) string {
	s := m.snapshot
	if m.playAnimFrames <= 0 || m.playAnimCardIndex < 0 || m.playAnimCardIndex >= len(s.CurrentHand.RoundCards) {
		return ""
	}

	pc := s.CurrentHand.RoundCards[m.playAnimCardIndex]
	localIdx := 0
	if s.CurrentPlayerIdx >= 0 {
		localIdx = s.CurrentPlayerIdx
	}
	localPlayerID := localIdx
	if localIdx >= 0 && localIdx < len(s.Players) {
		localPlayerID = s.Players[localIdx].ID
	}
	rel := relativePosition(pc.PlayerID, localPlayerID, s.Players)
	progress := playAnimationProgress(m.playAnimFrames)
	pos := lerp(animSourcePosition(rel), 0.5, progress)

	nameMax := defaultPlayedNameMax
	if compact {
		nameMax = compactPlayedNameMax
	}
	pName := clip(safePlayerName(s.Players, pc.PlayerID), nameMax)
	label := lipgloss.NewStyle().Foreground(lgCyan).Bold(true).Render(fmt.Sprintf("%s %s", animArrow(rel), pName))
	card := renderBigCard(pc.Card, false, compact)
	if pc.FaceDown {
		card = renderBigCardBack(compact)
	}
	flight := lipgloss.JoinHorizontal(lipgloss.Center, label, " ", card)
	row := lipgloss.PlaceHorizontal(width, lipgloss.Position(pos), flight, lipgloss.WithWhitespaceBackground(lgGreen))
	lead, trail := animVerticalPadding(rel, progress)
	if lead == 0 && trail == 0 {
		return row
	}
	lines := make([]string, 0, lead+1+trail)
	for i := 0; i < lead; i++ {
		lines = append(lines, fitSingleLine("", width))
	}
	lines = append(lines, row)
	for i := 0; i < trail; i++ {
		lines = append(lines, fitSingleLine("", width))
	}
	return strings.Join(lines, "\n")
}

func relativePosition(playerID, localPlayerID int, players []truco.Player) string {
	n := len(players)
	if n <= 0 {
		return "center"
	}
	playerIdx := playerIndexByID(players, playerID)
	localIdx := playerIndexByID(players, localPlayerID)
	if playerIdx == -1 || localIdx == -1 {
		playerIdx = playerID
		localIdx = localPlayerID
	}
	if playerIdx == localIdx {
		return "bottom"
	}
	dist := (playerIdx - localIdx + n) % n
	if n == 2 {
		if dist == 1 {
			return "top"
		}
		return "center"
	}
	switch dist {
	case 1:
		return "right"
	case 2:
		return "top"
	case 3:
		return "left"
	default:
		return "center"
	}
}

func animArrow(rel string) string {
	switch rel {
	case "left":
		return "→"
	case "right":
		return "←"
	case "top":
		return "↓"
	case "bottom":
		return "↑"
	default:
		return "→"
	}
}

func playerIndexByID(players []truco.Player, id int) int {
	for i := range players {
		if players[i].ID == id {
			return i
		}
	}
	return -1
}

func animSourcePosition(rel string) float64 {
	switch rel {
	case "left":
		return 0.1
	case "right":
		return 0.9
	case "top":
		return 0.5
	case "bottom":
		return 0.5
	default:
		return 0.5
	}
}

func playAnimationProgress(framesLeft int) float64 {
	total := float64(playAnimMaxFrames)
	if total <= 0 {
		return 1
	}
	progress := 1.0 - float64(framesLeft-1)/total
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return easeOutQuad(progress)
}

func easeOutQuad(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

func animVerticalPadding(rel string, progress float64) (int, int) {
	const maxShift = 2
	switch rel {
	case "top":
		lead := int(lerp(0, maxShift, progress) + 0.5)
		return lead, maxShift - lead
	case "bottom":
		lead := int(lerp(maxShift, 0, progress) + 0.5)
		return lead, maxShift - lead
	default:
		return 0, 0
	}
}

func lerp(from, to, t float64) float64 {
	return from + (to-from)*t
}

func enforceBackground(s, bgANSI string) string {
	if s == "" || bgANSI == "" {
		return s
	}
	// Alguns componentes internos renderizam reset ANSI no meio da linha.
	// Reaplicamos o fundo após cada reset para evitar "buracos" pretos na mesa.
	out := strings.ReplaceAll(s, ansiReset, ansiReset+bgANSI)
	return bgANSI + out + ansiReset
}

func (m UIModel) renderAlert(msg string) string {
	colors := []lipgloss.Color{lgRed, lipgloss.Color("208"), lgYellow, lipgloss.Color("208")}
	c := colors[m.uiFrame%len(colors)]
	return alertStyle.Foreground(c).Render(msg)
}

func cpuSpinnerFrame(frame int) string {
	frames := []string{"▶", "▷", "▹", "▸"}
	return frames[frame%len(frames)]
}

func overlayFadeColor(framesLeft int) lipgloss.Color {
	step := trickOverlayAnimFrames - framesLeft
	if step < 0 {
		step = 0
	}
	const fadeFrames = 3
	if step < fadeFrames {
		switch step {
		case 0:
			return lgGray
		case 1:
			return lipgloss.Color("248")
		default:
			return lgYellow
		}
	}
	if framesLeft <= fadeFrames {
		switch framesLeft {
		case 3:
			return lgYellow
		case 2:
			return lipgloss.Color("248")
		default:
			return lgGray
		}
	}
	return lgYellow
}

func trucoFlashPalette(framesLeft, frame int) (lipgloss.Color, lipgloss.Color, lipgloss.Color) {
	if framesLeft <= 0 {
		return lipgloss.Color("52"), lgYellow, lgRed
	}
	pulse := frame % 4
	switch pulse {
	case 0:
		return lipgloss.Color("52"), lgYellow, lgRed
	case 1:
		return lipgloss.Color("88"), lgWhite, lipgloss.Color("208")
	case 2:
		return lipgloss.Color("124"), lgWhite, lgYellow
	default:
		return lipgloss.Color("88"), lgYellow, lgRed
	}
}

func renderGhostCard(compact bool, flash bool) string {
	if compact {
		st := lipgloss.NewStyle().Background(lgWhite).Foreground(lgDim)
		if flash {
			st = st.Foreground(lgGreen).Bold(true)
		}
		return st.Render("┌··┐\n└──┘")
	}
	st := bigCardBlack.Foreground(lgDim)
	ghost := strings.Join([]string{
		"┌─────┐",
		"│··   │",
		"│  ·  │",
		"│   ··│",
		"└─────┘",
	}, "\n")
	if flash {
		st = st.Foreground(lgGreen).Bold(true)
	}
	return st.Render(ghost)
}

func cardsRevealed(total, framesLeft int) int {
	if total <= 0 {
		return 0
	}
	if framesLeft <= 0 {
		return total
	}
	elapsed := dealAnimFrames - framesLeft + 1
	if elapsed < 1 {
		elapsed = 1
	}
	step := maxInt(1, dealAnimFrames/total)
	reveal := (elapsed + step - 1) / step
	if reveal < 1 {
		reveal = 1
	}
	if reveal > total {
		reveal = total
	}
	return reveal
}

func formatTrickHistory(results []int) string {
	if len(results) == 0 {
		return tr("table_trick_history_empty")
	}
	parts := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		if i >= len(results) {
			parts = append(parts, fmt.Sprintf("%s%d: ...", tr("table_trick_prefix"), i+1))
			continue
		}
		switch results[i] {
		case 0:
			parts = append(parts, fmt.Sprintf("%s%d: T1 ✓", tr("table_trick_prefix"), i+1))
		case 1:
			parts = append(parts, fmt.Sprintf("%s%d: T2 ✓", tr("table_trick_prefix"), i+1))
		default:
			parts = append(parts, fmt.Sprintf("%s%d: %s", tr("table_trick_prefix"), i+1, tr("table_tie_word")))
		}
	}
	return strings.Join(parts, " | ")
}

func scoreHistoryFromLogs(logs []string, n int) []string {
	if n <= 0 {
		return nil
	}
	markers := append(
		allTranslationsForKey("log_hand_ended_prefix"),
		allTranslationsForKey("log_match_ended_prefix")...,
	)
	events := make([]string, 0, n)
	for i := len(logs) - 1; i >= 0 && len(events) < n; i-- {
		line := logs[i]
		for _, marker := range markers {
			if strings.Contains(line, marker) {
				events = append(events, clip(line, 48))
				break
			}
		}
	}
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}
	return events
}
