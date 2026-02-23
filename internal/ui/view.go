package ui

import (
	"fmt"
	"os"
	"strings"

	"truco-tui/internal/truco"

	"github.com/charmbracelet/lipgloss"
)

const (
	ansiReset = "\x1b[0m"
)

type uiPalette struct {
	FeltEdge       lipgloss.Color
	FeltMiddle     lipgloss.Color
	FeltCenter     lipgloss.Color
	TextPrimary    lipgloss.Color
	TextMuted      lipgloss.Color
	TextDim        lipgloss.Color
	TextDanger     lipgloss.Color
	TextWarning    lipgloss.Color
	TextAccent     lipgloss.Color
	TextCardDark   lipgloss.Color
	CardBackground lipgloss.Color
	FrameBorder    lipgloss.Color
	ScoreBarBG     lipgloss.Color
	ScoreChipBG    lipgloss.Color
	AccentChipBG   lipgloss.Color
	HelpBarBG      lipgloss.Color
	TabBarBG       lipgloss.Color
	TabActiveBG    lipgloss.Color
	TurnBadgeBG    lipgloss.Color
	TeamOne        lipgloss.Color
	TeamTwo        lipgloss.Color
	GridBorder     lipgloss.Color
}

var (
	darkPalette = uiPalette{
		FeltEdge:       lipgloss.Color("22"),
		FeltMiddle:     lipgloss.Color("28"),
		FeltCenter:     lipgloss.Color("29"),
		TextPrimary:    lipgloss.Color("255"),
		TextMuted:      lipgloss.Color("240"),
		TextDim:        lipgloss.Color("245"),
		TextDanger:     lipgloss.Color("196"),
		TextWarning:    lipgloss.Color("220"),
		TextAccent:     lipgloss.Color("51"),
		TextCardDark:   lipgloss.Color("232"),
		CardBackground: lipgloss.Color("255"),
		FrameBorder:    lipgloss.Color("57"),
		ScoreBarBG:     lipgloss.Color("235"),
		ScoreChipBG:    lipgloss.Color("238"),
		AccentChipBG:   lipgloss.Color("17"),
		HelpBarBG:      lipgloss.Color("236"),
		TabBarBG:       lipgloss.Color("237"),
		TabActiveBG:    lipgloss.Color("19"),
		TurnBadgeBG:    lipgloss.Color("33"),
		TeamOne:        lipgloss.Color("45"),
		TeamTwo:        lipgloss.Color("214"),
		GridBorder:     lipgloss.Color("60"),
	}
	lightPalette = uiPalette{
		FeltEdge:       lipgloss.Color("120"),
		FeltMiddle:     lipgloss.Color("114"),
		FeltCenter:     lipgloss.Color("108"),
		TextPrimary:    lipgloss.Color("16"),
		TextMuted:      lipgloss.Color("240"),
		TextDim:        lipgloss.Color("242"),
		TextDanger:     lipgloss.Color("124"),
		TextWarning:    lipgloss.Color("130"),
		TextAccent:     lipgloss.Color("24"),
		TextCardDark:   lipgloss.Color("16"),
		CardBackground: lipgloss.Color("255"),
		FrameBorder:    lipgloss.Color("24"),
		ScoreBarBG:     lipgloss.Color("188"),
		ScoreChipBG:    lipgloss.Color("153"),
		AccentChipBG:   lipgloss.Color("117"),
		HelpBarBG:      lipgloss.Color("187"),
		TabBarBG:       lipgloss.Color("188"),
		TabActiveBG:    lipgloss.Color("68"),
		TurnBadgeBG:    lipgloss.Color("68"),
		TeamOne:        lipgloss.Color("25"),
		TeamTwo:        lipgloss.Color("160"),
		GridBorder:     lipgloss.Color("24"),
	}
)

// ── Colour Palette ──────────────────────────────────────────────────────────
var (
	lgGreen  lipgloss.Color
	lgWhite  lipgloss.Color
	lgRed    lipgloss.Color
	lgBlack  lipgloss.Color
	lgYellow lipgloss.Color
	lgCyan   lipgloss.Color
	lgGray   lipgloss.Color
	lgPurple lipgloss.Color
	lgDim    lipgloss.Color

	teamOneColor lipgloss.Color
	teamTwoColor lipgloss.Color
)

