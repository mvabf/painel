package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestLoadTheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if got := loadTheme("Nord"); got.Name != "nord" || got.Border != "normal" {
		t.Errorf("tema pronto: %+v", got)
	}
	for _, name := range []string{"", "nao-existe", "../config"} {
		if got := loadTheme(name); got.Name != defaultTheme {
			t.Errorf("%q deveria cair no padrão, veio %q", name, got.Name)
		}
	}
}

func TestTemaPersonalizado(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "painel", "themes")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	body := "BASE=nord\nACCENT=#112233\nGREEN=#000000/#FFFFFF\nBORDER=double\n"
	if err := os.WriteFile(filepath.Join(dir, "meu"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	got := loadTheme("meu")
	nord := loadTheme("nord")
	if got.Name != "meu" || got.Accent != fixed("#112233") || got.Border != "double" {
		t.Errorf("valores do arquivo: %+v", got)
	}
	if got.Green != (lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}) {
		t.Errorf("cor claro/escuro: %+v", got.Green)
	}
	if got.Pink != nord.Pink || got.Clock != nord.Clock {
		t.Errorf("o que o arquivo não define deveria vir do BASE: %+v", got)
	}

	names := themeNames()
	if names[len(names)-1] != "meu" {
		t.Errorf("o tema do arquivo deveria vir depois dos prontos: %v", names)
	}
	if nextTheme("meu") != defaultTheme {
		t.Error("depois do último deveria voltar ao padrão")
	}
	if nextTheme(defaultTheme) != builtinThemes[1].Name {
		t.Error("t deveria seguir a ordem da lista")
	}
}

func TestTeclaTrocaTema(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	defer applyTheme(builtinThemes[0])

	m := fakeModel(150)
	m.theme = defaultTheme
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m = next.(model)
	if m.theme != "dracula" {
		t.Fatalf("t deveria ir para o dracula, foi para %q", m.theme)
	}
	if accent != builtinThemes[1].Accent {
		t.Error("as cores deveriam mudar junto")
	}
	if !strings.Contains(m.View(), "t tema: dracula") {
		t.Error("o rodapé deveria mostrar o tema atual")
	}

	applyTheme(loadTheme("nord"))
	if !strings.Contains(bigText("10"), "┃") {
		t.Error("o nord usa o relógio fino")
	}
	applyTheme(loadTheme("dracula"))
	if !strings.Contains(bigText("10"), "▓") || strings.Contains(bigText("10"), "█") {
		t.Error("o dracula usa o relógio sombreado")
	}
	applyTheme(builtinThemes[0])
	if !strings.Contains(bigText("10"), "█") {
		t.Error("o relógio sombreado não pode estragar o de blocos")
	}
}

func TestTodosOsTemasRenderizam(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	defer applyTheme(builtinThemes[0])
	for _, th := range builtinThemes {
		applyTheme(th)
		out := fakeModel(150).View()
		if !strings.Contains(out, "Revisões pendentes") || !strings.Contains(out, "18h") {
			t.Errorf("tema %s: tela incompleta", th.Name)
		}
	}
}
