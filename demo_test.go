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

func requested(p pr, at time.Time) pr {
	p.RequestedAt = at
	return p
}

func withStatus(p pr, ci, review string, conflict bool) pr {
	p.CI, p.Review, p.Conflict = ci, review, conflict
	return p
}

func demoModel(now time.Time, loaded bool) model {
	m := model{
		cfg: config{Name: "Ana", Agenda: "google",
			JiraURL: "https://exemplo.atlassian.net", JiraEmail: "ana@exemplo.com", JiraToken: "x"},
		now:           now,
		width:         112,
		height:        30,
		prsLoading:    !loaded,
		agendaLoading: !loaded,
		jiraLoading:   !loaded,
	}
	if !loaded {
		return m
	}

	at := func(h, min int) time.Time { return time.Date(2026, 10, 28, h, min, 0, 0, time.Local) }
	p := func(repo string, n int, title string, draft bool) pr {
		return pr{Number: n, Title: title, IsDraft: draft, Repo: repo, CI: "SUCCESS"}
	}
	ev := func(h, min, dur int, title string) event {
		return event{Title: title, Start: at(h, min), End: at(h, min).Add(time.Duration(dur) * time.Minute)}
	}

	m.prs = prsResult{
		At: at(14, 59),
		ToReview: []pr{
			requested(p("checkout", 482, "Corrige arredondamento do frete grátis", false), at(14, 59).Add(-4*24*time.Hour)),
			requested(p("api-pedidos", 1290, "Adiciona retry no webhook de pagamento", false), at(14, 59).Add(-30*time.Hour)),
			requested(p("app-mobile", 77, "Nova tela de rastreio", false), at(13, 12)),
		},
		ToReviewTotal: 3,
		MineTotal:     4,
		TeamReview:    12,
		Mine: []pr{
			withStatus(p("api-pedidos", 1287, "Cache na cotação de transportadoras", false), "SUCCESS", "APPROVED", false),
			withStatus(p("checkout", 479, "Remove cupom expirado do carrinho", false), "FAILURE", "CHANGES_REQUESTED", false),
			withStatus(p("api-pedidos", 1281, "Migra fila de notas fiscais", false), "PENDING", "", true),
			p("design-system", 341, "Tokens de cor para o tema escuro", true),
		},
	}
	m.cfg.JiraURL, m.cfg.JiraEmail, m.cfg.JiraToken = "https://exemplo.atlassian.net", "ana@exemplo.com", "x"
	m.jira = jiraResult{At: at(14, 59), Issues: []issue{
		{Key: "APP-24", Summary: "Corrige cálculo do frete grátis", Status: "Em revisão", Category: "indeterminate"},
		{Key: "APP-31", Summary: "Cache na cotação", Status: "Fazendo", Category: "indeterminate"},
		{Key: "APP-40", Summary: "Tela de rastreio no app", Status: "A fazer", Category: "new"},
		{Key: "APP-42", Summary: "Remover cupom expirado", Status: "A fazer", Category: "new"},
	}}
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