// ── Reusable Styles ─────────────────────────────────────────────────────────
var (
	frameBorderStyle        lipgloss.Style
	bigCardBlack            lipgloss.Style
	bigCardCompact          lipgloss.Style
	miniBack                lipgloss.Style
	miniBackCompact         lipgloss.Style
	headerStyle             lipgloss.Style
	scoreStyle              lipgloss.Style
	chipStyle               lipgloss.Style
	chipAccentStyle         lipgloss.Style
	chipTeamOneStyle        lipgloss.Style
	chipTeamTwoStyle        lipgloss.Style
	helpStyle               lipgloss.Style
	tabBarStyle             lipgloss.Style
	tabActiveStyle          lipgloss.Style
	tabInactiveStyle        lipgloss.Style
	nameLabelStyle          lipgloss.Style
	nameTeamOneStyle        lipgloss.Style
	nameTeamTwoStyle        lipgloss.Style
	indexStyle              lipgloss.Style
	viraLabelStyle          lipgloss.Style
	roundLabelStyle         lipgloss.Style
	alertStyle              lipgloss.Style
	winnerStyle             lipgloss.Style
	playerBoxStyle          lipgloss.Style
	activePlayerBoxStyle    lipgloss.Style
	activeNameStyle         lipgloss.Style
	centerPanelStyle        lipgloss.Style
	centerPanelCompactStyle lipgloss.Style
	turnBadgeStyle          lipgloss.Style
	leadingCardLabelStyle   lipgloss.Style
	keyHintStyle            lipgloss.Style
	keyHintKeyStyle         lipgloss.Style
	tableGridCellStyle      lipgloss.Style
)

var currentThemeKey string
var ansiFeltBG string

func ensureThemeStyles() {
	themeKey := strings.ToLower(strings.TrimSpace(os.Getenv("TRUCO_THEME")))
	if themeKey == "" {
		themeKey = "dark"
	}
	if themeKey == currentThemeKey {
		return
	}
	currentThemeKey = themeKey
	p := darkPalette
	if themeKey == "light" {
		p = lightPalette
	}
	applyThemeStyles(p)
}

func applyThemeStyles(p uiPalette) {
	lgGreen = p.FeltEdge
	lgWhite = p.TextPrimary
	lgRed = p.TextDanger
	lgBlack = p.TextCardDark
	lgYellow = p.TextWarning
	lgCyan = p.TextAccent
	lgGray = p.TextMuted
	lgPurple = p.FrameBorder
	lgDim = p.TextDim

	feltEdgeColor = p.FeltEdge
	feltMiddleColor = p.FeltMiddle
	feltCenterColor = p.FeltCenter
	ansiFeltBG = colorTo256BG(p.FeltEdge)

	frameBorderStyle = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(p.FrameBorder)
	bigCardBlack = lipgloss.NewStyle().Background(p.CardBackground).Foreground(p.TextCardDark)
	bigCardCompact = lipgloss.NewStyle().Background(p.CardBackground).Foreground(p.TextCardDark)
	miniBack = lipgloss.NewStyle().
		Background(p.FeltMiddle).
		Foreground(p.TextAccent).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.TextAccent).
		Width(4).
		Height(1).
		Align(lipgloss.Center, lipgloss.Center)
	miniBackCompact = lipgloss.NewStyle().Background(p.FeltMiddle).Foreground(p.TextAccent)
	headerStyle = lipgloss.NewStyle().Background(p.FrameBorder).Foreground(p.TextPrimary).Bold(true).Padding(0, 2)
	scoreStyle = lipgloss.NewStyle().Background(p.ScoreBarBG).Foreground(p.TextPrimary).Padding(0, 1)
	chipStyle = lipgloss.NewStyle().Background(p.ScoreChipBG).Foreground(p.TextPrimary).Padding(0, 1).Bold(true)
	chipAccentStyle = chipStyle.Background(p.AccentChipBG).Foreground(p.TextAccent)
	chipTeamOneStyle = chipStyle.Foreground(p.TeamOne)
	chipTeamTwoStyle = chipStyle.Foreground(p.TeamTwo)
	teamOneColor = p.TeamOne
	teamTwoColor = p.TeamTwo
	helpStyle = lipgloss.NewStyle().Background(p.HelpBarBG).Foreground(p.TextDim).Padding(0, 1)
	tabBarStyle = lipgloss.NewStyle().Background(p.TabBarBG).Foreground(p.TextPrimary).Padding(0, 1)
	tabActiveStyle = lipgloss.NewStyle().Background(p.TabActiveBG).Foreground(p.TextWarning).Bold(true).Padding(0, 1)
	tabInactiveStyle = lipgloss.NewStyle().Foreground(p.TextDim).Padding(0, 1)
	nameLabelStyle = lipgloss.NewStyle().Foreground(p.TextPrimary).Bold(true)
	nameTeamOneStyle = nameLabelStyle.Foreground(p.TeamOne)
	nameTeamTwoStyle = nameLabelStyle.Foreground(p.TeamTwo)
	indexStyle = lipgloss.NewStyle().Foreground(p.TextWarning).Bold(true).Align(lipgloss.Center)
	viraLabelStyle = lipgloss.NewStyle().Foreground(p.TextAccent).Bold(true)
	roundLabelStyle = lipgloss.NewStyle().Foreground(p.TextWarning).Bold(true)
	alertStyle = lipgloss.NewStyle().Foreground(p.TextDanger).Bold(true)
	winnerStyle = lipgloss.NewStyle().Foreground(p.TextWarning).Bold(true)
	playerBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.TextMuted).Background(p.FeltMiddle).Padding(0, 1)
	activePlayerBoxStyle = playerBoxStyle.BorderForeground(p.TextWarning)
	activeNameStyle = lipgloss.NewStyle().Foreground(p.TextWarning).Bold(true)
	centerPanelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.TextMuted).Background(p.FeltCenter).Padding(0, 1)
	centerPanelCompactStyle = lipgloss.NewStyle().Background(p.FeltCenter)
	turnBadgeStyle = lipgloss.NewStyle().Background(p.TurnBadgeBG).Foreground(p.TextWarning).Bold(true).Padding(0, 1)
	leadingCardLabelStyle = lipgloss.NewStyle().Foreground(p.TextWarning).Bold(true)
	keyHintStyle = lipgloss.NewStyle().Foreground(p.TextDim)
	keyHintKeyStyle = lipgloss.NewStyle().Foreground(p.TextAccent).Bold(true)
	tableGridCellStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(p.GridBorder).
		Background(p.FeltCenter).
		Padding(0, 1)
}

