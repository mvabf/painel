package main

import (
	"fmt"
	"time"
)

// demoModel é o painel com dados fictícios: aparece no `painel demo`, para
// ver a tela sem login em nada, e gera os quadros do GIF do README.
func requested(p pr, at time.Time) pr {
	p.RequestedAt = at
	return p
}

func withStatus(p pr, ci, review string, conflict bool) pr {
	p.CI, p.Review, p.Conflict = ci, review, conflict
	return p
}

func withLink(e event) event {
	e.Link = "https://meet.google.com/abc-defg-hij"
	return e
}

func demoModel(now time.Time, loaded bool) model {
	m := model{
		demo:  true,
		theme: defaultTheme,
		cfg: config{Name: "Ana", Agenda: "google",
			JiraURL: "https://exemplo.atlassian.net", JiraEmail: "ana@exemplo.com", JiraToken: "x"},
		now:           now,
		width:         112,
		height:        36,
		prsLoading:    !loaded,
		agendaLoading: !loaded,
		jiraLoading:   !loaded,
	}
	if !loaded {
		return m
	}

	at := func(h, min int) time.Time {
		return time.Date(now.Year(), now.Month(), now.Day(), h, min, 0, 0, time.Local)
	}
	p := func(repo string, n int, title string, draft bool) pr {
		return pr{Number: n, Title: title, IsDraft: draft, Repo: repo, CI: "SUCCESS",
			URL: fmt.Sprintf("https://github.com/exemplo/%s/pull/%d", repo, n)}
	}
	ev := func(h, min, dur int, title string) event {
		return event{Title: title, Start: at(h, min), End: at(h, min).Add(time.Duration(dur) * time.Minute)}
	}

	card := func(p pr, key string) pr {
		p.Keys = []string{key}
		return p
	}
	m.prs = prsResult{
		At: now,
		ToReview: []pr{
			requested(card(p("checkout", 482, "Corrige arredondamento do frete grátis", false), "APP-24"), now.Add(-4*24*time.Hour)),
			requested(p("api-pedidos", 1290, "Adiciona retry no webhook de pagamento", false), now.Add(-30*time.Hour)),
			requested(p("app-mobile", 77, "Nova tela de rastreio", false), now.Add(-107*time.Minute)),
		},
		ToReviewTotal: 3,
		MineTotal:     4,
		TeamReview:    12,
		Mine: []pr{
			withStatus(card(p("api-pedidos", 1287, "Cache na cotação de transportadoras", false), "APP-31"), "SUCCESS", "APPROVED", false),
			withStatus(card(p("checkout", 479, "Remove cupom expirado do carrinho", false), "APP-42"), "FAILURE", "CHANGES_REQUESTED", false),
			withStatus(p("api-pedidos", 1281, "Migra fila de notas fiscais", false), "PENDING", "", true),
			p("design-system", 341, "Tokens de cor para o tema escuro", true),
		},
	}
	m.cfg.JiraURL, m.cfg.JiraEmail, m.cfg.JiraToken = "https://exemplo.atlassian.net", "ana@exemplo.com", "x"
	m.jira = jiraResult{At: now, Issues: []issue{
		{Key: "APP-24", Summary: "Corrige cálculo do frete grátis", Status: "Em revisão", Category: "indeterminate"},
		{Key: "APP-31", Summary: "Cache na cotação", Status: "Fazendo", Category: "indeterminate"},
		{Key: "APP-40", Summary: "Tela de rastreio no app", Status: "A fazer", Category: "new"},
		{Key: "APP-42", Summary: "Remover cupom expirado", Status: "A fazer", Category: "new"},
	}}
	m.agenda = agendaResult{At: now, Events: []event{
		{Title: "Aniversário do Pedro", AllDay: true, Start: at(0, 0), End: at(24, 0)},
		ev(9, 30, 15, "Daily"),
		ev(11, 0, 60, "Planejamento da sprint"),
		ev(14, 0, 60, "Refinamento"),
		withLink(ev(15, 0, 45, "Pareamento com o time")),
		withLink(ev(17, 30, 30, "1:1 com a liderança")),
	}}
	return m
}
