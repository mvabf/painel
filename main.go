package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "auth":
			if err := googleAuth(); err != nil {
				fmt.Fprintln(os.Stderr, "painel auth:", err)
				os.Exit(1)
			}
			return
		default:
			fmt.Fprintln(os.Stderr, "uso: painel [auth]")
			os.Exit(2)
		}
	}

	cfg := loadConfig()

	p := tea.NewProgram(newModel(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "painel:", err)
		os.Exit(1)
	}
}
