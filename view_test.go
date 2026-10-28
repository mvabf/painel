package main

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func fakeModel(width int) model {
	now := time.Date(2026, 10, 28, 14, 7, 30, 0, time.Local)
	p := func(repo string, n int, title string) pr {
		return pr{Number: n, Title: title, UpdatedAt: now, Repo: repo}
	}
	ev := func(h, min, dur int, title string) event {
		s := time.Date(2026, 10, 28, h, min, 0, 0, time.Local)
		return event{Title: title, Start: s, End: s.Add(time.Duration(dur) * time.Minute)}
	}
	return model{
		cfg:    config{Name: "Ana", Agenda: "ics"},
		now:    now,
		width:  width,
		height: 40,
		prs: prsResult{
			At:         now,
			ToReview:   []pr{p("web-app", 4052, "Corrige retry do webhook"), p("app-mobile", 294, "Nova tela de onboarding")},
			TeamReview: 52,
			Mine:       []pr{},
		},
		agenda: agendaResult{At: now, Events: []event{
			{Title: "Feriado", AllDay: true},
			ev(9, 30, 15, "Daily"),
			ev(14, 0, 30, "Refinamento"),
			ev(16, 0, 60, "1:1"),
		}},
	}
}

func TestViewRenders(t *testing.T) {
	for _, w := range []int{160, 80} {
		out := fakeModel(w).View()
		for _, want := range []string{"Boa tarde, Ana", "quarta-feira, 28 de outubro", "Revisões pendentes", "web-app#4052", "nada aqui", "+52 pedidos aos seus times", "Refinamento", "agora: "} {
			if !strings.Contains(out, want) {
				t.Errorf("largura %d: faltou %q na tela:\n%s", w, want, out)
			}
		}
		if testing.Verbose() {
			t.Log("\n" + out)
		}
	}
}

func TestViewComJira(t *testing.T) {
	m := fakeModel(150)
	m.cfg.JiraURL, m.cfg.JiraEmail, m.cfg.JiraToken = "https://x", "a", "b"
	m.jira = jiraResult{At: m.now, Issues: []issue{
		{Key: "APP-1", Summary: "Corrige frete", Status: "Em revisão", Category: "indeterminate"},
		{Key: "APP-2", Summary: "Tela de login", Status: "A fazer", Category: "new"},
	}}
	out := m.View()
	for _, want := range []string{"2 Jira", "1 em andamento · 1 a fazer", "APP-1 Em revisão Corrige frete", "APP-2 Tela de login", "j abre o Jira"} {
		if !strings.Contains(out, want) {
			t.Errorf("faltou %q na tela:\n%s", want, out)
		}
	}
}

func TestViewLoading(t *testing.T) {
	m := model{cfg: config{Agenda: "ics"}, now: time.Now(), width: 160, prsLoading: true, agendaLoading: true}
	out := m.View()
	for _, want := range []string{"… Revisões pendentes", "… Meus PRs abertos", "… Agenda de hoje"} {
		if !strings.Contains(out, want) {
			t.Errorf("antes da primeira busca deveria mostrar %q:\n%s", want, out)
		}
	}
}

func TestHumanDuration(t *testing.T) {
	cases := map[time.Duration]string{12 * time.Minute: "12 min", time.Hour: "1h", 90 * time.Minute: "1h30"}
	for d, want := range cases {
		if got := humanDuration(d); got != want {
			t.Errorf("humanDuration(%v) = %q, quero %q", d, got, want)
		}
	}
}

func TestPRLine(t *testing.T) {
	base := pr{Repo: "api", Number: 7, Title: "Título"}
	cases := []struct {
		name string
		p    pr
		want []string
	}{
		{"CI ok e aprovado", pr{Repo: "api", Number: 7, Title: "Título", CI: "SUCCESS", Review: "APPROVED"}, []string{"✓ api#7 aprovado Título"}},
		{"CI falhando, mudanças e conflito", pr{Repo: "api", Number: 7, Title: "Título", CI: "FAILURE", Review: "CHANGES_REQUESTED", Conflict: true}, []string{"✗ api#7 conflito mudanças Título"}},
		{"CI rodando", pr{Repo: "api", Number: 7, Title: "Título", CI: "PENDING"}, []string{"● api#7 Título"}},
		{"sem checks", base, []string{"· api#7 Título"}},
		{"draft ignora status", pr{Repo: "api", Number: 7, Title: "Título", IsDraft: true, CI: "FAILURE"}, []string{"· api#7 [draft] Título"}},
	}
	for _, c := range cases {
		got := ansi.Strip(prLine(c.p))
		if got != c.want[0] {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want[0])
		}
	}
}

