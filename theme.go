package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// theme é uma skin: as cores, o tipo de borda das caixas e a fonte do relógio.
type theme struct {
	Name   string
	Accent lipgloss.AdaptiveColor
	Pink   lipgloss.AdaptiveColor
	Green  lipgloss.AdaptiveColor
	Yellow lipgloss.AdaptiveColor
	Blue   lipgloss.AdaptiveColor
	Subtle lipgloss.AdaptiveColor
	Err    lipgloss.AdaptiveColor
	SelFg  lipgloss.AdaptiveColor // texto da linha selecionada, sobre o Accent
	Border string                 // rounded, normal, thick, double, ascii ou hidden
	Clock  string                 // blocks, slim ou shade
}

func fixed(hex string) lipgloss.AdaptiveColor { return lipgloss.AdaptiveColor{Light: hex, Dark: hex} }

const defaultTheme = "default"

var builtinThemes = []theme{
	{
		Name:   defaultTheme,
		Accent: lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#B49CFF"},
		Pink:   lipgloss.AdaptiveColor{Light: "#D6336C", Dark: "#FF7AB2"},
		Green:  lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#69DB7C"},
		Yellow: lipgloss.AdaptiveColor{Light: "#E67700", Dark: "#FFD43B"},
		Blue:   lipgloss.AdaptiveColor{Light: "#1971C2", Dark: "#74C0FC"},
		Subtle: lipgloss.AdaptiveColor{Light: "#868E96", Dark: "#6C7086"},
		Err:    lipgloss.AdaptiveColor{Light: "#C92A2A", Dark: "#FF6B6B"},
		SelFg:  lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1E1E2E"},
		Border: "rounded",
		Clock:  "blocks",
	},
	{
		Name:   "dracula",
		Accent: fixed("#BD93F9"), Pink: fixed("#FF79C6"), Green: fixed("#50FA7B"), Yellow: fixed("#F1FA8C"),
		Blue: fixed("#8BE9FD"), Subtle: fixed("#6272A4"), Err: fixed("#FF5555"), SelFg: fixed("#282A36"),
		Border: "rounded", Clock: "shade",
	},
	{
		Name:   "nord",
		Accent: fixed("#88C0D0"), Pink: fixed("#B48EAD"), Green: fixed("#A3BE8C"), Yellow: fixed("#EBCB8B"),
		Blue: fixed("#81A1C1"), Subtle: fixed("#616E88"), Err: fixed("#BF616A"), SelFg: fixed("#2E3440"),
		Border: "normal", Clock: "slim",
	},
	{
		Name:   "gruvbox",
		Accent: fixed("#FE8019"), Pink: fixed("#D3869B"), Green: fixed("#B8BB26"), Yellow: fixed("#FABD2F"),
		Blue: fixed("#83A598"), Subtle: fixed("#7C6F64"), Err: fixed("#FB4934"), SelFg: fixed("#282828"),
		Border: "thick", Clock: "blocks",
	},
	{
		Name:   "tokyo-night",
		Accent: fixed("#7AA2F7"), Pink: fixed("#BB9AF7"), Green: fixed("#9ECE6A"), Yellow: fixed("#E0AF68"),
		Blue: fixed("#7DCFFF"), Subtle: fixed("#565F89"), Err: fixed("#F7768E"), SelFg: fixed("#1A1B26"),
		Border: "rounded", Clock: "slim",
	},
	{
		Name:   "solarized",
		Accent: fixed("#6C71C4"), Pink: fixed("#D33682"), Green: fixed("#859900"), Yellow: fixed("#B58900"),
		Blue:   fixed("#268BD2"),
		Subtle: lipgloss.AdaptiveColor{Light: "#93A1A1", Dark: "#586E75"},
		Err:    fixed("#DC322F"),
		SelFg:  lipgloss.AdaptiveColor{Light: "#FDF6E3", Dark: "#002B36"},
		Border: "double", Clock: "blocks",
	},
	{
		Name:   "mono",
		Accent: lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"},
		Pink:   lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"},
		Green:  lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"},
		Yellow: lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"},
		Blue:   lipgloss.AdaptiveColor{Light: "#333333", Dark: "#DDDDDD"},
		Subtle: lipgloss.AdaptiveColor{Light: "#888888", Dark: "#777777"},
		Err:    lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"},
		SelFg:  lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#000000"},
		Border: "ascii", Clock: "blocks",
	},
}

var (
	accent, pink, green, yellow, blue, subtle, errColr lipgloss.AdaptiveColor

	clockStyle, greetStyle, dimStyle, errStyle, boxStyle lipgloss.Style
	okStyle, failStyle, waitStyle, cardStyle, selStyle   lipgloss.Style
	glyphs                                               map[rune][5]string
)

func init() { applyTheme(builtinThemes[0]) }

