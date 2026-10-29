package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func navModel() model {
	m := fakeModel(150)
	m.prs.ToReview[0].URL = "https://github.com/acme/web-app/pull/4052"
	m.prs.ToReview[1].URL = "https://github.com/acme/app-mobile/pull/294"
	m.prs.Mine = []pr{{Number: 7, Title: "Meu PR", Repo: "web-app", URL: "https://github.com/acme/web-app/pull/7"}}
	m.cfg.JiraURL, m.cfg.JiraEmail, m.cfg.JiraToken = "https://exemplo.atlassian.net", "a", "b"
	m.jira = jiraResult{At: m.now, Issues: []issue{{Key: "APP-1", Summary: "Corrige frete", Category: "new"}}}
	m.agenda.Events[3].Link = "https://meet.google.com/abc-defg-hij"
	return m
}

func press(m model, keys ...string) (model, tea.Cmd) {
	var cmd tea.Cmd
	for _, k := range keys {
		var msg tea.KeyMsg
		switch k {
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "shift+tab":
			msg = tea.KeyMsg{Type: tea.KeyShiftTab}
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		}
		var next tea.Model
		next, cmd = m.Update(msg)
		m = next.(model)
	}
	return m, cmd
}

func TestItemsNaOrdemDaTela(t *testing.T) {
	var got []string
	for _, it := range navModel().items() {
		got = append(got, it.Box+" "+it.URL)
	}
	want := []string{
		"review https://github.com/acme/web-app/pull/4052",
		"review https://github.com/acme/app-mobile/pull/294",
		"mine https://github.com/acme/web-app/pull/7",
		"jira https://exemplo.atlassian.net/browse/APP-1",
		"agenda https://meet.google.com/abc-defg-hij",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("itens:\n%s\nesperado:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestNavegacao(t *testing.T) {
	m := navModel()

	m, _ = press(m, "down")
	if it, _ := m.selected(); it.Box != boxReview || !strings.HasSuffix(it.URL, "/4052") {
		t.Fatalf("↓ deveria começar no primeiro PR, veio %+v", it)
	}
	m, _ = press(m, "up")
	if it, _ := m.selected(); it.Box != boxAgenda {
		t.Errorf("↑ no primeiro deveria dar a volta para o último, veio %+v", it)
	}
	m, _ = press(m, "tab")
	if it, _ := m.selected(); it.Box != boxReview {
		t.Errorf("tab no último deveria voltar à primeira caixa, veio %+v", it)
	}
	m, _ = press(m, "tab", "tab")
	if it, _ := m.selected(); it.Box != boxJira {
		t.Errorf("dois tabs deveriam chegar no Jira, veio %+v", it)
	}
	m, _ = press(m, "shift+tab")
	if it, _ := m.selected(); it.Box != boxMine {
		t.Errorf("shift+tab deveria voltar para Meus PRs, veio %+v", it)
	}

	if _, cmd := press(m, "enter"); cmd == nil {
		t.Error("enter com seleção deveria abrir o link")
	}

	m, cmd := press(m, "esc")
	if m.selKey != "" || cmd != nil {
		t.Errorf("esc com seleção deveria só desmarcar")
	}
	if _, cmd := press(m, "enter"); cmd != nil {
		t.Error("enter sem seleção não deveria fazer nada")
	}
	if _, cmd := press(m, "esc"); cmd == nil {
		t.Error("esc sem seleção deveria sair")
	}
}

func TestSelecaoNaTela(t *testing.T) {
	m, _ := press(navModel(), "down")
	out := m.View()
	if !strings.Contains(out, "enter abre") {
		t.Errorf("o rodapé deveria mostrar enter abre:\n%s", out)
	}
	if !strings.Contains(m.mark(m.selKey, "x", 5), "x    ") {
		t.Error("a linha selecionada deveria ocupar a largura toda")
	}
	if got := m.mark("outra", "texto comprido", 5); ansi.Strip(got) != "text…" {
		t.Errorf("linha não selecionada deveria só ser cortada, veio %q", got)
	}
}

func TestSelecaoSomeQuandoItemSai(t *testing.T) {
	m, _ := press(navModel(), "down")
	m.prs.ToReview = m.prs.ToReview[1:]
	if _, i := m.selected(); i >= 0 {
		t.Error("o item saiu da lista, a seleção deveria sumir")
	}
	m, _ = press(m, "down")
	if it, _ := m.selected(); !strings.HasSuffix(it.URL, "/294") {
		t.Errorf("↓ depois disso deveria recomeçar do topo, veio %+v", it)
	}
}

func TestMeetingLink(t *testing.T) {
	cases := []struct {
		texts []string
		want  string
	}{
		{[]string{"https://meet.google.com/abc-defg-hij", "", "Sala 2", ""}, "https://meet.google.com/abc-defg-hij"},
		{[]string{"", "", "Sala 2", "Entre pela chamada: https://empresa.zoom.us/j/123?pwd=x\nOu ligue"}, "https://empresa.zoom.us/j/123?pwd=x"},
		{[]string{"", "", "", "Pauta em https://docs.exemplo.com/x e chamada em https://teams.microsoft.com/l/meetup-join/1"}, "https://teams.microsoft.com/l/meetup-join/1"},
		{[]string{"", "", "Sala 2", "Sem link"}, ""},
	}
	for _, c := range cases {
		if got := meetingLink(c.texts...); got != c.want {
			t.Errorf("meetingLink(%q) = %q, esperado %q", c.texts, got, c.want)
		}
	}
}
