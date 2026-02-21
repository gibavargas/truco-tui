package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *UIModel) clampChatCursor() {
	r := []rune(m.chatInput)
	if m.chatCursor < 0 {
		m.chatCursor = 0
	}
	if m.chatCursor > len(r) {
		m.chatCursor = len(r)
	}
}

func (m *UIModel) handleChatInputKey(msg tea.KeyMsg) (submitted string, handled bool) {
	if m.activeTab != "chat" {
		return "", false
	}
	m.clampChatCursor()

	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.chatInput)
		m.chatInput = ""
		m.chatCursor = 0
		return text, true
	case "left":
		if m.chatCursor > 0 {
			m.chatCursor--
		}
		return "", true
	case "right":
		if m.chatCursor < len([]rune(m.chatInput)) {
			m.chatCursor++
		}
		return "", true
	case "home", "ctrl+a":
		m.chatCursor = 0
		return "", true
	case "end", "ctrl+e":
		m.chatCursor = len([]rune(m.chatInput))
		return "", true
	case "backspace", "ctrl+h":
		rs := []rune(m.chatInput)
		if m.chatCursor > 0 && len(rs) > 0 {
			rs = append(rs[:m.chatCursor-1], rs[m.chatCursor:]...)
			m.chatInput = string(rs)
			m.chatCursor--
		}
		return "", true
	case "delete", "ctrl+d":
		rs := []rune(m.chatInput)
		if m.chatCursor >= 0 && m.chatCursor < len(rs) {
			rs = append(rs[:m.chatCursor], rs[m.chatCursor+1:]...)
			m.chatInput = string(rs)
		}
		return "", true
	}

	if len(msg.Runes) > 0 {
		rs := []rune(m.chatInput)
		in := msg.Runes
		out := make([]rune, 0, len(rs)+len(in))
		out = append(out, rs[:m.chatCursor]...)
		out = append(out, in...)
		out = append(out, rs[m.chatCursor:]...)
		m.chatInput = string(out)
		m.chatCursor += len(in)
		return "", true
	}

	return "", false
}

func (m UIModel) renderChatInputWithCursor() string {
	rs := []rune(m.chatInput)
	cursor := m.chatCursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(rs) {
		cursor = len(rs)
	}
	caret := "│"
	if (m.uiFrame/2)%2 == 1 {
		caret = " "
	}
	return fmt.Sprintf("%s%s%s", string(rs[:cursor]), caret, string(rs[cursor:]))
}

func appendChatLine(log []string, actor, text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return log
	}
	log = append(log, fmt.Sprintf("%s: %s", actor, text))
	if len(log) > 100 {
		log = log[len(log)-100:]
	}
	return log
}
