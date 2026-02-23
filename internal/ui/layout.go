package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type layoutProfile struct {
	w          int
	h          int
	compact    bool
	panelLines int
	tableBodyH int
	topH       int
	midH       int
	botH       int
	sideW      int
	centerMinW int
}

func computeLayout(w, h int) layoutProfile {
	if w < minRenderWidth {
		w = minRenderWidth
	}
	if h < minRenderHeight {
		h = minRenderHeight
	}
	compact := w <= compactMaxWidth || h <= compactMaxHeight
	panelLines := panelLinesFull
	if compact {
		panelLines = panelLinesCompact
	}

	// header + score + role + tabs + panel + help + outer table border
	fixedOutsideTable := headerLines + scoreLines + roleLines + tabLines + panelLines + helpLines + frameBorderLines
	tableBodyH := h - fixedOutsideTable
	if tableBodyH < minTableBodyHeight {
		tableBodyH = minTableBodyHeight
	}

	topFloor := compactTopMin
	botFloor := compactBottomMin
	minMid := compactMidMin
	topH := maxInt(topFloor, tableBodyH/compactTopDivisor)
	botH := maxInt(botFloor, tableBodyH/compactBottomDivisor)

	if !compact {
		topFloor = fullTopMin
		botFloor = fullBottomMin
		minMid = fullMidMin
		topH = maxInt(topFloor, tableBodyH/fullTopDivisor)
		botH = maxInt(botFloor, tableBodyH/fullBottomDivisor)
	}

	midH := tableBodyH - topH - botH
	if midH < minMid {
		short := minMid - midH
		cutTop := minInt(short, topH-topFloor)
		topH -= cutTop
		short -= cutTop
		if short > 0 {
			cutBot := minInt(short, botH-botFloor)
			botH -= cutBot
		}
		midH = tableBodyH - topH - botH
	}

	sideW := clampInt(w/sideWidthDivisor, sideWidthMin, sideWidthMax)
	centerMinW := centerMinWidthFull
	if compact {
		sideW = clampInt(w/compactSideWidthDivisor, compactSideWidthMin, compactSideWidthMax)
		centerMinW = centerMinWidthCompact
	}

	return layoutProfile{
		w:          w,
		h:          h,
		compact:    compact,
		panelLines: panelLines,
		tableBodyH: tableBodyH,
		topH:       topH,
		midH:       midH,
		botH:       botH,
		sideW:      sideW,
		centerMinW: centerMinW,
	}
}

func fitWidth(block string, width int, align lipgloss.Position) string {
	if width <= 0 {
		return ""
	}
	lines := strings.Split(strings.ReplaceAll(block, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	for i := range lines {
		lines[i] = alignLine(lines[i], width, align)
	}
	return strings.Join(lines, "\n")
}

func fitBlock(block string, width, height int, hAlign, vAlign lipgloss.Position, bg lipgloss.Color) string {
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}

	lines := strings.Split(strings.ReplaceAll(block, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	for i := range lines {
		lines[i] = alignLine(lines[i], width, hAlign)
	}

	if len(lines) > height {
		start := 0
		switch alignBucket(vAlign) {
		case alignMiddle:
			start = (len(lines) - height) / 2
		case alignEnd:
			start = len(lines) - height
		}
		lines = lines[start : start+height]
	}

	if len(lines) < height {
		missing := height - len(lines)
		blank := strings.Repeat(" ", width)
		topPad, bottomPad := 0, missing
		switch alignBucket(vAlign) {
		case alignMiddle:
			topPad = missing / 2
			bottomPad = missing - topPad
		case alignEnd:
			topPad = missing
			bottomPad = 0
		}
		withPad := make([]string, 0, height)
		for i := 0; i < topPad; i++ {
			withPad = append(withPad, blank)
		}
		withPad = append(withPad, lines...)
		for i := 0; i < bottomPad; i++ {
			withPad = append(withPad, blank)
		}
		lines = withPad
	}

	lineStyle := lipgloss.NewStyle().Background(bg)
	for i := range lines {
		lines[i] = lineStyle.Render(lines[i])
	}
	return strings.Join(lines, "\n")
}

func alignLine(line string, width int, align lipgloss.Position) string {
	if width <= 0 {
		return ""
	}
	line = strings.ReplaceAll(line, "\n", " ")
	line = ansi.Truncate(line, width, "")
	lineW := lipgloss.Width(line)
	if lineW >= width {
		return line
	}

	pad := width - lineW
	leftPad, rightPad := 0, pad
	switch alignBucket(align) {
	case alignMiddle:
		leftPad = pad / 2
		rightPad = pad - leftPad
	case alignEnd:
		leftPad = pad
		rightPad = 0
	}
	return strings.Repeat(" ", leftPad) + line + strings.Repeat(" ", rightPad)
}

func alignBucket(pos lipgloss.Position) int {
	switch {
	case pos >= 0.75:
		return alignEnd
	case pos >= 0.25:
		return alignMiddle
	default:
		return alignStart
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
