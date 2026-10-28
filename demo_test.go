package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// TestDemoFrames gera os quadros do GIF do README (docs/demo.gif) com dados
// fictícios. Só roda quando PAINEL_DEMO_DIR está definido; veja `make demo-gif`.
func TestDemoFrames(t *testing.T) {
	dir := os.Getenv("PAINEL_DEMO_DIR")
	if dir == "" {
		t.Skip("defina PAINEL_DEMO_DIR para gerar os quadros do GIF")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)

	// 14:59:53 → 15:00:06: carrega os dados e vira o minuto, quando a
	// reunião das 14h termina e a das 15h começa.
	start := time.Date(2026, 10, 28, 14, 59, 53, 0, time.Local)
	for i := 0; i < 14; i++ {
		m := demoModel(start.Add(time.Duration(i)*time.Second), i >= 3)
		name := filepath.Join(dir, fmt.Sprintf("frame_%03d.ans", i))
		if err := os.WriteFile(name, []byte(m.View()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func demoModel(now time.Time, loaded bool) model {
	m := model{
		cfg:           config{Name: "Ana", Agenda: "google"},
		now:           now,
		width:         112,
		height:        30,
		prsLoading:    !loaded,
		agendaLoading: !loaded,
	}
	if !loaded {
		return m
	}

	at := func(h, min int) time.Time { return time.Date(2026, 10, 28, h, min, 0, 0, time.Local) }
	p := func(repo string, n int, title string, draft bool) pr {
		x := pr{Number: n, Title: title, IsDraft: draft}
		x.Repository.Name = repo
		return x
	}
	ev := func(h, min, dur int, title string) event {
		return event{Title: title, Start: at(h, min), End: at(h, min).Add(time.Duration(dur) * time.Minute)}
	}

	m.prs = prsResult{
		At: at(14, 59),
		ToReview: []pr{
			p("checkout", 482, "Corrige arredondamento do frete grátis", false),
			p("api-pedidos", 1290, "Adiciona retry no webhook de pagamento", false),
			p("app-mobile", 77, "Nova tela de rastreio", false),
		},
		TeamReview: 12,
		Mine: []pr{
			p("api-pedidos", 1287, "Cache na cotação de transportadoras", false),
			p("design-system", 341, "Tokens de cor para o tema escuro", true),
		},
	}
	m.agenda = agendaResult{At: at(14, 59), Events: []event{
		{Title: "Aniversário do Pedro", AllDay: true, Start: at(0, 0), End: at(24, 0)},
		ev(9, 30, 15, "Daily"),
		ev(11, 0, 60, "Planejamento da sprint"),
		ev(14, 0, 60, "Refinamento"),
		ev(15, 0, 45, "Pareamento com o time"),
		ev(17, 30, 30, "1:1 com a liderança"),
	}}
	return m
}
