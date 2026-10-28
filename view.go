package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	accent  = lipgloss.AdaptiveColor{Light: "#7D56F4", Dark: "#B49CFF"}
	pink    = lipgloss.AdaptiveColor{Light: "#D6336C", Dark: "#FF7AB2"}
	green   = lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#69DB7C"}
	yellow  = lipgloss.AdaptiveColor{Light: "#E67700", Dark: "#FFD43B"}
	subtle  = lipgloss.AdaptiveColor{Light: "#868E96", Dark: "#6C7086"}
	errColr = lipgloss.AdaptiveColor{Light: "#C92A2A", Dark: "#FF6B6B"}

	clockStyle = lipgloss.NewStyle().Foreground(accent)
	greetStyle = lipgloss.NewStyle().Bold(true)
	dimStyle   = lipgloss.NewStyle().Foreground(subtle)
	errStyle   = lipgloss.NewStyle().Foreground(errColr)
	boxStyle   = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle).
			Padding(0, 1)
)

var (
	weekdays = []string{"domingo", "segunda-feira", "terça-feira", "quarta-feira", "quinta-feira", "sexta-feira", "sábado"}
	months   = []string{"janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
)

const maxListItems = 6

func (m model) View() string {
	if m.width == 0 {
		return ""
	}

	header := m.viewHeader()

	// Três colunas lado a lado quando cabe; senão empilha.
	cols := 3
	colWidth := (m.width - 4 - 2*(cols-1)) / cols
	if colWidth < 34 {
		cols = 1
		colWidth = m.width - 4
	}

	var teamNote string
	if m.prs.TeamReview > 0 {
		teamNote = fmt.Sprintf("+%d pedidos aos seus times", m.prs.TeamReview)
	}
	review := m.viewPRBox("Revisões pendentes", teamNote, m.prs.ToReview, m.prs.ToReviewTotal, pink, colWidth)
	mine := m.viewPRBox("Meus PRs abertos", mineSummary(m.prs.Mine), m.prs.Mine, m.prs.MineTotal, green, colWidth)
	agenda := m.viewAgendaBox(colWidth)

	var body string
	box := boxStyle.Width(colWidth - 2)
	if cols == 3 {
		// Iguala a altura das caixas para a linha ficar alinhada.
		box = box.Height(max(lipgloss.Height(review), lipgloss.Height(mine), lipgloss.Height(agenda)))
		body = lipgloss.JoinHorizontal(lipgloss.Top, box.Render(review), "  ", box.Render(mine), "  ", box.Render(agenda))
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, box.Render(review), box.Render(mine), box.Render(agenda))
	}

	out := lipgloss.JoinVertical(lipgloss.Left, header, "", body, "", m.viewFooter())
	return lipgloss.NewStyle().Padding(1, 2).Render(out)
}

func (m model) viewHeader() string {
	now := m.now
	clock := clockStyle.Render(bigText(now.Format("15:04")))
	secs := dimStyle.Render(now.Format(":05"))
	clock = lipgloss.JoinHorizontal(lipgloss.Bottom, clock, " ", secs)

	greeting := "Boa noite"
	switch h := now.Hour(); {
	case h >= 5 && h < 12:
		greeting = "Bom dia"
	case h >= 12 && h < 18:
		greeting = "Boa tarde"
	}
	if m.cfg.Name != "" {
		greeting += ", " + m.cfg.Name
	}

	date := fmt.Sprintf("%s, %d de %s", weekdays[now.Weekday()], now.Day(), months[now.Month()-1])

	info := []string{
		greetStyle.Render(greeting),
		dimStyle.Render(date),
		"",
		m.viewDayProgress(),
	}
	if next := m.viewNextEvent(); next != "" {
		info = append(info, next)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, clock, "     ", lipgloss.JoinVertical(lipgloss.Left, info...))
}