func colorTo256BG(c lipgloss.Color) string {
	s := strings.TrimSpace(string(c))
	if s == "" {
		return ""
	}
	return "\x1b[48;5;" + s + "m"
}

func init() {
	ensureThemeStyles()
}

// ═══════════════════════════════════════════════════════════════════════════════
// View
// ═══════════════════════════════════════════════════════════════════════════════

func (m UIModel) View() string {
	ensureThemeStyles()
	if m.width == 0 {
		return tr("ui_loading")
	}
	return m.renderTable()
}

// ── Main renderer ───────────────────────────────────────────────────────────

func (m UIModel) renderTable() string {
	lp := computeLayout(m.width, m.height)
	w := lp.w

	localIdx := 0
	if m.snapshot.CurrentPlayerIdx >= 0 {
		localIdx = m.snapshot.CurrentPlayerIdx
	}
	playersMap := getRelativePlayers(m.snapshot.Players, localIdx)
	turnID := m.snapshot.TurnPlayer
	turnName := safePlayerName(m.snapshot.Players, turnID)
	scoreTurnName := clip(turnName, defaultTurnNameMax)
	if lp.compact {
		scoreTurnName = clip(turnName, compactTurnNameMax)
	}

	// ─── 1. Header ──────────────────────────────────────────────────────
	headerTitle := tr("header_title")
	header := headerStyle.Width(w).Align(lipgloss.Center).Render(fitSingleLine(headerTitle, maxInt(1, w-headerHorizontalPadding)))

	// ─── 2. Score bar ───────────────────────────────────────────────────
	t1Score := fmt.Sprintf("T1 %d", m.snapshot.MatchPoints[0])
	t2Score := fmt.Sprintf("T2 %d", m.snapshot.MatchPoints[1])
	stake := fmt.Sprintf("%s %d", tr("score_stake"), m.snapshot.CurrentHand.Stake)
	stakeLadder := fmt.Sprintf("%s %s", tr("score_stake_ladder"), stakeLadderLabel(m.snapshot.CurrentHand.Stake, m.snapshot.PendingRaiseTo))
	viraStr := suitSymbol(m.snapshot.CurrentHand.Vira)
	manilha := string(m.snapshot.CurrentHand.Manilha)
	roundInfo := fmt.Sprintf("R%d/3", m.snapshot.CurrentHand.Round)
	turnInfo := fmt.Sprintf("%s %s", tr("score_turn"), scoreTurnName)
	scoreInnerW := maxInt(1, w-scoreHorizontalPadding)
	var scoreSegments []string
	if lp.compact {
		scoreSegments = []string{
			chipTeamOneStyle.Render(t1Score),
			chipStyle.Render("x"),
			chipTeamTwoStyle.Render(t2Score),
			chipAccentStyle.Render(stake),
			chipStyle.Render(stakeLadder),
			chipStyle.Render("M " + manilha),
			chipAccentStyle.Render(turnInfo),
		}
	} else {
		scoreSegments = []string{
			chipTeamOneStyle.Render(t1Score),
			chipStyle.Render("x"),
			chipTeamTwoStyle.Render(t2Score),
			chipAccentStyle.Render(stake),
			chipStyle.Render(stakeLadder),
			chipStyle.Render(tr("score_flip_prefix") + viraStr),
			chipStyle.Render(tr("score_trump_prefix") + manilha),
			chipStyle.Render(roundInfo),
			chipAccentStyle.Render(turnInfo),
		}
	}
	scoreLine := joinSegmentsWithinWidth(scoreInnerW, scoreSegments...)
	scoreBar := scoreStyle.Width(w).Render(scoreLine)
	roleBar := helpStyle.Width(w).Render(
		fitSingleLine(m.renderRoleLane(maxInt(1, w-statusHorizontalPadding)), maxInt(1, w-statusHorizontalPadding)),
	)

	// ─── 3. Table area ──────────────────────────────────────────────────
	innerW := w - 2
	if innerW < 1 {
		innerW = 1
	}
	// Opponent at the top
	topSection := m.renderTopPlayer(playersMap["top"], turnID, lp.compact)

	// Center section (left player | table center | right player)
	centerSection := m.renderMiddle(playersMap["left"], playersMap["right"], turnID, turnName, innerW, lp)

	// Local player at the bottom
	botSection := m.renderBottomPlayer(localIdx, turnID, lp.compact)

	topBlock := fitBlock(topSection, innerW, lp.topH, lipgloss.Center, lipgloss.Center, feltEdgeColor)
	midBlock := fitBlock(centerSection, innerW, lp.midH, lipgloss.Center, lipgloss.Center, feltMiddleColor)
	botBlock := fitBlock(botSection, innerW, lp.botH, lipgloss.Center, lipgloss.Bottom, feltEdgeColor)

	tableBody := lipgloss.JoinVertical(lipgloss.Center, topBlock, midBlock, botBlock)
	// Keep the felt background active after nested style resets so the
	// "table" area never falls back to terminal default black.
	tableBody = enforceBackground(tableBody, ansiFeltBG)

	framedTable := frameBorderStyle.Render(tableBody)
	if m.trucoFlashFrames > 0 {
		framedTable = m.renderTrucoOverlay(innerW, lp.tableBodyH)
	} else if m.tableSweepFrames > 0 {
		framedTable = m.renderTrickSweepOverlay(innerW, lp.tableBodyH)
	} else if m.trickOverlayMsg != "" {
		framedTable = m.renderTrickOverlay(innerW, lp.tableBodyH, m.trickOverlayMsg)
	}

	tabBar := tabBarStyle.Width(w).Render(fitSingleLine(m.renderTabs(), maxInt(1, w-tabHorizontalPadding)))
	tabPanel := m.renderTabPanel(w, lp.panelLines)

	// ─── 4. Status / Help bar ───────────────────────────────────────────
	statusLine := fitSingleLine(m.buildStatusLine(), maxInt(1, w-statusHorizontalPadding))
	helpBar := helpStyle.Width(w).Render(statusLine)

	// ─── Assemble ───────────────────────────────────────────────────────
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		scoreBar,
		roleBar,
		framedTable,
		tabBar,
		tabPanel,
		helpBar,
	)
}

