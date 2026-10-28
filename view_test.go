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

func TestMineSummary(t *testing.T) {
	prs := []pr{
		{CI: "FAILURE", Review: "CHANGES_REQUESTED"},
		{CI: "SUCCESS", Review: "APPROVED"},
		{CI: "SUCCESS", Review: "APPROVED", Conflict: true},
		{CI: "FAILURE", IsDraft: true}, // draft não conta
	}
	want := "1 com CI falhando · 1 com mudanças pedidas · 1 com conflito · 2 aprovados"
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
