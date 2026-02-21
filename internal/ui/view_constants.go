package ui

import "github.com/charmbracelet/lipgloss"

const (
	minRenderWidth  = 80
	minRenderHeight = 24

	compactMaxWidth  = 104
	compactMaxHeight = 32

	headerLines      = 1
	scoreLines       = 1
	tabLines         = 1
	helpLines        = 1
	frameBorderLines = 2

	panelLinesCompact = 1
	panelLinesFull    = 3

	minTableBodyHeight = 10

	compactTopMin        = 1
	compactBottomMin     = 3
	compactMidMin        = 4
	compactTopDivisor    = 6
	compactBottomDivisor = 4

	fullTopMin        = 4
	fullBottomMin     = 8
	fullMidMin        = 7
	fullTopDivisor    = 5
	fullBottomDivisor = 4

	sideWidthDivisor        = 6
	sideWidthMin            = 12
	sideWidthMax            = 30
	compactSideWidthDivisor = 8
	compactSideWidthMin     = 8
	compactSideWidthMax     = 12
	centerMinWidthFull      = 30
	centerMinWidthCompact   = 18

	headerHorizontalPadding = 4
	scoreHorizontalPadding  = 2
	statusHorizontalPadding = 2
	tabHorizontalPadding    = 2
	panelHorizontalPadding  = 2

	compactTurnNameMax   = 8
	defaultTurnNameMax   = 12
	compactPlayerNameMax = 12
	defaultPlayerNameMax = 18
	compactBottomNameMax = 16
	defaultBottomNameMax = 24
	compactPlayedNameMax = 7
	defaultPlayedNameMax = 12

	compactCardIndexWidth = 4
	defaultCardIndexWidth = 8

	compactGap     = " "
	defaultGap     = "   "
	compactCardGap = " "
	defaultCardGap = "  "

	animShiftStep          = 2
	playAnimMaxFrames      = 12
	trickOverlayAnimFrames = 10
	trickSweepAnimFrames   = 3
	trucoFlashAnimFrames   = 12
	dealAnimFrames         = 8
	ghostAnimFrames        = 6
	inputConfirmFrames     = 3

	alignStart  = 0
	alignMiddle = 1
	alignEnd    = 2
)

var (
	feltEdgeColor   lipgloss.Color
	feltMiddleColor lipgloss.Color
	feltCenterColor lipgloss.Color
)