func (m UIModel) renderTrickOverlay(innerW, bodyH int, msg string) string {
	tone := overlayFadeColor(m.trickOverlayFrames)
	title := roundLabelStyle.Foreground(tone).Render(tr("overlay_trick_end_title"))
	info := winnerStyle.Foreground(tone).Render(msg)
	content := lipgloss.JoinVertical(lipgloss.Center, title, "", info)
	panel := fitBlock(content, innerW, bodyH, lipgloss.Center, lipgloss.Center, lgGreen)
	panel = enforceBackground(panel, ansiFeltBG)
	frame := frameBorderStyle.BorderForeground(tone)
	return frame.Render(panel)
}

func (m UIModel) renderTrucoOverlay(innerW, bodyH int) string {
	bg, fg, border := trucoFlashPalette(m.trucoFlashFrames, m.uiFrame)
	title := lipgloss.NewStyle().Foreground(fg).Bold(true).Render("TRUCO!")
	sub := lipgloss.NewStyle().Foreground(fg).Render(tr("overlay_stake_in_dispute"))
	content := lipgloss.JoinVertical(lipgloss.Center, title, "", sub)
	panel := fitBlock(content, innerW, bodyH, lipgloss.Center, lipgloss.Center, bg)
	frame := frameBorderStyle.BorderForeground(border)
	return frame.Render(panel)
}

func (m UIModel) renderTrickSweepOverlay(innerW, bodyH int) string {
	progress := float64(trickSweepAnimFrames-m.tableSweepFrames+1) / float64(trickSweepAnimFrames)
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	localIdx := 0
	if m.snapshot.CurrentPlayerIdx >= 0 {
		localIdx = m.snapshot.CurrentPlayerIdx
	}
	localPlayerID := localIdx
	if localIdx >= 0 && localIdx < len(m.snapshot.Players) {
		localPlayerID = m.snapshot.Players[localIdx].ID
	}

	rel := "center"
	if m.snapshot.LastTrickWinner >= 0 {
		rel = relativePosition(m.snapshot.LastTrickWinner, localPlayerID, m.snapshot.Players)
	}

	cardBacks := strings.TrimSpace(strings.Repeat("▒▒  ", maxInt(2, m.snapshot.NumPlayers)))
	pileWidth := lipgloss.Width(cardBacks)
	startX := maxInt(0, (innerW-pileWidth)/2)
	targetX := startX
	switch rel {
	case "left":
		targetX = 0
	case "right":
		targetX = maxInt(0, innerW-pileWidth)
	}
	x := int(lerp(float64(startX), float64(targetX), progress) + 0.5)
	x = clampInt(x, 0, maxInt(0, innerW-pileWidth))

	startY := maxInt(1, bodyH/2)
	targetY := startY
	switch rel {
	case "top":
		targetY = 1
	case "bottom":
		targetY = maxInt(1, bodyH-2)
	}
	y := int(lerp(float64(startY), float64(targetY), progress) + 0.5)
	y = clampInt(y, 1, maxInt(1, bodyH-1))

	lines := make([]string, bodyH)
	for i := range lines {
		lines[i] = fitSingleLine("", innerW)
	}
	titleText := tr("overlay_collecting_trick")
	if m.snapshot.LastTrickWinner >= 0 {
		titleText = fmt.Sprintf(tr("overlay_collecting_trick_by_format"), safePlayerName(m.snapshot.Players, m.snapshot.LastTrickWinner))
	}
	lines[0] = fitSingleLine(lipgloss.NewStyle().Foreground(lgGray).Bold(true).Render(titleText), innerW)

	pile := lipgloss.NewStyle().Foreground(lgDim).Render(cardBacks)
	lines[y] = fitSingleLine(strings.Repeat(" ", x)+pile, innerW)

	panel := fitBlock(strings.Join(lines, "\n"), innerW, bodyH, lipgloss.Left, lipgloss.Top, lgGreen)
	panel = enforceBackground(panel, ansiFeltBG)
	frame := frameBorderStyle.BorderForeground(lgDim)
	return frame.Render(panel)
}

