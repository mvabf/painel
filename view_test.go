package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
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
		got := ansi.Strip(prLine(c.p, nil))
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
		if got := ansi.Strip(reviewPRLine(p, now, nil)); got != want {
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

func TestJiraKeys(t *testing.T) {
	got := jiraKeys("APP-24: corrige frete (UTF-8)", "app-24-corrige-frete")
	if strings.Join(got, ",") != "APP-24,UTF-8" {
		t.Errorf("got %q", got)
	}
	if got := jiraKeys("ajusta layout", "feature/app-77"); strings.Join(got, ",") != "APP-77" {
		t.Errorf("chave só na branch: got %q", got)
	}
	if got := jiraKeys("sem card", "main"); len(got) != 0 {
		t.Errorf("sem chave: got %q", got)
	}
}

func TestPRComCard(t *testing.T) {
	card := &issue{Key: "APP-24", Status: "Em revisão", Category: "indeterminate"}
	p := pr{Repo: "api", Number: 7, Title: "APP-24 Frete", CI: "SUCCESS", Review: "APPROVED", Keys: []string{"APP-24"}}
	if got := ansi.Strip(prLine(p, card)); got != "✓ api#7 aprovado Em revisão APP-24 Frete" {
		t.Errorf("prLine: got %q", got)
	}
	now := time.Date(2026, 10, 28, 15, 0, 0, 0, time.Local)
	p.RequestedAt = now.Add(-time.Hour)
	if got := ansi.Strip(reviewPRLine(p, now, card)); got != " 1h api#7 Em revisão APP-24 Frete" {
		t.Errorf("reviewPRLine: got %q", got)
	}

	cards := map[string]issue{"APP-24": *card}
	if cardFor(pr{Keys: []string{"UTF-8", "APP-24"}}, cards) == nil {
		t.Error("deveria achar o card pulando a chave que não está na busca")
	}
	if cardFor(pr{Keys: []string{"UTF-8"}}, cards) != nil {
		t.Error("chave fora da busca não deveria virar card")
	}
}

func TestIssueComPR(t *testing.T) {
	it := issue{Key: "APP-24", Summary: "Frete", Status: "Em revisão", Category: "indeterminate"}
	mine := []pr{
		{Number: 1, Keys: []string{"APP-9"}},
		{Number: 2, Keys: []string{"APP-24"}, CI: "FAILURE", Review: "APPROVED"},
	}
	linked := prFor(it, mine)
	if linked == nil || linked.Number != 2 {
		t.Fatalf("PR errado: %+v", linked)
	}
	if got := ansi.Strip(issueLine(it, linked)); got != "APP-24 Em revisão PR CI falhando Frete" {
		t.Errorf("CI falhando vem antes de aprovado: got %q", got)
	}
	if got := ansi.Strip(issueLine(it, nil)); got != "APP-24 Em revisão Frete" {
		t.Errorf("sem PR: got %q", got)
	}
	cases := map[string]pr{
		"PR aprovado": {Review: "APPROVED", CI: "SUCCESS"},
		"PR mudanças": {Review: "CHANGES_REQUESTED"},
		"PR aberto":   {CI: "PENDING"},
		"PR draft":    {IsDraft: true, CI: "FAILURE"},
	}
	for want, p := range cases {
		if got := ansi.Strip(prTag(p)); got != want {
			t.Errorf("prTag(%+v) = %q, esperado %q", p, got, want)
		}
	}
}

func TestListaAcompanhaAltura(t *testing.T) {
	m := fakeModel(150)
	for i := 0; i < 20; i++ {
		m.prs.ToReview = append(m.prs.ToReview, pr{Number: 100 + i, Title: "PR", Repo: "web-app", URL: fmt.Sprint(i), RequestedAt: m.now})
	}
	m.prs.ToReviewTotal = len(m.prs.ToReview)

	m.height = 0
	if got := m.listLimit(); got != maxListItems {
		t.Errorf("sem altura conhecida deveria usar %d, veio %d", maxListItems, got)
	}

	for _, h := range []int{30, 40, 60} {
		m.height = h
		limit := m.listLimit()
		if got := lipgloss.Height(m.View()); got > h {
			t.Errorf("altura %d: a tela ficou com %d linhas (limite %d)", h, got, limit)
		}
		m.height = h + 1
		if bigger := m.listLimit(); bigger < limit {
			t.Errorf("mais altura não pode mostrar menos itens: %d → %d", limit, bigger)
		}
	}

	m.height = 60
	if got := m.listLimit(); got < 15 {
		t.Errorf("com 60 linhas deveriam caber bem mais que 6 itens, veio %d", got)
	}
	m.height = 10
	if got := lipgloss.Height(m.View()); got > 10 {
		t.Errorf("num terminal baixo a tela deveria compactar até caber, ficou com %d linhas", got)
	}
}

func TestTimelineCells(t *testing.T) {
	at := func(h, min int) time.Time { return time.Date(2026, 10, 28, h, min, 0, 0, time.Local) }
	events := []event{
		{Title: "Feriado", AllDay: true, Start: at(0, 0), End: at(24, 0)},
		{Title: "Daily", Start: at(10, 0), End: at(11, 0)},
		{Title: "Refinamento", Start: at(14, 0), End: at(15, 0)},
	}
	cells := timelineCells(at(12, 15), events, at(9, 0), at(18, 0), 18)
	var got strings.Builder
	for _, c := range cells {
		got.WriteString(map[cell]string{cellFuture: ".", cellPast: "=", cellMeeting: "M", cellMeetingPast: "m", cellNow: "@"}[c])
	}
	if want := "==mm==@...MM......"; got.String() != want {
		t.Errorf("got %s, want %s", got.String(), want)
	}

	none := timelineCells(at(20, 0), nil, at(9, 0), at(18, 0), 9)
	for i, c := range none {
		if c != cellPast {
			t.Errorf("depois do expediente tudo já passou; célula %d = %v", i, c)
		}
	}
}

func TestHourLabels(t *testing.T) {
	if got := hourLabels(36); got != "9h  10h 11h 12h 13h 14h 15h 16h  18h" {
		t.Errorf("got %q", got)
	}
	for _, w := range []int{20, 36, 54, 108} {
		got := hourLabels(w)
		if !strings.HasPrefix(got, "9h") || !strings.HasSuffix(got, "18h") {
			t.Errorf("largura %d: deveria ir de 9h a 18h: %q", w, got)
		}
		if len(got) != w {
			t.Errorf("largura %d: rótulos passaram da linha: %q", w, got)
		}
		if strings.Contains(got, "h1") || strings.Contains(got, "hh") {
			t.Errorf("largura %d: rótulos encostados: %q", w, got)
		}
	}
}
