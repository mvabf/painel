package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const usage = `painel: dashboard pessoal de terminal (PRs, agenda e relógio)

uso:
  painel          abre o dashboard
  painel auth     faz login no Google Calendar (veja docs/agenda.md)
  painel help     mostra esta ajuda

configuração: ~/.config/painel/config ou variáveis PAINEL_* (veja docs/configuracao.md)
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "auth":
			if err := googleAuth(); err != nil {
				fmt.Fprintln(os.Stderr, "painel auth:", err)
				os.Exit(1)
			}
			return
		case "help", "-h", "--help":
			fmt.Print(usage)
			return
		default:
			fmt.Fprint(os.Stderr, usage)
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