func (m UIModel) renderTabs() string {
	tabs := []string{"mesa", "chat", "log"}
	tabLabels := map[string]string{
		"mesa": tr("tab_mesa"),
		"chat": tr("tab_chat"),
		"log":  tr("tab_log"),
	}
	rendered := make([]string, 0, len(tabs))
	for _, t := range tabs {
		label := strings.ToUpper(tabLabels[t])
		if t == m.activeTab {
			rendered = append(rendered, tabActiveStyle.Render(label))
		} else {
			rendered = append(rendered, tabInactiveStyle.Render(label))
		}
	}
	return tr("tabs_label") + strings.Join(rendered, " ")
}

func (m UIModel) renderTabPanel(w int, linesLimit int) string {
	panel := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lgWhite).
		Padding(0, 1).
		Width(w)
	contentW := maxInt(1, w-panelHorizontalPadding)

	lines := make([]string, 0, 3)
	switch m.activeTab {
	case "chat":
		lines = append(lines, tr("panel_chat_desc")+" "+tr("panel_chat_message_prefix")+m.renderChatInputWithCursor())
		if strings.TrimSpace(m.chatInput) == "" {
			lines = append(lines, tr("panel_chat_hint"))
		}
		lines = append(lines, m.chatCommandsHint())
		for _, line := range tailLines(m.chatLog, 2) {
			lines = append(lines, styleEventLine(line))
		}
	case "log":
		lines = append(lines, tr("panel_log_title"))
		scoreEvents := scoreHistoryFromLogs(m.snapshot.Logs, 2)
		if len(scoreEvents) > 0 {
			lines = append(lines, tr("panel_recent_score_prefix")+strings.Join(scoreEvents, " | "))
		}
		if len(m.errorLog) > 0 {
			lines = append(lines, tr("panel_error_log_title"))
			for _, line := range tailLines(m.errorLog, 2) {
				lines = append(lines, styleEventLine(line))
			}
		}
		for _, line := range tailLines(m.snapshot.Logs, 2) {
			lines = append(lines, styleEventLine(line))
		}
	default:
		wins := m.snapshot.CurrentHand.TrickWins
		lines = append(lines, fmt.Sprintf("%s: %s T1 %d x %d T2 | %s %d", tr("panel_table_label"), tr("panel_tricks_label"), wins[0], wins[1], tr("panel_round_label"), m.snapshot.CurrentHand.Round))
		lines = append(lines, tr("panel_history_prefix")+formatTrickHistory(m.snapshot.CurrentHand.TrickResults))
		lines = append(lines, fmt.Sprintf("%s %s | %s %s | %s %d", tr("panel_trump_label"), m.snapshot.CurrentHand.Manilha, tr("panel_flip_label"), suitSymbol(m.snapshot.CurrentHand.Vira), tr("panel_stake_label"), m.snapshot.CurrentHand.Stake))
		if m.snapshot.PendingRaiseFor != -1 {
			raiseBy := safePlayerName(m.snapshot.Players, m.snapshot.CurrentHand.RaiseRequester)
			lines = append(lines, fmt.Sprintf(tr("panel_raise_pending_format"), strings.ToUpper(raiseLabelUI(m.snapshot.PendingRaiseTo)), m.snapshot.PendingRaiseTo, raiseBy))
		}
	}
	if linesLimit < 1 {
		linesLimit = 1
	}
	for len(lines) < linesLimit {
		lines = append(lines, "")
	}
	for i := range lines {
		lines[i] = fitSingleLine(lines[i], contentW)
	}
	return panel.Render(strings.Join(lines[:linesLimit], "\n"))
}

func tailLines(items []string, n int) []string {
	if n <= 0 || len(items) == 0 {
		return nil
	}
	if len(items) <= n {
		return items
	}
	return items[len(items)-n:]
}

func styleEventLine(line string) string {
	trimmed := strings.TrimSpace(strings.ToLower(line))
	errPrefix := strings.TrimSpace(strings.ToLower(tr("error_prefix")))
	switch {
	case strings.HasPrefix(trimmed, "[system]"):
		return keyHintStyle.Render(line)
	case errPrefix != "" && strings.HasPrefix(trimmed, errPrefix):
		return alertStyle.Render(line)
	default:
		return line
	}
}

// ── Sections ────────────────────────────────────────────────────────────────

