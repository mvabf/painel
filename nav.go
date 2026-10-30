package main

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// item é uma linha que dá para selecionar com as setas e abrir com enter.
type item struct {
	Box string
	Key string
	URL string
}

const (
	boxReview = "review"
	boxMine   = "mine"
	boxJira   = "jira"
	boxAgenda = "agenda"
)

// items lista, na ordem da tela, as linhas visíveis que têm um link.
func (m model) items() []item {
	var its []item
	limit := m.listLimit()
	review, _, _ := m.reviewList()
	for i, p := range review {
		if i == limit {
			break
		}
		its = append(its, item{boxReview, p.URL, p.URL})
	}
	for i, p := range m.prs.Mine {
		if i == limit {
			break
		}
		its = append(its, item{boxMine, p.URL, p.URL})
	}
	if m.cfg.jiraEnabled() {
		for i, it := range m.jira.Issues {
			if i == limit {
				break
			}
			its = append(its, item{boxJira, issueKey(it), m.cfg.JiraURL + "/browse/" + it.Key})
		}
	}
	for i, e := range m.agenda.Events {
		if i == agendaLimit(limit) {
			break
		}
		if e.Link != "" {
			its = append(its, item{boxAgenda, eventKey(e), e.Link})
		}
	}
	return its
}

func issueKey(it issue) string { return "jira:" + it.Key }

func eventKey(e event) string { return "event:" + e.Start.Format("15:04") + " " + e.Title }

func (m model) selected() (item, int) {
	if m.selKey == "" {
		return item{}, -1
	}
	for i, it := range m.items() {
		if it.Key == m.selKey {
			return it, i
		}
	}
	return item{}, -1
}

// move anda delta itens, dando a volta no fim e no começo.
func (m model) move(delta int) model {
	its := m.items()
	if len(its) == 0 {
		return m
	}
	_, i := m.selected()
	switch {
	case i < 0 && delta > 0:
		i = 0
	case i < 0:
		i = len(its) - 1
	default:
		i = (i + delta + len(its)) % len(its)
	}
	m.selKey = its[i].Key
	return m
}

// jump vai para o primeiro item da próxima caixa (ou da anterior).
func (m model) jump(dir int) model {
	its := m.items()
	if len(its) == 0 {
		return m
	}
	cur, i := m.selected()
	if i < 0 {
		return m.move(dir)
	}
	var boxes []string
	for _, it := range its {
		if len(boxes) == 0 || boxes[len(boxes)-1] != it.Box {
			boxes = append(boxes, it.Box)
		}
	}
	b := 0
	for j, box := range boxes {
		if box == cur.Box {
			b = j
		}
	}
	target := boxes[(b+dir+len(boxes))%len(boxes)]
	for _, it := range its {
		if it.Box == target {
			m.selKey = it.Key
			break
		}
	}
	return m
}

// mark corta a linha na largura da caixa e, se for a selecionada, pinta a
// linha inteira.
func (m model) mark(key, line string, width int) string {
	line = ansi.Truncate(line, width, "…")
	if key == "" || key != m.selKey {
		return line
	}
	plain := ansi.Strip(line)
	return selStyle.Render(plain + strings.Repeat(" ", max(0, width-ansi.StringWidth(plain))))
}