// viewDayProgress mostra quanto do expediente (9h–18h) já passou.
func (m model) viewDayProgress() string {
	now := m.now
	if wd := now.Weekday(); wd == time.Saturday || wd == time.Sunday {
		return dimStyle.Render("fim de semana 🎉")
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.Local)
	end := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, time.Local)
	frac := float64(now.Sub(start)) / float64(end.Sub(start))
	frac = max(0, min(1, frac))

	const width = 20
	filled := int(frac * width)
	bar := lipgloss.NewStyle().Foreground(accent).Render(strings.Repeat("━", filled)) +
		dimStyle.Render(strings.Repeat("━", width-filled))
	return fmt.Sprintf("expediente %s %s", bar, dimStyle.Render(fmt.Sprintf("%3.0f%%", frac*100)))
}

func (m model) viewNextEvent() string {
	for _, e := range m.agenda.Events {
		if e.AllDay {
			continue
		}
		if !m.now.Before(e.Start) && m.now.Before(e.End) {
			return lipgloss.NewStyle().Foreground(yellow).Render("agora: ") + e.Title
		}
		if e.Start.After(m.now) {
			return dimStyle.Render("próxima: ") + e.Title + dimStyle.Render(" em "+humanDuration(e.Start.Sub(m.now)))
		}
	}
	return ""
}

// viewPRBox e viewAgendaBox devolvem só o conteúdo; a borda é aplicada em View.
func (m model) viewPRBox(title, note string, prs []pr, total int, color lipgloss.AdaptiveColor, width int) string {
	inner := width - 2 - 2 // borda + padding
	titleStyle := lipgloss.NewStyle().Foreground(color).Bold(true)

	var lines []string
	count := "…"
	if m.prs.ToReview != nil || m.prs.Mine != nil { // já houve uma busca com sucesso
		count = fmt.Sprint(max(total, len(prs)))
	}
	lines = append(lines, titleStyle.Render(count)+" "+title+m.spinner(m.prsLoading))
	if note != "" {
		lines = append(lines, dimStyle.Render(note))
	}
	lines = append(lines, "")

	if m.prs.Err != nil {
		lines = append(lines, errStyle.Render(ansi.Truncate(m.prs.Err.Error(), inner, "…")))
	}
	if len(prs) == 0 && m.prs.Err == nil && !m.prs.At.IsZero() {
		lines = append(lines, dimStyle.Render("nada aqui ✨"))
	}

	for i, p := range prs {
		if i == maxListItems {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("+%d mais", max(total, len(prs))-maxListItems)))
			break
		}
		lines = append(lines, ansi.Truncate(prLine(p), inner, "…"))
	}

	return strings.Join(lines, "\n")
}

var (
	okStyle   = lipgloss.NewStyle().Foreground(green)
	failStyle = lipgloss.NewStyle().Foreground(errColr)
	waitStyle = lipgloss.NewStyle().Foreground(yellow)
)

// prLine: ícone do CI, repo#número, marcadores de revisão/conflito e título.
func prLine(p pr) string {
	ref := fmt.Sprintf("%s#%d", p.Repo, p.Number)
	if p.IsDraft {
		return dimStyle.Render("· " + ref + " [draft] " + p.Title)
	}

	icon := dimStyle.Render("·")
	switch p.CI {
	case "SUCCESS":
		icon = okStyle.Render("✓")
	case "FAILURE", "ERROR":
		icon = failStyle.Render("✗")
	case "PENDING", "EXPECTED":
		icon = waitStyle.Render("●")
	}

	var tags []string
	if p.Conflict {
		tags = append(tags, failStyle.Render("conflito"))
	}
	switch p.Review {
	case "CHANGES_REQUESTED":
		tags = append(tags, failStyle.Render("mudanças"))
	case "APPROVED":
		tags = append(tags, okStyle.Render("aprovado"))
	}

	parts := append([]string{icon, dimStyle.Render(ref)}, tags...)
	return strings.Join(append(parts, p.Title), " ")
}