func (m UIModel) renderTopPlayer(p *truco.Player, turnID int, compact bool) string {
	if p == nil {
		return ""
	}
	nameMax := defaultPlayerNameMax
	if compact {
		nameMax = compactPlayerNameMax
	}
	visibleCards := len(p.Hand)
	if m.dealFrames > 0 {
		visibleCards = cardsRevealed(visibleCards, m.dealFrames)
	}
	hand := renderMiniHand(visibleCards, compact)
	return m.renderOpponentPlayer(p, turnID, nameMax, compact, "top", hand, true)
}

func (m UIModel) renderMiddle(left, right *truco.Player, turnID int, turnName string, w int, lp layoutProfile) string {
	sideW := lp.sideW
	centerW := w - sideW*2
	if centerW < lp.centerMinW {
		centerW = lp.centerMinW
		sideW = maxInt(1, (w-centerW)/2)
	}
	if sideW < 1 {
		sideW = 1
	}
	if centerW < 1 {
		centerW = 1
	}

	nameMax := maxInt(4, sideW-4)
	leftBlock := m.renderSidePlayer(left, turnID, nameMax, lp.compact, "left")
	rightBlock := m.renderSidePlayer(right, turnID, nameMax, lp.compact, "right")

	// Center: vira card + played cards
	centerBlock := m.renderCenter(centerW, turnName, lp.compact)

	leftCell := fitWidth(leftBlock, sideW, lipgloss.Center)
	centerCell := fitWidth(centerBlock, centerW, lipgloss.Center)
	rightCell := fitWidth(rightBlock, sideW, lipgloss.Center)

	return lipgloss.JoinHorizontal(lipgloss.Center,
		paintBlockBackground(leftCell, feltMiddleColor),
		paintBlockBackground(centerCell, feltCenterColor),
		paintBlockBackground(rightCell, feltMiddleColor),
	)
}

func (m UIModel) renderSidePlayer(p *truco.Player, turnID, nameMax int, compact bool, side string) string {
	if p == nil {
		return ""
	}
	visibleCards := len(p.Hand)
	if m.dealFrames > 0 {
		visibleCards = cardsRevealed(visibleCards, m.dealFrames)
	}
	hand := renderMiniHandVertical(visibleCards, compact)
	return m.renderOpponentPlayer(p, turnID, nameMax, compact, side, hand, false)
}

func turnMarkerForPosition(pos string) string {
	switch pos {
	case "top":
		return "▼"
	case "bottom":
		return "▲"
	case "left":
		return "▶"
	case "right":
		return "◀"
	default:
		return "▶"
	}
}

func (m UIModel) renderOpponentPlayer(p *truco.Player, turnID, nameMax int, compact bool, pos, hand string, compactInline bool) string {
	if p == nil {
		return ""
	}
	nameText := clip(playerNameTag(*p), nameMax)
	labelStyle := teamNameStyleForTeam(p.Team)
	boxStyle := playerBoxStyle
	if p.ID == turnID {
		nameText = fmt.Sprintf("%s %s (%s)", turnMarkerForPosition(pos), nameText, tr("ui_turn_short"))
		labelStyle = activeNameStyle.Foreground(teamColorForTeam(p.Team))
		boxStyle = activePlayerBoxStyle
	}
	label := labelStyle.Render(nameText)
	localIdx := m.localPlayerIdx
	if m.snapshot.CurrentPlayerIdx >= 0 {
		localIdx = m.snapshot.CurrentPlayerIdx
	}
	roleInfo := deriveSeatRoles(m.snapshot, localIdx, m.isOnline)[p.ID]
	roleLine := keyHintStyle.Render(strings.Join(m.roleBadgeLabels(roleInfo), " · "))
	if compact {
		if compactInline {
			return lipgloss.JoinHorizontal(lipgloss.Left, label, " ", roleLine, " ", hand)
		}
		return lipgloss.JoinVertical(lipgloss.Left, label, roleLine, hand)
	}
	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, label, roleLine, hand))
}

