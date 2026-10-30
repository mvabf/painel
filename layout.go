package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// layout diz o quanto a tela foi compactada para caber no terminal. O valor
// zero é a tela completa; cada campo liga um passo de compactação.
type layout struct {
	tight      bool // sem linhas em branco entre as partes nem margem
	smallClock bool // relógio numa linha só, no lugar dos dígitos grandes
	noSummary  bool // caixas sem as linhas de resumo
	tabs       bool // uma caixa por vez, trocada com tab
}

const (
	minBoxWidth = 34
	boxGap      = 2
)

func (m model) View() string {
	if m.width == 0 {
		return ""
	}
	lay, limit := m.plan()
	m.lay = lay
	return m.render(limit)
}

func (m model) render(limit int) string {
	body := m.viewBody(limit)
	if m.lay.tight {
		out := lipgloss.JoinVertical(lipgloss.Left, m.viewHeader(), body, m.viewFooter())
		return lipgloss.NewStyle().Padding(0, 2).Render(out)
	}
	out := lipgloss.JoinVertical(lipgloss.Left, m.viewHeader(), "", body, "", m.viewFooter())
	return lipgloss.NewStyle().Padding(1, 2).Render(out)
}

// plan escolhe o layout menos compacto em que a tela cabe na altura do
// terminal e quantos itens cada lista mostra nele. Os passos, em ordem: tirar
// os espaços, encolher o relógio, tirar os resumos, baixar as listas até 2
// itens e, por fim, mostrar uma caixa por vez.
func (m model) plan() (layout, int) {
	if m.height == 0 {
		return layout{}, maxListItems
	}
	steps := []struct {
		lay   layout
		floor int
	}{
		{layout{}, minListItems},
		{layout{tight: true}, minListItems},
		{layout{tight: true, smallClock: true}, minListItems},
		{layout{tight: true, smallClock: true, noSummary: true}, minListItems},
		{layout{tight: true, smallClock: true, noSummary: true}, 2},
		{layout{tight: true, smallClock: true, tabs: true}, minListItems},
		{layout{tight: true, smallClock: true, tabs: true, noSummary: true}, 1},
	}
	var limit int
	for _, s := range steps {
		var ok bool
		if limit, ok = m.fitLimit(s.lay, s.floor); ok {
			return s.lay, limit
		}
	}
	return steps[len(steps)-1].lay, limit
}

// fitLimit mede a tela com 6 itens por lista, distribui a sobra de altura
// entre as linhas da grade e confere se coube, sem descer abaixo de floor.
func (m model) fitLimit(lay layout, floor int) (int, bool) {
	m.lay = lay
	fixed := lipgloss.Height(m.render(0)) - lipgloss.Height(m.viewBody(0))
	rows := len(m.gridRows())

	slack := m.height - fixed - lipgloss.Height(m.viewBody(maxListItems))
	limit := max(floor, min(maxFetched, maxListItems+slack/rows))
	for limit > floor && fixed+lipgloss.Height(m.viewBody(limit)) > m.height {
		limit--
	}
	return limit, fixed+lipgloss.Height(m.viewBody(limit)) <= m.height
}

// listLimit é quantos itens cada lista mostra no tamanho atual.
func (m model) listLimit() int {
	_, limit := m.plan()
	return limit
}

func (m model) boxCount() int { return len(m.boxKeys()) }

// boxKeys são as caixas na ordem da tela.
func (m model) boxKeys() []string {
	keys := []string{boxReview, boxMine}
	if m.cfg.jiraEnabled() {
		keys = append(keys, boxJira)
	}
	return append(keys, boxAgenda, boxReminders)
}

var boxTitles = map[string]string{
	boxReview:    "Revisões",
	boxMine:      "Meus PRs",
	boxJira:      "Jira",
	boxAgenda:    "Agenda",
	boxReminders: "Lembretes",
}

// activeTab é a caixa visível no modo de abas.
func (m model) activeTab() string {
	keys := m.boxKeys()
	return keys[((m.tab%len(keys))+len(keys))%len(keys)]
}

// gridRows reparte as caixas visíveis em linhas da grade.
func (m model) gridRows() [][]string {
	if m.lay.tabs {
		return [][]string{{m.activeTab()}}
	}
	keys := m.boxKeys()
	cols, _ := gridLayout(m.width-4, len(keys))
	var rows [][]string
	for start := 0; start < len(keys); start += cols {
		rows = append(rows, keys[start:min(start+cols, len(keys))])
	}
	return rows
}

// viewBody monta a grade de caixas, cada lista com até limit itens. Uma
// linha com menos caixas que as outras estica as suas até a largura toda.
func (m model) viewBody(limit int) string {
	var rows []string
	if m.lay.tabs {
		rows = append(rows, m.viewTabBar())
	}
	for _, row := range m.gridRows() {
		// A sobra da divisão vai para as primeiras caixas, para toda linha
		// terminar na mesma coluna.
		free := m.width - 4 - boxGap*(len(row)-1)
		widths := make([]int, len(row))
		contents := make([]string, len(row))
		h := 0
		for i, key := range row {
			widths[i] = free / len(row)
			if i < free%len(row) {
				widths[i]++
			}
			contents[i] = m.viewBox(key, widths[i], limit)
			h = max(h, lipgloss.Height(contents[i]))
		}
		var parts []string
		for i, c := range contents {
			if i > 0 {
				parts = append(parts, strings.Repeat(" ", boxGap))
			}
			box := boxStyle.Width(widths[i] - 2)
			if len(row) > 1 {
				box = box.Height(h)
			}
			parts = append(parts, box.Render(c))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, parts...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m model) viewBox(key string, width, limit int) string {
	switch key {
	case boxReview:
		toReview, total, notes := m.reviewList()
		cards := m.cardsByKey()
		line := func(p pr) string { return reviewPRLine(p, m.now, cardFor(p, cards)) }
		return m.viewPRBox("Revisões pendentes", strings.Join(notes, " · "), toReview, total, line, pink, width, limit)
	case boxMine:
		cards := m.cardsByKey()
		line := func(p pr) string { return prLine(p, cardFor(p, cards)) }
		return m.viewPRBox("Meus PRs abertos", mineSummary(m.prs.Mine), m.prs.Mine, m.prs.MineTotal, line, green, width, limit)
	case boxJira:
		return m.viewJiraBox(width, limit)
	case boxAgenda:
		return m.viewAgendaBox(width, limit)
	}
	return m.viewRemindersBox(width, limit)
}

// viewTabBar é a linha de abas: a caixa visível em destaque.
func (m model) viewTabBar() string {
	active := m.activeTab()
	var parts []string
	for _, key := range m.boxKeys() {
		if key == active {
			parts = append(parts, selStyle.Render(" "+boxTitles[key]+" "))
		} else {
			parts = append(parts, dimStyle.Render(boxTitles[key]))
		}
	}
	return ansi.Truncate(strings.Join(parts, dimStyle.Render(" · ")), m.width-4, "…")
}

// gridLayout escolhe quantas caixas por linha cabem em width e equilibra as
// linhas: 4 caixas que não cabem lado a lado viram 2×2, e não 3+1.
func gridLayout(width, n int) (cols, colWidth int) {
	cols = max(1, min(n, (width+boxGap)/(minBoxWidth+boxGap)))
	rows := (n + cols - 1) / cols
	cols = (n + rows - 1) / rows
	return cols, (width - boxGap*(cols-1)) / cols
}

// gap é a linha em branco entre o título da caixa e a lista, que some quando
// falta altura.
func (m model) gap(lines []string) []string {
	if m.lay.noSummary {
		return lines
	}
	return append(lines, "")
}