// mineSummary resume o que pede ação nos meus PRs, mesmo os que não cabem na lista.
func mineSummary(prs []pr) string {
	var failing, changes, conflicts, approved int
	for _, p := range prs {
		if p.IsDraft {
			continue
		}
		if p.CI == "FAILURE" || p.CI == "ERROR" {
			failing++
		}
		if p.Conflict {
			conflicts++
		}
		switch p.Review {
		case "CHANGES_REQUESTED":
			changes++
		case "APPROVED":
			approved++
		}
	}

	var parts []string
	add := func(n int, one, many string) {
		switch {
		case n == 1:
			parts = append(parts, "1 "+one)
		case n > 1:
			parts = append(parts, fmt.Sprintf("%d %s", n, many))
		}
	}
	add(failing, "com CI falhando", "com CI falhando")
	add(changes, "com mudanças pedidas", "com mudanças pedidas")
	add(conflicts, "com conflito", "com conflito")
	add(approved, "aprovado", "aprovados")
	return strings.Join(parts, " · ")
}

func (m model) viewAgendaBox(width int) string {
	inner := width - 2 - 2
	titleStyle := lipgloss.NewStyle().Foreground(yellow).Bold(true)

	if m.cfg.Agenda == "" {
		return strings.Join([]string{
			titleStyle.Render("Agenda de hoje"), "",
			dimStyle.Render(ansi.Wrap("Rode `painel auth` para usar o Google Calendar ou configure ICS_URL (veja o README).", inner, "")),
		}, "\n")
	}

	count := "…"
	if !m.agenda.At.IsZero() { // já houve uma busca
		count = fmt.Sprint(len(m.agenda.Events))
	}
	lines := []string{titleStyle.Render(count) + " Agenda de hoje" + m.spinner(m.agendaLoading), ""}
	if m.agenda.Err != nil {
		lines = append(lines, errStyle.Render(ansi.Truncate(m.agenda.Err.Error(), inner, "…")))
	}
	if len(m.agenda.Events) == 0 && m.agenda.Err == nil && !m.agenda.At.IsZero() {
		lines = append(lines, dimStyle.Render("agenda livre 🌴"))
	}

	shown := 0
	for _, e := range m.agenda.Events {
		if shown == maxListItems+2 {
			lines = append(lines, dimStyle.Render("…"))
			break
		}
		shown++

		var when string
		style := lipgloss.NewStyle()
		switch {
		case e.AllDay:
			when = "dia todo"
		default:
			when = e.Start.Format("15:04")
		}
		switch {
		case !e.AllDay && !m.now.Before(e.End):
			style = dimStyle.Strikethrough(true)
		case !e.AllDay && !m.now.Before(e.Start):
			style = style.Foreground(yellow).Bold(true)
		}
		text := fmt.Sprintf("%-8s %s", when, e.Title)
		lines = append(lines, style.Render(ansi.Truncate(text, inner, "…")))
	}

	return strings.Join(lines, "\n")
}

func (m model) viewFooter() string {
	parts := []string{}
	if !m.prs.At.IsZero() {
		parts = append(parts, "PRs às "+m.prs.At.Format("15:04"))
	}
	if !m.agenda.At.IsZero() {
		parts = append(parts, "agenda às "+m.agenda.At.Format("15:04"))
	}
	parts = append(parts, "r atualiza", "p abre revisões", "q sai")
	return dimStyle.Render(strings.Join(parts, " · "))
}

func (m model) spinner(loading bool) string {
	if !loading {
		return ""
	}
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return " " + dimStyle.Render(frames[m.now.Second()%len(frames)])
}

func humanDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	if d < time.Hour {
		return fmt.Sprintf("%d min", int(d.Minutes()))
	}
	h, min := int(d.Hours()), int(d.Minutes())%60
	if min == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh%02d", h, min)
}
