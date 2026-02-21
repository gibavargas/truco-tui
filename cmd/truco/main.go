package main

import (
	"fmt"
	"os"

	"truco-tui/internal/ui"
)

func main() {
	t := ui.New()
	if err := t.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}
