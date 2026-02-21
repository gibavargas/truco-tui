package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"truco-tui/internal/netp2p"
)

const onlineLobbyTickInterval = 220 * time.Millisecond

type onlineLobbyTickMsg struct{}

type onlineLobbyHostEventMsg struct {
	text string
	ok   bool
}

type onlineLobbyClientEventMsg struct {
	text string
	ok   bool
}

type lobbyMode int

const (
	lobbyModeHost lobbyMode = iota
	lobbyModeClient
)

type onlineLobbyModel struct {
	mode lobbyMode

	host *netp2p.HostSession
	cli  *netp2p.ClientSession
	key  string

	startMatch bool
	back       bool

	input  string
	cursor int
	events []string
	err    error
	frame  int
}

func newOnlineLobbyHostModel(host *netp2p.HostSession, key string) onlineLobbyModel {
	return onlineLobbyModel{
		mode:   lobbyModeHost,
		host:   host,
		key:    key,
		events: []string{},
	}
}

func newOnlineLobbyClientModel(cli *netp2p.ClientSession) onlineLobbyModel {
	return onlineLobbyModel{
		mode:   lobbyModeClient,
		cli:    cli,
		events: []string{},
	}
}

func (m onlineLobbyModel) Init() tea.Cmd {
	cmds := []tea.Cmd{onlineLobbyTickCmd()}
	if m.mode == lobbyModeHost {
		cmds = append(cmds, waitHostLobbyEventCmd(m.host))
	} else {
		cmds = append(cmds, waitClientLobbyEventCmd(m.cli))
	}
	return tea.Batch(cmds...)
}

func onlineLobbyTickCmd() tea.Cmd {
	return tea.Tick(onlineLobbyTickInterval, func(time.Time) tea.Msg {
		return onlineLobbyTickMsg{}
	})
}

func waitHostLobbyEventCmd(host *netp2p.HostSession) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-host.Events()
		return onlineLobbyHostEventMsg{text: ev, ok: ok}
	}
}

func waitClientLobbyEventCmd(cli *netp2p.ClientSession) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-cli.Events()
		return onlineLobbyClientEventMsg{text: ev, ok: ok}
	}
}

func (m onlineLobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case onlineLobbyTickMsg:
		m.frame++
		if m.mode == lobbyModeClient && m.cli != nil && m.cli.GameStarted() {
			m.startMatch = true
			return m, tea.Quit
		}
		return m, onlineLobbyTickCmd()

	case onlineLobbyHostEventMsg:
		if !msg.ok {
			return m, nil
		}
		if strings.TrimSpace(msg.text) != "" {
			m.events = append(m.events, msg.text)
		}
		return m, waitHostLobbyEventCmd(m.host)

	case onlineLobbyClientEventMsg:
		if !msg.ok {
			return m, nil
		}
		if strings.TrimSpace(msg.text) != "" {
			m.events = append(m.events, msg.text)
		}
		if m.cli != nil && m.cli.GameStarted() {
			m.startMatch = true
			return m, tea.Quit
		}
		return m, waitClientLobbyEventCmd(m.cli)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.back = true
			return m, tea.Quit
		case "enter":
			line := strings.TrimSpace(m.input)
			m.input = ""
			m.cursor = 0
			return m, m.handleSubmit(line)
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "right":
			if m.cursor < len([]rune(m.input)) {
				m.cursor++
			}
			return m, nil
		case "home", "ctrl+a":
			m.cursor = 0
			return m, nil
		case "end", "ctrl+e":
			m.cursor = len([]rune(m.input))
			return m, nil
		case "backspace", "ctrl+h":
			rs := []rune(m.input)
			if m.cursor > 0 && len(rs) > 0 {
				rs = append(rs[:m.cursor-1], rs[m.cursor:]...)
				m.input = string(rs)
				m.cursor--
			}
			return m, nil
		case "delete", "ctrl+d":
			rs := []rune(m.input)
			if m.cursor >= 0 && m.cursor < len(rs) {
				rs = append(rs[:m.cursor], rs[m.cursor+1:]...)
				m.input = string(rs)
			}
			return m, nil
		}
		if len(msg.Runes) > 0 {
			rs := []rune(m.input)
			in := msg.Runes
			out := make([]rune, 0, len(rs)+len(in))
			out = append(out, rs[:m.cursor]...)
			out = append(out, in...)
			out = append(out, rs[m.cursor:]...)
			m.input = string(out)
			m.cursor += len(in)
		}
		return m, nil
	}
	return m, nil
}

