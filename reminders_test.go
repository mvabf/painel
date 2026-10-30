package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestParseReminder(t *testing.T) {
	now := time.Date(2026, 10, 28, 14, 7, 30, 0, time.Local) // quarta
	at := func(mo time.Month, d, h, min int) time.Time { return time.Date(2026, mo, d, h, min, 0, 0, time.Local) }
	cases := []struct {
		in   string
		at   time.Time
		text string
	}{
		{"15:30 ligar pro fornecedor", at(10, 28, 15, 30), "ligar pro fornecedor"},
		{"9:05 daily", at(10, 29, 9, 5), "daily"},
		{"15h café", at(10, 28, 15, 0), "café"},
		{"15h45 café", at(10, 28, 15, 45), "café"},
		{"+20m alongar", at(10, 28, 14, 27), "alongar"},
		{"+1h30m deploy", at(10, 28, 15, 37), "deploy"},
		{"+10min água", at(10, 28, 14, 17), "água"},
		{"amanhã 9:00 revisar", at(10, 29, 9, 0), "revisar"},
		{"amanha 9h revisar", at(10, 29, 9, 0), "revisar"},
		{"hoje 18:00 sair", at(10, 28, 18, 0), "sair"},
		{"25/12 10:00 natal", at(12, 25, 10, 0), "natal"},
		{"01/02 8h férias", time.Date(2027, 2, 1, 8, 0, 0, 0, time.Local), "férias"},
	}
	for _, c := range cases {
		r, err := parseReminder(c.in, now)
		if err != nil {
			t.Errorf("parseReminder(%q): %v", c.in, err)
			continue
		}
		if !r.At.Equal(c.at) || r.Text != c.text {
			t.Errorf("parseReminder(%q) = %s %q, esperado %s %q", c.in, r.At, r.Text, c.at, c.text)
		}
	}

	for _, in := range []string{"", "15:30", "+20m", "amanhã", "ligar 15:30", "25:00 x", "31/02 9h x", "hoje 9:00 passou", "+abc x"} {
		if _, err := parseReminder(in, now); err == nil {
			t.Errorf("parseReminder(%q) deveria dar erro", in)
		}
	}
}

func TestLembretesNoArquivo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "lembretes")
	at := time.Date(2026, 10, 28, 15, 30, 0, 0, time.Local)
	rs := []reminder{
		{At: at.Add(time.Hour), Text: "depois"},
		{At: at, Text: "antes, com espaço", Fired: true},
	}
	sortReminders(rs)
	if err := saveReminders(path, rs); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	want := "! 2026-10-28 15:30 antes, com espaço\n- 2026-10-28 16:30 depois\n"
	if string(data) != want {
		t.Errorf("arquivo:\n%s\nesperado:\n%s", data, want)
	}

	os.WriteFile(path, append(data, []byte("lixo\n- 2026-13-01 10:00 data ruim\n- 2026-10-28 09:00 editado à mão\n")...), 0o644)
	got := loadReminders(path)
	if len(got) != 3 || got[0].Text != "editado à mão" || !got[1].Fired || got[2].Fired {
		t.Errorf("leitura: %+v", got)
	}
}

func TestLembreteToca(t *testing.T) {
	var rang []string
	alarm = func(title, body string) error { rang = append(rang, body); return nil }
	defer func() { alarm = desktopAlarm }()

	m := fakeModel(150)
	m.remindersFile = filepath.Join(t.TempDir(), "lembretes")
	m.reminders = []reminder{
		{At: m.now.Add(-time.Hour), Text: "venceu com o painel fechado"},
		{At: m.now.Add(time.Minute), Text: "daqui a pouco"},
	}

	next, cmd := m.Update(tickMsg(m.now))
	m = next.(model)
	runCmd(cmd)
	if strings.Join(rang, "|") != "venceu com o painel fechado" {
		t.Fatalf("deveria tocar só o vencido, tocou %q", rang)
	}

	next, cmd = m.Update(tickMsg(m.now.Add(time.Minute)))
	m = next.(model)
	runCmd(cmd)
	if strings.Join(rang, "|") != "venceu com o painel fechado|daqui a pouco" {
		t.Errorf("cada lembrete deveria tocar uma vez, tocou %q", rang)
	}
	if got := loadReminders(m.remindersFile); len(got) != 2 || !got[0].Fired || !got[1].Fired {
		t.Errorf("os avisados deveriam ficar marcados no arquivo: %+v", got)
	}
}

