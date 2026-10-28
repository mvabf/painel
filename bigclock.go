package main

import "strings"

// Fonte de 5 linhas para o relógio grande. Cada glifo tem a mesma largura
// para o relógio não "pular" quando os dígitos mudam.
var glyphs = map[rune][5]string{
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
