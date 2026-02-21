package ui

import (
	"strings"

	"truco-tui/internal/truco"
)

func getRelativePlayers(players []truco.Player, localIdx int) map[string]*truco.Player {
	m := make(map[string]*truco.Player)
	n := len(players)

	for i := 0; i < n; i++ {
		// Guardamos ponteiro para o elemento do slice (e não para cópia local)
		// para deixar explícito que o mapa referencia o snapshot atual.
		p := &players[i]
		dist := (i - localIdx + n) % n

		if n == 2 {
			if dist == 1 {
				m["top"] = p
			}
			continue
		}

		switch dist {
		case 1:
			m["right"] = p
		case 2:
			m["top"] = p
		case 3:
			m["left"] = p
		}
	}
	return m
}

func tail(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return items[len(items)-n:]
}

func clip(s string, width int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 3 {
		return string(r[:width])
	}
	return string(r[:width-3]) + "..."
}