func (m UIModel) renderCenter(w int, turnName string, compact bool) string {
	s := m.snapshot

	badgeName := clip(turnName, defaultTurnNameMax)
	if compact {
		badgeName = clip(turnName, compactTurnNameMax)
	}
	turnBadgeText := tr("turn_badge_prefix") + badgeName
	if s.TurnPlayer >= 0 && s.TurnPlayer < len(s.Players) && s.Players[s.TurnPlayer].CPU && !s.MatchFinished && s.PendingRaiseFor == -1 {
		turnBadgeText = fmt.Sprintf("%s %s", turnBadgeText, cpuSpinnerFrame(m.uiFrame))
	}
	turnBadge := turnBadgeStyle.Render(turnBadgeText)
	if compact {
		turnBadge = turnBadgeStyle.Padding(0, 0).Render(turnBadgeText)
	}

	// Vira card
	viraLabel := viraLabelStyle.Render(tr("score_flip_label"))
	viraCard := renderBigCard(s.CurrentHand.Vira, false, compact)
	viraBlock := lipgloss.JoinVertical(lipgloss.Center, viraLabel, viraCard)

	// Played cards
	leading := leadingCardIndexes(s.CurrentHand.RoundCards, s.CurrentHand.Manilha)
	var playedCards []string
	nameMax := defaultPlayedNameMax
	if compact {
		nameMax = compactPlayedNameMax
	}
	for i, pc := range s.CurrentHand.RoundCards {
		pName := clip(safePlayerName(s.Players, pc.PlayerID), nameMax)
		label := lipgloss.NewStyle().Foreground(lgDim).Render(pName)
		card := renderBigCard(pc.Card, leading[i], compact)
		if leading[i] {
			label = leadingCardLabelStyle.Render("★ " + pName)
		}
		cardBlock := lipgloss.JoinVertical(lipgloss.Center, label, card)
		if m.playAnimFrames > 0 && i == m.playAnimCardIndex {
			// Pulso breve de destaque no destino enquanto a carta "viaja".
			cardBlock = lipgloss.NewStyle().Bold(true).Render(cardBlock)
		}
		playedCards = append(playedCards, cardBlock)
	}

	panelStyle := centerPanelStyle
	gap := "   "
	cardGap := "  "
	if compact {
		panelStyle = centerPanelCompactStyle
		gap = compactGap
		cardGap = compactCardGap
	} else {
		gap = defaultGap
		cardGap = defaultCardGap
	}

	flight := m.renderFlightRow(w, compact)

	if len(playedCards) == 0 {
		emptyMsg := lipgloss.NewStyle().Foreground(lgGray).Render(tr("table_empty"))
		content := lipgloss.JoinHorizontal(lipgloss.Center, viraBlock, gap, emptyMsg)
		if !compact {
			viraCell := tableGridCellStyle.Render(viraBlock)
			playedCell := tableGridCellStyle.Render(emptyMsg)
			content = lipgloss.JoinHorizontal(lipgloss.Top, viraCell, " ", playedCell)
		}
		rows := []string{turnBadge}
		if m.dealFrames > 0 {
			rows = append(rows, lipgloss.NewStyle().Foreground(lgCyan).Render(tr("table_shuffling_dealing")))
		}
		if flight != "" {
			rows = append(rows, flight)
		}
		rows = append(rows, content)
		panel := lipgloss.JoinVertical(lipgloss.Center, rows...)
		return panelStyle.Render(panel)
	}

	played := joinHorizontalWithSpacer(lipgloss.Top, cardGap, playedCards...)
	content := lipgloss.JoinHorizontal(lipgloss.Top, viraBlock, gap, played)
	if !compact {
		viraCell := tableGridCellStyle.Render(viraBlock)
		playedCell := tableGridCellStyle.Render(played)
		content = lipgloss.JoinHorizontal(lipgloss.Top, viraCell, " ", playedCell)
	}
	rows := []string{turnBadge}
	if m.dealFrames > 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(lgCyan).Render(tr("table_shuffling_dealing")))
	}
	if flight != "" {
		rows = append(rows, flight)
	}
	rows = append(rows, content)
	panel := lipgloss.JoinVertical(lipgloss.Center, rows...)
	return panelStyle.Render(panel)
}

func (m UIModel) renderBottomPlayer(localIdx int, turnID int, compact bool) string {
	me := m.snapshot.Players[localIdx]

	// Render face‑up cards with numeric indices underneath
	var cards []string
	var indices []string
	idxWidth := 8
	if compact {
		idxWidth = compactCardIndexWidth
	} else {
		idxWidth = defaultCardIndexWidth
	}

	hand := me.Hand
	if m.dealFrames > 0 {
		reveal := cardsRevealed(len(hand), m.dealFrames)
		if reveal < len(hand) {
			hand = hand[:reveal]
		}
	}

	ghostActive := m.ghostFrames > 0 && m.ghostCardSlot >= 0
	totalSlots := len(hand)
	if ghostActive {
		totalSlots++
	}
	playableIdx := 1
	for slot := 0; slot < totalSlots; slot++ {
		if ghostActive && slot == m.ghostCardSlot {
			flash := m.inputFlashFrames > 0 && m.inputFlashSlot == m.ghostCardSlot
			cards = append(cards, renderGhostCard(compact, flash))
			indices = append(indices, indexStyle.Width(idxWidth).Render("   "))
			continue
		}
		handIdx := slot
		if ghostActive && slot > m.ghostCardSlot {
			handIdx--
		}
		if handIdx < 0 || handIdx >= len(hand) {
			continue
		}
		cards = append(cards, renderBigCard(hand[handIdx], false, compact))
		idx := indexStyle.Width(idxWidth).Render(fmt.Sprintf("[%d]", playableIdx))
		indices = append(indices, idx)
		playableIdx++
	}

	handRow := lipgloss.NewStyle().Foreground(lgGray).Render(tr("table_no_cards"))
	if len(cards) > 0 {
		handRow = lipgloss.JoinHorizontal(lipgloss.Top, cards...)
	}
	idxRow := ""
	if len(indices) > 0 {
		idxRow = lipgloss.JoinHorizontal(lipgloss.Top, indices...)
	}

	// Name with turn indicator
	nameMax := defaultBottomNameMax
	if compact {
		nameMax = compactBottomNameMax
	}
	nameText := clip(playerNameTag(me), nameMax) + " (" + tr("ui_you_label") + ")"
	boxStyle := playerBoxStyle
	nameStyle := teamNameStyleForTeam(me.Team)
	if turnID == me.ID {
		nameText = fmt.Sprintf("%s %s (%s)", turnMarkerForPosition("bottom"), nameText, tr("ui_turn_short"))
		nameStyle = activeNameStyle.Foreground(teamColorForTeam(me.Team))
		boxStyle = activePlayerBoxStyle
	}
	nameLabel := nameStyle.Render(nameText)
	roleLine := keyHintStyle.Render(strings.Join(m.roleBadgeLabels(deriveSeatRoles(m.snapshot, localIdx, m.isOnline)[me.ID]), " · "))

	if compact {
		rows := []string{handRow}
		if idxRow != "" {
			rows = append(rows, idxRow)
		}
		rows = append(rows, nameLabel)
		rows = append(rows, roleLine)
		return lipgloss.JoinVertical(lipgloss.Center, rows...)
	}
	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, handRow, idxRow, nameLabel, roleLine))
}

