package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type lobbyMenuModel struct {
	cursor   int
	choices  []string
	selected string
}

var (
	lobbyFrameStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lgPurple).
			Background(lipgloss.Color("236")).
			Padding(1, 2)

	lobbyHelpStyle = lipgloss.NewStyle().
			Foreground(lgDim)

	lobbyOptionStyle = lipgloss.NewStyle().
				Foreground(lgWhite)

	lobbyOptionActiveStyle = lipgloss.NewStyle().
				Foreground(lgYellow).
				Bold(true)

	lobbyHintStyle = lipgloss.NewStyle().
			Foreground(lgGray)
)

func newLobbyMenuModel() lobbyMenuModel {
	return lobbyMenuModel{
		choices: []string{
			tr("menu_offline"),
			tr("menu_online"),
			tr("menu_language"),
		},
	}
}

func (m lobbyMenuModel) Init() tea.Cmd {
	return nil
}

func (m lobbyMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.selected = tr("lobby_choose_exit")
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "1":
			m.selected = tr("lobby_choose_offline")
			return m, tea.Quit
		case "2":
			m.selected = tr("lobby_choose_online")
			return m, tea.Quit
		case "3":
			m.selected = tr("lobby_choose_language")
			return m, tea.Quit
		case "0":
			m.selected = tr("lobby_choose_exit")
			return m, tea.Quit
		case "enter":
			switch m.cursor {
			case 0:
				m.selected = tr("lobby_choose_offline")
			case 1:
				m.selected = tr("lobby_choose_online")
			case 2:
				m.selected = tr("lobby_choose_language")
			default:
				m.selected = tr("lobby_choose_exit")
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m lobbyMenuModel) View() string {
	ensureThemeStyles()
	titleText := fitSingleLine(tr("app_title"), 42)
	title := headerStyle.Copy().
		Width(46).
		Align(lipgloss.Center).
		Render(titleText)
	lines := []string{
		title,
		lobbyHelpStyle.Render(tr("lobby_help")),
		"",
	}
	for i, c := range m.choices {
		row := fmt.Sprintf("%d) %s", i+1, c)
		if i == m.cursor {
			lines = append(lines, lobbyOptionActiveStyle.Render("▶ "+row))
			continue
		}
		lines = append(lines, lobbyOptionStyle.Render("  "+row))
	}
	lines = append(lines,
		lobbyOptionStyle.Render("  0) "+tr("menu_exit")),
		"",
		lobbyHintStyle.Render(tr("lobby_shortcuts")),
	)
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return "\n" + lobbyFrameStyle.Render(body) + "\n"
}