// applyTheme troca as cores e estilos usados pela tela. Só é chamada no
// início e pela tecla t, sempre na goroutine do Bubble Tea.
func applyTheme(t theme) {
	accent, pink, green, yellow, blue, subtle, errColr = t.Accent, t.Pink, t.Green, t.Yellow, t.Blue, t.Subtle, t.Err

	clockStyle = lipgloss.NewStyle().Foreground(accent)
	greetStyle = lipgloss.NewStyle().Bold(true)
	dimStyle = lipgloss.NewStyle().Foreground(subtle)
	errStyle = lipgloss.NewStyle().Foreground(errColr)
	boxStyle = lipgloss.NewStyle().
		Border(borderByName(t.Border)).
		BorderForeground(subtle).
		Padding(0, 1)
	okStyle = lipgloss.NewStyle().Foreground(green)
	failStyle = lipgloss.NewStyle().Foreground(errColr)
	waitStyle = lipgloss.NewStyle().Foreground(yellow)
	cardStyle = lipgloss.NewStyle().Foreground(blue)
	selStyle = lipgloss.NewStyle().Foreground(t.SelFg).Background(accent)
	glyphs = clockFonts[t.Clock]
	if glyphs == nil {
		glyphs = clockFonts["blocks"]
	}
}

func borderByName(name string) lipgloss.Border {
	switch name {
	case "normal":
		return lipgloss.NormalBorder()
	case "thick":
		return lipgloss.ThickBorder()
	case "double":
		return lipgloss.DoubleBorder()
	case "ascii":
		return lipgloss.ASCIIBorder()
	case "hidden":
		return lipgloss.HiddenBorder()
	}
	return lipgloss.RoundedBorder()
}

func themesDir() string { return filepath.Join(configDir(), "themes") }

// themeNames lista os temas prontos e, depois, os arquivos de
// ~/.config/painel/themes, na ordem em que a tecla t passa por eles.
func themeNames() []string {
	var names []string
	for _, t := range builtinThemes {
		names = append(names, t.Name)
	}
	entries, _ := os.ReadDir(themesDir())
	var custom []string
	for _, e := range entries {
		if !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			custom = append(custom, e.Name())
		}
	}
	sort.Strings(custom)
	return append(names, custom...)
}

// loadTheme acha um tema pronto pelo nome ou lê um arquivo de
// ~/.config/painel/themes. Nome desconhecido cai no padrão.
func loadTheme(name string) theme {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, t := range builtinThemes {
		if t.Name == name {
			return t
		}
	}
	if name == "" || strings.ContainsAny(name, `/\`) {
		return builtinThemes[0]
	}
	vals := readKV(filepath.Join(themesDir(), name))
	if len(vals) == 0 {
		return builtinThemes[0]
	}
	return customTheme(name, vals)
}

// customTheme parte do tema BASE (o padrão, se não houver) e troca o que o
// arquivo definir. Cores aceitam "#RRGGBB" ou "#claro/#escuro".
func customTheme(name string, vals map[string]string) theme {
	t := loadTheme(vals["BASE"])
	t.Name = name
	for key, dst := range map[string]*lipgloss.AdaptiveColor{
		"ACCENT": &t.Accent, "PINK": &t.Pink, "GREEN": &t.Green, "YELLOW": &t.Yellow,
		"BLUE": &t.Blue, "SUBTLE": &t.Subtle, "ERROR": &t.Err, "SELECTED_TEXT": &t.SelFg,
	} {
		if v := vals[key]; v != "" {
			light, dark, ok := strings.Cut(v, "/")
			if !ok {
				dark = light
			}
			*dst = lipgloss.AdaptiveColor{Light: strings.TrimSpace(light), Dark: strings.TrimSpace(dark)}
		}
	}
	if v := vals["BORDER"]; v != "" {
		t.Border = v
	}
	if v := vals["CLOCK"]; v != "" {
		t.Clock = v
	}
	return t
}

// nextTheme é o tema seguinte na lista, dando a volta no fim.
func nextTheme(current string) string {
	names := themeNames()
	for i, n := range names {
		if n == current {
			return names[(i+1)%len(names)]
		}
	}
	return names[0]
}

// printThemes mostra cada tema com uma amostra das cores, da borda e do relógio.
func printThemes() {
	for _, name := range themeNames() {
		t := loadTheme(name)
		applyTheme(t)
		var swatch []string
		for _, c := range []lipgloss.AdaptiveColor{t.Accent, t.Pink, t.Green, t.Yellow, t.Blue, t.Subtle, t.Err} {
			swatch = append(swatch, lipgloss.NewStyle().Foreground(c).Render("███"))
		}
		sample := boxStyle.Render(clockStyle.Render(bigText("12:34")))
		info := lipgloss.JoinVertical(lipgloss.Left,
			greetStyle.Render(name),
			dimStyle.Render("borda "+t.Border+" · relógio "+t.Clock),
			"",
			strings.Join(swatch, " "),
		)
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Top, sample, "  ", info))
	}
	fmt.Println(dimStyle.Render("Escolha com THEME=<nome> em ~/.config/painel/config, ou troque com t no painel."))
}
