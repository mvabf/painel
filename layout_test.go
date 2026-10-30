package main

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func sizedDemo(w, h int) model {
	m := demoModel(time.Date(2026, 10, 28, 15, 0, 10, 0, time.Local), true)
	m.width, m.height = w, h
	return m
}

func TestCabeEmQualquerTamanho(t *testing.T) {
	for _, w := range []int{40, 50, 60, 70, 80, 100, 112, 140, 180, 250} {
		for _, h := range []int{15, 20, 24, 30, 40, 60} {
			for _, jira := range []bool{true, false} {
				m := sizedDemo(w, h)
				if !jira {
					m.cfg.JiraToken = ""
				}
				out := m.View()
				for i, line := range strings.Split(out, "\n") {
					if got := lipgloss.Width(line); got > w {
						t.Errorf("%dx%d: linha %d com %d colunas: %q", w, h, i, got, ansi.Strip(line))
					}
				}
				if got := lipgloss.Height(out); got > h {
					t.Errorf("%dx%d (jira %v): a tela ficou com %d linhas", w, h, jira, got)
				}
				if !strings.Contains(out, "q sai") {
					t.Errorf("%dx%d: o rodapé sumiu", w, h)
				}
			}
		}
	}
}

func TestCompactaAosPoucos(t *testing.T) {
	cases := []struct {
		w, h int
		want layout
	}{
		{112, 36, layout{}},
		{112, 31, layout{tight: true}},
		{100, 30, layout{tight: true, smallClock: true, noSummary: true}},
		{80, 24, layout{tight: true, smallClock: true, noSummary: true}},
		{60, 20, layout{tight: true, smallClock: true, tabs: true}},
	}
	for _, c := range cases {
		if got, _ := sizedDemo(c.w, c.h).plan(); got != c.want {
			t.Errorf("%dx%d: layout %+v, esperado %+v", c.w, c.h, got, c.want)
		}
	}

	small := ansi.Strip(sizedDemo(80, 24).View())
	if !strings.Contains(small, "15:00:10  Boa tarde, Ana · quarta, 28/10 · agora: Pareamento com o time") {
		t.Errorf("o relógio pequeno deveria ficar numa linha:\n%s", small)
	}
	if strings.Contains(small, "PRs às") {
		t.Error("com pouco espaço, os horários das buscas deveriam sair do rodapé")
	}
}

func TestAbas(t *testing.T) {
	m := sizedDemo(60, 20)
	out := ansi.Strip(m.View())
	if !strings.Contains(out, " Revisões  · Meus PRs · Jira · Agenda · Lembretes") || strings.Contains(out, "Meus PRs abertos") {
		t.Fatalf("deveria mostrar só a primeira caixa, com as abas:\n%s", out)
	}
	for _, it := range m.items() {
		if it.Box != boxReview {
			t.Errorf("as setas só deveriam andar pela caixa visível, veio %+v", it)
		}
	}

	m, _ = press(m, "down", "tab")
	if m.selKey != "" || !strings.Contains(ansi.Strip(m.View()), "Meus PRs abertos") {
		t.Errorf("tab deveria trocar para Meus PRs e limpar a seleção")
	}
	m, _ = press(m, "shift+tab", "shift+tab")
	if got := m.activeTab(); got != boxReminders {
		t.Errorf("shift+tab no começo deveria dar a volta para Lembretes, veio %s", got)
	}
	if !strings.Contains(ansi.Strip(m.View()), "tab troca de caixa") {
		t.Error("o rodapé deveria ensinar a trocar de caixa")
	}
}

func TestLinhasDaGradeAlinhadas(t *testing.T) {
	for _, w := range []int{100, 112, 113, 150} {
		m := sizedDemo(w, 60)
		for _, l := range strings.Split(ansi.Strip(m.View()), "\n") {
			if strings.Contains(l, "╭") {
				if got := len([]rune(strings.TrimSpace(l))); got != w-4 {
					t.Errorf("largura %d: toda linha da grade deveria ocupar %d colunas, esta ocupa %d: %q", w, w-4, got, l)
				}
			}
		}
	}
}

func TestCabecalhoEstreito(t *testing.T) {
	m := sizedDemo(64, 60)
	out := ansi.Strip(m.viewHeader())
	lines := strings.Split(out, "\n")
	if !strings.Contains(out, "Boa tarde, Ana") || !strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "expediente") {
		t.Errorf("sem largura, a linha do tempo deveria descer para baixo do relógio:\n%s", out)
	}
	for _, l := range lines {
		if w := lipgloss.Width(l); w > 60 {
			t.Errorf("linha com %d colunas: %q", w, l)
		}
	}
}