func TestWrapParts(t *testing.T) {
	got := wrapParts("1 CI falhando · 1 com mudanças · 1 com conflito · 2 aprovados", 30)
	var plain []string
	for _, l := range got {
		plain = append(plain, ansi.Strip(l))
	}
	want := []string{"1 CI falhando · 1 com mudanças", "1 com conflito · 2 aprovados"}
	if strings.Join(plain, "|") != strings.Join(want, "|") {
		t.Errorf("got %q, want %q", plain, want)
	}
}

func TestMineSummary(t *testing.T) {
	prs := []pr{
		{CI: "FAILURE", Review: "CHANGES_REQUESTED"},
		{CI: "SUCCESS", Review: "APPROVED"},
		{CI: "SUCCESS", Review: "APPROVED", Conflict: true},
		{CI: "FAILURE", IsDraft: true}, // draft não conta
	}
	want := "1 CI falhando · 1 com mudanças · 1 com conflito · 2 aprovados"
	if got := mineSummary(prs); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	if got := mineSummary([]pr{{CI: "SUCCESS"}}); got != "" {
		t.Errorf("sem nada a destacar deveria ser vazio, veio %q", got)
	}
}

func TestReviewPRLine(t *testing.T) {
	now := time.Date(2026, 10, 28, 15, 0, 0, 0, time.Local)
	p := pr{Repo: "api", Number: 7, Title: "Título"}
	cases := map[time.Duration]string{
		20 * time.Minute:    "20m api#7 Título",
		5 * time.Hour:       " 5h api#7 Título",
		50 * time.Hour:      " 2d api#7 Título",
		12 * 24 * time.Hour: "12d api#7 Título",
	}
	for ago, want := range cases {
		p.RequestedAt = now.Add(-ago)
		if got := ansi.Strip(reviewPRLine(p, now)); got != want {
			t.Errorf("pedido há %v: got %q, want %q", ago, got, want)
		}
	}
}

func TestParseAge(t *testing.T) {
	cases := map[string]time.Duration{
		"7d": 7 * 24 * time.Hour, "1d": 24 * time.Hour, "36h": 36 * time.Hour, "90m": 90 * time.Minute,
		"": 0, "0d": 0, "-3d": 0, "sete": 0, "7 d": 0, "7.5d": 0,
	}
	for in, want := range cases {
		if got := parseAge(in); got != want {
			t.Errorf("parseAge(%q) = %v, quero %v", in, got, want)
		}
	}
}

func TestAgeLabel(t *testing.T) {
	cases := map[time.Duration]string{24 * time.Hour: "1 dia", 7 * 24 * time.Hour: "7 dias", 36 * time.Hour: "36h", 90 * time.Minute: "1h30m", 45 * time.Minute: "45m", 30 * time.Second: "30s"}
	for d, want := range cases {
		if got := ageLabel(d); got != want {
			t.Errorf("ageLabel(%v) = %q, quero %q", d, got, want)
		}
	}
}

func TestReviewMaxAge(t *testing.T) {
	m := fakeModel(160)
	m.prs.ToReview = []pr{
		{Repo: "velho", Number: 1, Title: "Antigo", RequestedAt: m.now.Add(-30 * 24 * time.Hour)},
		{Repo: "medio", Number: 2, Title: "Oito dias", RequestedAt: m.now.Add(-8 * 24 * time.Hour)},
		{Repo: "novo", Number: 3, Title: "Ontem", RequestedAt: m.now.Add(-20 * time.Hour)},
	}
	m.prs.ToReviewTotal = 3

	out := ansi.Strip(m.View())
	if !strings.Contains(out, "3 Revisões pendentes") || !strings.Contains(out, "velho#1") {
		t.Fatalf("sem limite deveria mostrar tudo:\n%s", out)
	}

	m.cfg.ReviewMaxAge = 7 * 24 * time.Hour
	out = ansi.Strip(m.View())
	for _, want := range []string{"1 Revisões pendentes", "novo#3", "ocultos: 2 com mais de 7 dias"} {
		if !strings.Contains(out, want) {
			t.Errorf("faltou %q:\n%s", want, out)
		}
	}
	for _, bad := range []string{"velho#1", "medio#2"} {
		if strings.Contains(out, bad) {
			t.Errorf("%q deveria estar oculto", bad)
		}
	}
}
