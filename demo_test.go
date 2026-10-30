package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

	// 14:59:53 → 15:00:19: carrega os dados e vira o minuto, quando a
	// reunião das 14h termina e a das 15h começa; as setas descem pelas
	// revisões até os meus PRs e, no fim, a tecla t passa pelos temas.
	defer applyTheme(builtinThemes[0])
	start := time.Date(2026, 10, 28, 14, 59, 53, 0, time.Local)
	loaded := demoModel(start.Add(3*time.Second), true)
	themes := len(builtinThemes) - 1
	for i := 0; i < 21+themes; i++ {
		now := start.Add(time.Duration(i) * time.Second)
		m := demoModel(now, false)
		if i >= 3 {
			m = loaded
			m.now = now
		}
		m.demo = false
		for step := 14; step <= min(i, 20); step += 2 {
			m = m.move(1)
		}
		th := builtinThemes[max(0, i-20)]
		applyTheme(th)
		m.theme = th.Name
		name := filepath.Join(dir, fmt.Sprintf("frame_%03d.ans", i))
		if err := os.WriteFile(name, []byte(m.View()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPainelDemo(t *testing.T) {
	m := demoModel(time.Now(), true)
	if !strings.Contains(m.View(), "demo com dados fictícios") {
		t.Error("o rodapé deveria avisar que é demo")
	}
	if _, cmd := press(m, "down", "enter"); cmd != nil {
		t.Error("na demo, enter não deveria abrir os links fictícios")
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd != nil || next.(model).prsLoading {
		t.Error("na demo, r não deveria buscar nada")
	}
}