// ── Cards ───────────────────────────────────────────────────────────────────

func renderBigCard(c truco.Card, leading bool, compact bool) string {
	sym := cardSuitGlyph(c.Suit)
	rank := string(c.Rank)
	fg := lgBlack
	if c.Suit == "Ouros" || c.Suit == "Copas" {
		fg = lgRed
	}
	if compact {
		face := fmt.Sprintf("┌%s%s┐\n└───┘", rank, sym)
		style := bigCardCompact.Foreground(fg)
		if leading {
			style = style.Bold(true)
		}
		return style.Render(face)
	}

	topLeft := rank + sym
	botRight := sym + rank
	lines := []string{
		"┌─────┐",
		fmt.Sprintf("│%-2s   │", topLeft),
		fmt.Sprintf("│  %s  │", sym),
		fmt.Sprintf("│   %-2s│", botRight),
		"└─────┘",
	}
	cardFace := strings.Join(lines, "\n")
	style := bigCardBlack.Foreground(fg)
	if leading {
		style = style.Bold(true)
	}
	rendered := style.Render(cardFace)
	if leading {
		return lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lgYellow).
			Render(rendered)
	}
	return rendered
}

// renderMiniHand renders N compact card‑backs horizontally (for top opponent).
func renderMiniHand(n int, compact bool) string {
	if n == 0 {
		return ""
	}
	cardStyle := miniBack
	if compact {
		cardStyle = miniBackCompact
	}
	var cards []string
	for i := 0; i < n; i++ {
		cards = append(cards, cardStyle.Render("▒▒"))
	}
	return lipgloss.JoinHorizontal(lipgloss.Center, cards...)
}

// renderMiniHandVertical renders N compact card‑backs stacked vertically (for side opponents).
func renderMiniHandVertical(n int, compact bool) string {
	if n == 0 {
		return ""
	}
	cardStyle := miniBack
	if compact {
		cardStyle = miniBackCompact
	}
	var cards []string
	for i := 0; i < n; i++ {
		cards = append(cards, cardStyle.Render("▒▒"))
	}
	return lipgloss.JoinVertical(lipgloss.Center, cards...)
}

func suitSymbol(c truco.Card) string {
	sym := cardSuitGlyph(c.Suit)
	return fmt.Sprintf("%s%s", string(c.Rank), sym)
}

func cardSuitGlyph(suit truco.Suit) string {
	switch suit {
	case "Ouros":
		return "♦"
	case "Copas":
		return "♥"
	case "Espadas":
		return "♠"
	case "Paus":
		return "♣"
	default:
		return "?"
	}
}

func playerNameTag(p truco.Player) string {
	tag := fmt.Sprintf("%s [T%d]", p.Name, p.Team+1)
	if p.CPU && p.ProvisionalCPU {
		return tag + " [CPU*]"
	}
	if p.CPU {
		return tag + " [CPU]"
	}
	return tag
}

func teamColorForTeam(team int) lipgloss.Color {
	if team%2 == 0 {
		return teamOneColor
	}
	return teamTwoColor
}

func teamNameStyleForTeam(team int) lipgloss.Style {
	if team%2 == 0 {
		return nameTeamOneStyle
	}
	return nameTeamTwoStyle
}

func safePlayerName(players []truco.Player, id int) string {
	if id >= 0 && id < len(players) {
		return players[id].Name
	}
	return "?"
}

func leadingCardIndexes(played []truco.PlayedCard, manilha truco.Rank) map[int]bool {
	out := make(map[int]bool, len(played))
	if len(played) == 0 {
		return out
	}
	best := truco.CardPower(played[0].Card, manilha)
	for i := 1; i < len(played); i++ {
		p := truco.CardPower(played[i].Card, manilha)
		if p > best {
			best = p
		}
	}
	for i := range played {
		if truco.CardPower(played[i].Card, manilha) == best {
			out[i] = true
		}
	}
	return out
}