func typeKeys(m model, s string) model {
	for _, r := range s {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		if r == ' ' {
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{r}}
		}
		next, _ := m.Update(msg)
		m = next.(model)
	}
	return m
}

func TestCriarEConcluirLembrete(t *testing.T) {
	m := navModel()
	m.remindersFile = filepath.Join(t.TempDir(), "lembretes")

	m = typeKeys(m, "n")
	if !m.typing {
		t.Fatal("n deveria abrir o campo")
	}
	m = typeKeys(m, "16h ligar pra x")
	m, _ = press(m, "backspace")
	m = typeKeys(m, "q")
	if !strings.Contains(ansi.Strip(m.View()), "novo lembrete: 16h ligar pra q") {
		t.Errorf("o rodapé deveria mostrar o que foi digitado:\n%s", m.View())
	}

	m, _ = press(m, "enter")
	if m.typing || len(m.reminders) != 1 || m.reminders[0].Text != "ligar pra q" {
		t.Fatalf("enter deveria salvar o lembrete: %+v", m.reminders)
	}
	if got := loadReminders(m.remindersFile); len(got) != 1 {
		t.Errorf("o lembrete deveria ir para o arquivo: %+v", got)
	}
	if !strings.Contains(ansi.Strip(m.View()), "ligar pra q") {
		t.Error("o lembrete deveria aparecer na caixa")
	}

	m, _ = press(m, "up")
	if it, _ := m.selected(); it.Box != boxReminders {
		t.Fatalf("↑ do começo deveria cair no lembrete, veio %+v", it)
	}
	if _, cmd := press(m, "enter"); cmd != nil {
		t.Error("enter num lembrete sem link não deveria abrir nada")
	}
	if !strings.Contains(ansi.Strip(m.View()), "d exclui") {
		t.Error("um lembrete que ainda não tocou deveria mostrar d exclui")
	}
	m = typeKeys(m, "d")
	if len(m.reminders) != 0 || len(loadReminders(m.remindersFile)) != 0 {
		t.Errorf("d deveria excluir o lembrete: %+v", m.reminders)
	}
	if strings.Contains(ansi.Strip(m.View()), "Lembretes") {
		t.Error("sem lembretes, a caixa deveria sumir")
	}
}

func TestConcluirLembreteQueTocou(t *testing.T) {
	m := navModel()
	m.remindersFile = filepath.Join(t.TempDir(), "lembretes")
	m.reminders = []reminder{{At: m.now.Add(-time.Minute), Text: "café", Fired: true}}
	m, _ = press(m, "up")
	if it, _ := m.selected(); it.Box != boxReminders {
		t.Fatalf("↑ deveria cair no lembrete, veio %+v", it)
	}
	if !strings.Contains(ansi.Strip(m.View()), "x conclui") {
		t.Error("um lembrete que tocou deveria mostrar x conclui")
	}
	m = typeKeys(m, "x")
	if len(m.reminders) != 0 || m.selKey != "" {
		t.Errorf("x deveria concluir o lembrete: %+v", m.reminders)
	}
}

func TestCampoDeLembreteComErro(t *testing.T) {
	m := typeKeys(navModel(), "nq")
	m, _ = press(m, "enter")
	if !m.typing || m.inputErr == "" {
		t.Fatal("texto inválido deveria manter o campo aberto com o erro")
	}
	if strings.Contains(m.View(), "q sai") {
		t.Error("enquanto digita, o rodapé deveria ser o campo")
	}
	m, cmd := press(m, "esc")
	if m.typing || cmd != nil || len(m.reminders) != 0 {
		t.Error("esc deveria só fechar o campo")
	}
}

func TestReminderWhen(t *testing.T) {
	now := time.Date(2026, 10, 28, 14, 0, 0, 0, time.Local) // quarta
	cases := map[time.Time]string{
		time.Date(2026, 10, 28, 15, 30, 0, 0, time.Local): "15:30",
		time.Date(2026, 10, 29, 9, 0, 0, 0, time.Local):   "amanhã 09:00",
		time.Date(2026, 10, 31, 9, 0, 0, 0, time.Local):   "sáb 09:00",
		time.Date(2026, 12, 25, 9, 0, 0, 0, time.Local):   "25/12 09:00",
	}
	for at, want := range cases {
		if got := reminderWhen(at, now); got != want {
			t.Errorf("reminderWhen(%s) = %q, esperado %q", at, got, want)
		}
	}
}