func (m *onlineLobbyModel) handleSubmit(line string) tea.Cmd {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" || lower == "refresh" {
		return nil
	}
	if lower == "back" {
		m.back = true
		return tea.Quit
	}
	m.err = nil
	switch m.mode {
	case lobbyModeHost:
		if lower == "start" {
			if err := m.host.StartGame(); err != nil {
				m.err = err
				return nil
			}
			m.startMatch = true
			return tea.Quit
		}
		if strings.HasPrefix(lower, "chat ") {
			m.host.SendHostChat(strings.TrimSpace(line[len("chat "):]))
			return nil
		}
	case lobbyModeClient:
		if strings.HasPrefix(lower, "chat ") {
			if err := m.cli.SendChat(strings.TrimSpace(line[len("chat "):])); err != nil {
				m.err = err
			}
			return nil
		}
	}
	m.err = errors.New(tr("lobby_unknown_command"))
	return nil
}

func (m onlineLobbyModel) renderInputWithCursor() string {
	rs := []rune(m.input)
	cursor := m.cursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(rs) {
		cursor = len(rs)
	}
	caret := "│"
	if (m.frame/2)%2 == 1 {
		caret = " "
	}
	return fmt.Sprintf("%s%s%s", string(rs[:cursor]), caret, string(rs[cursor:]))
}

func (m onlineLobbyModel) View() string {
	ensureThemeStyles()
	titleKey := tr("join_lobby_title")
	cmdKey := tr("join_commands")
	prompt := tr("join_prompt")
	lines := []string{}
	if m.mode == lobbyModeHost {
		titleKey = tr("host_lobby_title")
		cmdKey = tr("host_commands")
		prompt = tr("host_prompt")
	}

	lines = append(lines, headerStyle.AlignHorizontal(0.5).Render(tr("online_title")))
	lines = append(lines, lobbyHelpStyle.Render(tr("lobby_help")))
	lines = append(lines, "")
	lines = append(lines, titleKey)
	if m.mode == lobbyModeHost {
		lines = append(lines, tr("host_share_key"))
		lines = append(lines, m.key)
	}
	lines = append(lines, tr("line_sep"))

	var slots []string
	if m.mode == lobbyModeHost && m.host != nil {
		slots = m.host.Slots()
		lines = append(lines, fmt.Sprintf(tr("lobby_full_format"), m.host.IsFull()))
	}
	if m.mode == lobbyModeClient && m.cli != nil {
		slots = m.cli.Slots()
		lines = append(lines, fmt.Sprintf(tr("join_slot_format"), m.cli.AssignedSeat()+1))
	}
	for i, s := range slots {
		if s == "" {
			s = tr("slot_empty")
		}
		lines = append(lines, fmt.Sprintf(tr("slot_format"), i+1, s))
	}
	lines = append(lines, "")
	lines = append(lines, cmdKey)
	for _, line := range tail(m.events, 6) {
		lines = append(lines, "  "+clip(line, 96))
	}
	if m.err != nil {
		lines = append(lines, "")
		lines = append(lines, alertStyle.Foreground(lgRed).Render(tr("error_prefix")+m.err.Error()))
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("%s %s", prompt, m.renderInputWithCursor()))

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return "\n" + lobbyFrameStyle.Render(body) + "\n"
}
