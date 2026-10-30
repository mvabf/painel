package main

import "strings"

// Fontes de 5 linhas para o relógio grande, escolhidas pelo tema. Cada glifo
// tem a mesma largura para o relógio não "pular" quando os dígitos mudam.
var clockFonts = map[string]map[rune][5]string{
	"blocks": {
		'0': {"███", "█ █", "█ █", "█ █", "███"},
		'1': {" █ ", "██ ", " █ ", " █ ", "███"},
		'2': {"███", "  █", "███", "█  ", "███"},
		'3': {"███", "  █", "███", "  █", "███"},
		'4': {"█ █", "█ █", "███", "  █", "  █"},
		'5': {"███", "█  ", "███", "  █", "███"},
		'6': {"███", "█  ", "███", "█ █", "███"},
		'7': {"███", "  █", "  █", "  █", "  █"},
		'8': {"███", "█ █", "███", "█ █", "███"},
		'9': {"███", "█ █", "███", "  █", "███"},
		':': {" ", "▪", " ", "▪", " "},
	},
	"slim": {
		'0': {"┏━┓", "┃ ┃", "┃ ┃", "┃ ┃", "┗━┛"},
		'1': {"  ╻", "  ┃", "  ┃", "  ┃", "  ╹"},
		'2': {"╺━┓", "  ┃", "┏━┛", "┃  ", "┗━╸"},
		'3': {"╺━┓", "  ┃", "╺━┫", "  ┃", "╺━┛"},
		'4': {"╻ ╻", "┃ ┃", "┗━┫", "  ┃", "  ╹"},
		'5': {"┏━╸", "┃  ", "┗━┓", "  ┃", "╺━┛"},
		'6': {"┏━╸", "┃  ", "┣━┓", "┃ ┃", "┗━┛"},
		'7': {"╺━┓", "  ┃", "  ┃", "  ┃", "  ╹"},
		'8': {"┏━┓", "┃ ┃", "┣━┫", "┃ ┃", "┗━┛"},
		'9': {"┏━┓", "┃ ┃", "┗━┫", "  ┃", "╺━┛"},
		':': {" ", "•", " ", "•", " "},
	},
}

func init() {
	shade := map[rune][5]string{}
	for r, g := range clockFonts["blocks"] {
		for i := range g {
			g[i] = strings.ReplaceAll(g[i], "█", "▓")
		}
		shade[r] = g
	}
	clockFonts["shade"] = shade
}

func bigText(s string) string {
	var rows [5]strings.Builder
	for i, r := range s {
		g, ok := glyphs[r]
		if !ok {
			continue
		}
		for row := range rows {
			if i > 0 {
				rows[row].WriteString(" ")
			}
			rows[row].WriteString(g[row])
		}
	}
	lines := make([]string, len(rows))
	for i := range rows {
		lines[i] = rows[i].String()
	}
	return strings.Join(lines, "\n")
}
