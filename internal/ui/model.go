package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"truco-tui/internal/truco"
)

const (
	cpuTurnDelay         = 1500 * time.Millisecond
	errorDisplayDuration = 2 * time.Second
	uiTickInterval       = 140 * time.Millisecond
)

// UIModel is the main Bubble Tea model state
type UIModel struct {
	game     *truco.Game
	snapshot truco.Snapshot

	width  int
	height int

	// State management
	activeTab  string
	chatInput  string
	chatCursor int
	chatLog    []string
	errorLog   []string

	err error

	// Game State
	localPlayerIdx int

	visualState
}

type syncMsg struct {
	snapshot truco.Snapshot
	err      error
}

type clearTrickOverlayMsg struct {
	id int
}

type playAnimTickMsg struct {
	id int
}

type clearErrorMsg struct {
	id int
}

type uiTickMsg struct{}

func InitialModel(g *truco.Game) UIModel {
	snap := g.Snapshot(0)
	return UIModel{
		game:           g,
		snapshot:       snap,
		activeTab:      "mesa",
		chatLog:        []string{tr("chat_welcome")},
		localPlayerIdx: 0,
		visualState:    newVisualState(snap),
	}
}

func (m UIModel) Init() tea.Cmd {
	return tea.Batch(
		maybeScheduleCPUCmd(m.game, m.snapshot, m.localPlayerIdx),
		uiTickCmd(),
	)
}

func snapshotCmd(g *truco.Game, localPlayerIdx int) tea.Cmd {
	return func() tea.Msg {
		return syncMsg{snapshot: g.Snapshot(localPlayerIdx)}
	}
}

func cpuStepCmd(g *truco.Game, localPlayerIdx int) tea.Cmd {
	return tea.Tick(cpuTurnDelay, func(time.Time) tea.Msg {
		snap := g.Snapshot(localPlayerIdx)
		if snap.MatchFinished {
			return syncMsg{snapshot: snap}
		}
		isCPU, pid := g.IsCPUTurn()
		if !isCPU {
			return syncMsg{snapshot: snap}
		}
		act := truco.DecideCPUAction(g, pid)
		if err := applyCPUActionToGame(g, pid, act); err != nil {
			return syncMsg{snapshot: g.Snapshot(localPlayerIdx), err: err}
		}
		return syncMsg{snapshot: g.Snapshot(localPlayerIdx)}
	})
}

func maybeScheduleCPUCmd(g *truco.Game, snap truco.Snapshot, localPlayerIdx int) tea.Cmd {
	if snap.MatchFinished {
		return nil
	}
	isCPU, _ := g.IsCPUTurn()
	if !isCPU {
		return nil
	}
	return cpuStepCmd(g, localPlayerIdx)
}

func clearTrickOverlayCmd(id int) tea.Cmd {
	return tea.Tick(1400*time.Millisecond, func(time.Time) tea.Msg {
		return clearTrickOverlayMsg{id: id}
	})
}

func playAnimTickCmd(id int) tea.Cmd {
	return tea.Tick(55*time.Millisecond, func(time.Time) tea.Msg {
		return playAnimTickMsg{id: id}
	})
}

func clearErrorCmd(id int) tea.Cmd {
	return tea.Tick(errorDisplayDuration, func(time.Time) tea.Msg {
		return clearErrorMsg{id: id}
	})
}

func uiTickCmd() tea.Cmd {
	return tea.Tick(uiTickInterval, func(time.Time) tea.Msg {
		return uiTickMsg{}
	})
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case syncMsg:
		prev := m.snapshot
		m.snapshot = msg.snapshot
		m.visualState.applySnapshotVisualTransitions(prev, m.snapshot)
		cmds := []tea.Cmd{
			m.visualState.updateTrickOverlay(m.snapshot, m.localPlayerIdx),
			m.visualState.updatePlayAnimation(m.snapshot),
			maybeScheduleCPUCmd(m.game, m.snapshot, m.localPlayerIdx),
		}
		if msg.err != nil {
			if errCmd := m.setTransientError(msg.err); errCmd != nil {
				cmds = append(cmds, errCmd)
			}
		}
		return m, tea.Batch(cmds...)

	case clearTrickOverlayMsg:
		m.visualState.onClearTrickOverlay(msg.id)

	case playAnimTickMsg:
		if cmd := m.visualState.onPlayAnimTick(msg.id); cmd != nil {
			return m, cmd
		}

	case clearErrorMsg:
		m.visualState.onClearError(msg.id, &m.err)

	case uiTickMsg:
		m.visualState.onUITick()
		return m, uiTickCmd()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.activeTab == "mesa" {
				m.activeTab = "chat"
			} else if m.activeTab == "chat" {
				m.activeTab = "log"
			} else {
				m.activeTab = "mesa"
			}
			if m.activeTab == "chat" {
				m.chatCursor = len([]rune(m.chatInput))
			}
			m.snapshot = m.game.Snapshot(m.localPlayerIdx)
			return m, maybeScheduleCPUCmd(m.game, m.snapshot, m.localPlayerIdx)
		}
		if submitted, handled := m.handleChatInputKey(msg); handled {
			if submitted != "" {
				actor := safePlayerName(m.snapshot.Players, m.localPlayerIdx)
				m.chatLog = appendChatLine(m.chatLog, actor, submitted)
			}
			return m, nil
		}
		switch msg.String() {
		case "1", "2", "3":
			if m.trickOverlayMsg != "" {
				return m, nil
			}
			idx := int(msg.Runes[0] - '1')
			if err := m.game.PlayCard(m.localPlayerIdx, idx); err != nil {
				return m, tea.Batch(snapshotCmd(m.game, m.localPlayerIdx), m.setTransientError(err))
			}
			m.visualState.onCardAccepted(idx)
			return m, snapshotCmd(m.game, m.localPlayerIdx)
		case "t":
			if m.trickOverlayMsg != "" {
				return m, nil
			}
			if err := requestOrRaiseTruco(m.game, m.localPlayerIdx); err != nil {
				return m, tea.Batch(snapshotCmd(m.game, m.localPlayerIdx), m.setTransientError(err))
			}
			return m, snapshotCmd(m.game, m.localPlayerIdx)
		case "a":
			if m.trickOverlayMsg != "" {
				return m, nil
			}
			if err := m.game.RespondTruco(m.localPlayerIdx, true); err != nil {
				return m, tea.Batch(snapshotCmd(m.game, m.localPlayerIdx), m.setTransientError(err))
			}
			return m, snapshotCmd(m.game, m.localPlayerIdx)
		case "r":
			if m.trickOverlayMsg != "" {
				return m, nil
			}
			if err := m.game.RespondTruco(m.localPlayerIdx, false); err != nil {
				return m, tea.Batch(snapshotCmd(m.game, m.localPlayerIdx), m.setTransientError(err))
			}
			return m, snapshotCmd(m.game, m.localPlayerIdx)
		}
	}
	return m, cmd
}

func applyCPUActionToGame(g *truco.Game, pid int, a truco.CPUAction) error {
	switch a.Kind {
	case "ask_truco":
		return g.AskTruco(pid)
	case "raise":
		return g.RaiseTruco(pid)
	case "accept":
		return g.RespondTruco(pid, true)
	case "refuse":
		return g.RespondTruco(pid, false)
	case "play":
		return g.PlayCard(pid, a.CardIndex)
	}
	return nil
}
