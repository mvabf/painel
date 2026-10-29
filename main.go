package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// version é preenchida no build de release (-X main.version=...).
var version = "dev"

const usage = `painel: dashboard pessoal de terminal (PRs, agenda e relógio)

uso:
  painel          abre o dashboard
  painel auth     faz login no Google Calendar (veja docs/agenda.md)
  painel demo     abre o dashboard com dados fictícios, sem login em nada
  painel help     mostra esta ajuda
  painel version  mostra a versão

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
		case "demo":
			run(demoModel(time.Now(), true))
			return
		case "version", "-v", "--version":
			fmt.Println("painel", version)
			return
		case "help", "-h", "--help":
			fmt.Print(usage)
			return
		default:
			fmt.Fprint(os.Stderr, usage)
			os.Exit(2)
		}
	}

	run(newModel(loadConfig()))
}

func run(m model) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "painel:", err)
		os.Exit(1)
	}
}
