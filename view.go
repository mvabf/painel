package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	weekdays = []string{"domingo", "segunda-feira", "terça-feira", "quarta-feira", "quinta-feira", "sexta-feira", "sábado"}
	months   = []string{"janeiro", "fevereiro", "março", "abril", "maio", "junho", "julho", "agosto", "setembro", "outubro", "novembro", "dezembro"}
)

const (
	maxListItems = 6  // itens por lista quando não se sabe a altura
	minListItems = 3  // mesmo num terminal baixo
	maxFetched   = 50 // o tamanho das buscas; não adianta pedir mais
)

// reviewList aplica o REVIEW_MAX_AGE e monta as notas da caixa de revisões.
func (m model) reviewList() (prs []pr, total int, notes []string) {
	prs, total = m.prs.ToReview, m.prs.ToReviewTotal
	if m.prs.TeamReview > 0 {
		notes = append(notes, fmt.Sprintf("+%d pedidos aos seus times", m.prs.TeamReview))
	}
	if m.cfg.ReviewMaxAge > 0 && prs != nil {
		var hidden int
		prs, hidden = recentReviews(prs, m.cfg.ReviewMaxAge, m.now)
		total = len(prs)
		if hidden > 0 {
			notes = append(notes, fmt.Sprintf("ocultos: %d com mais de %s", hidden, ageLabel(m.cfg.ReviewMaxAge)))
		}
	}
	return prs, total, notes
}

// agendaLimit: a agenda não tem linha de resumo, então cabem dois a mais.
func agendaLimit(limit int) int { return limit + 2 }

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

	width := m.width - 4
	if m.lay.smallClock || width < lipgloss.Width(clock)+5+max(len(greeting), len(date)) {
		return m.viewSmallHeader(greeting)
	}

	info := []string{greetStyle.Render(greeting), dimStyle.Render(date)}
	next := m.viewNextEvent()
	if free := width - lipgloss.Width(clock) - 5 - timelineExtra; free >= minTimeline {
		info = append(append(info, ""), m.viewDayProgress(min(maxTimeline, free))...)
		if next != "" {
			info = append(info, next)
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, clock, "     ", lipgloss.JoinVertical(lipgloss.Left, info...))
	}

	// Estreito: a linha do tempo desce para baixo do relógio.
	if next != "" {
		info = append(info, "", ansi.Truncate(next, width-lipgloss.Width(clock)-5, "…"))
	}
	top := lipgloss.JoinHorizontal(lipgloss.Top, clock, "     ", lipgloss.JoinVertical(lipgloss.Left, info...))
	return lipgloss.JoinVertical(lipgloss.Left, append([]string{top, ""}, m.viewDayProgress(timelineFor(width))...)...)
}

// viewSmallHeader é o cabeçalho de quando falta espaço: hora, saudação, data
// e a próxima reunião numa linha e a linha do tempo, sem as horas, na outra.
// Se a primeira linha não couber, saem a data e depois a saudação.
func (m model) viewSmallHeader(greeting string) string {
	now := m.now
	width := m.width - 4
	day := strings.Split(weekdays[now.Weekday()], "-")[0]
	clock := clockStyle.Bold(true).Render(now.Format("15:04:05"))
	parts := []string{greetStyle.Render(greeting), dimStyle.Render(fmt.Sprintf("%s, %02d/%02d", day, now.Day(), now.Month()))}
	if next := m.viewNextEvent(); next != "" {
		parts = append(parts, next)
	}
	line := clock + "  " + strings.Join(parts, dimStyle.Render(" · "))
	for _, drop := range []int{1, 0} {
		if lipgloss.Width(line) <= width || drop >= len(parts) {
			break
		}
		parts = append(parts[:drop], parts[drop+1:]...)
		line = clock + "  " + strings.Join(parts, dimStyle.Render(" · "))
	}
	progress := m.viewDayProgress(timelineFor(width))
	return lipgloss.JoinVertical(lipgloss.Left, ansi.Truncate(line, width, "…"), progress[len(progress)-1])
}

const (
	minTimeline   = 20
	maxTimeline   = 108
	timelineExtra = len("expediente ") + len(" 100%")
)

// timelineFor é a largura da linha do tempo quando ela tem a linha toda.
func timelineFor(width int) int {
	return max(10, min(maxTimeline, width-timelineExtra))
}

// viewDayProgress mostra o expediente (9h–18h) como uma linha do tempo: o
// que já passou em destaque, as reuniões em amarelo e um ponto no agora.
func (m model) viewDayProgress(width int) []string {
	now := m.now
	if wd := now.Weekday(); wd == time.Saturday || wd == time.Sunday {
		return []string{dimStyle.Render("fim de semana 🎉")}
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), workStart, 0, 0, 0, time.Local)
	end := time.Date(now.Year(), now.Month(), now.Day(), workEnd, 0, 0, 0, time.Local)
	frac := max(0, min(1, float64(now.Sub(start))/float64(end.Sub(start))))

	const label = "expediente "
	var bar strings.Builder
	for _, c := range timelineCells(now, m.agenda.Events, start, end, width) {
		switch c {
		case cellNow:
			bar.WriteString(lipgloss.NewStyle().Foreground(pink).Render("●"))
		case cellMeetingPast:
			bar.WriteString(lipgloss.NewStyle().Foreground(yellow).Faint(true).Render("━"))
		case cellMeeting:
			bar.WriteString(lipgloss.NewStyle().Foreground(yellow).Render("━"))
		case cellPast:
			bar.WriteString(lipgloss.NewStyle().Foreground(accent).Render("━"))
		default:
			bar.WriteString(dimStyle.Render("━"))
		}
	}
	pct := dimStyle.Render(fmt.Sprintf("%3.0f%%", frac*100))
	return []string{
		strings.Repeat(" ", len(label)) + dimStyle.Render(hourLabels(width)),
		label + bar.String() + " " + pct,
	}
}

const (
	workStart = 9
	workEnd   = 18
)

type cell int

const (
	cellFuture cell = iota
	cellPast
	cellMeeting
	cellMeetingPast
	cellNow
)

// timelineCells divide o expediente em width pedaços iguais e diz o que há
// no meio de cada um.
func timelineCells(now time.Time, events []event, start, end time.Time, width int) []cell {
	cells := make([]cell, width)
	span := end.Sub(start)
	nowCell := -1
	if !now.Before(start) && now.Before(end) {
		nowCell = int(float64(now.Sub(start)) / float64(span) * float64(width))
	}
	for i := range cells {
		t := start.Add(time.Duration((float64(i) + 0.5) / float64(width) * float64(span)))
		past := t.Before(now)
		meeting := false
		for _, e := range events {
			if !e.AllDay && !t.Before(e.Start) && t.Before(e.End) {
				meeting = true
				break
			}
		}
		switch {
		case i == nowCell:
			cells[i] = cellNow
		case meeting && past:
			cells[i] = cellMeetingPast
		case meeting:
			cells[i] = cellMeeting
		case past:
			cells[i] = cellPast
		}
	}
	return cells
}

// hourLabels escreve as horas cheias sobre a linha do tempo. O 18h fica
// sempre, alinhado ao fim; as outras horas são puladas se encostariam na
// vizinha.
func hourLabels(width int) string {
	line := []rune(strings.Repeat(" ", width))
	last := fmt.Sprintf("%dh", workEnd)
	lastPos := width - len(last)
	copy(line[lastPos:], []rune(last))

	next := 0
	for h := workStart; h < workEnd; h++ {
		pos := int(float64(h-workStart) / float64(workEnd-workStart) * float64(width))
		text := fmt.Sprintf("%dh", h)
		if pos < next || pos+len(text) >= lastPos {
			continue
		}
		copy(line[pos:], []rune(text))
		next = pos + len(text) + 1
	}
	return string(line)
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
func (m model) viewPRBox(title, note string, prs []pr, total int, line func(pr) string, color lipgloss.AdaptiveColor, width, limit int) string {
	inner := width - 2 - 2 // borda + padding
	titleStyle := lipgloss.NewStyle().Foreground(color).Bold(true)

	var lines []string
	count := "…"
	if m.prs.ToReview != nil || m.prs.Mine != nil { // já houve uma busca com sucesso
		count = fmt.Sprint(max(total, len(prs)))
	}
	lines = append(lines, titleStyle.Render(count)+" "+title+m.spinner(m.prsLoading))
	if note != "" && !m.lay.noSummary {
		lines = append(lines, wrapParts(note, inner)...)
	}
	lines = m.gap(lines)

	if m.prs.Err != nil {
		lines = append(lines, errStyle.Render(ansi.Truncate(m.prs.Err.Error(), inner, "…")))
	}
	if len(prs) == 0 && m.prs.Err == nil && !m.prs.At.IsZero() {
		lines = append(lines, dimStyle.Render("nada aqui ✨"))
	}

	for i, p := range prs {
		if i == limit {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("+%d mais", max(total, len(prs))-limit)))
			break
		}
		lines = append(lines, m.mark(p.URL, line(p), inner))
	}

	return strings.Join(lines, "\n")
}

func (m model) cardsByKey() map[string]issue {
	cards := map[string]issue{}
	for _, it := range m.jira.Issues {
		cards[it.Key] = it
	}
	return cards
}

// cardFor é o primeiro card da busca do Jira citado pelo PR.
func cardFor(p pr, cards map[string]issue) *issue {
	for _, k := range p.Keys {
		if it, ok := cards[k]; ok {
			return &it
		}
	}
	return nil
}

// prFor é o PR meu mais recente que cita o card.
func prFor(it issue, prs []pr) *pr {
	for _, p := range prs {
		for _, k := range p.Keys {
			if k == it.Key {
				return &p
			}
		}
	}
	return nil
}

// prLine: ícone do CI, repo#número, marcadores de revisão/conflito, status do
// card no Jira e título.
func prLine(p pr, card *issue) string {
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

	if card != nil {
		tags = append(tags, cardStyle.Render(card.Status))
	}

	parts := append([]string{icon, dimStyle.Render(ref)}, tags...)
	return strings.Join(append(parts, p.Title), " ")
}

// reviewPRLine mostra há quanto tempo o pedido de revisão espera: cinza até
// 1 dia, amarelo até 3 e vermelho depois disso.
func reviewPRLine(p pr, now time.Time, card *issue) string {
	ref := fmt.Sprintf("%s#%d", p.Repo, p.Number)
	if p.IsDraft {
		return dimStyle.Render(ref + " [draft] " + p.Title)
	}

	waited := now.Sub(p.RequestedAt)
	style := dimStyle
	switch {
	case waited >= 3*24*time.Hour:
		style = failStyle
	case waited >= 24*time.Hour:
		style = waitStyle
	}
	line := style.Render(fmt.Sprintf("%3s", shortAge(waited))) + " " + dimStyle.Render(ref) + " "
	if card != nil {
		line += cardStyle.Render(card.Status) + " "
	}
	return line + p.Title
}

// recentReviews separa os pedidos feitos há no máximo maxAge.
func recentReviews(prs []pr, maxAge time.Duration, now time.Time) (recent []pr, hidden int) {
	recent = []pr{}
	for _, p := range prs {
		if now.Sub(p.RequestedAt) <= maxAge {
			recent = append(recent, p)
		} else {
			hidden++
		}
	}
	return recent, hidden
}

// ageLabel: "7 dias", "1 dia", "36h".
func ageLabel(d time.Duration) string {
	const day = 24 * time.Hour
	switch {
	case d == day:
		return "1 dia"
	case d%day == 0:
		return fmt.Sprintf("%d dias", d/day)
	}
	// time.Duration escreve "36h0m0s"; tira só os zeros do fim.
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = strings.TrimSuffix(s, "0s")
	}
	if strings.HasSuffix(s, "h0m") {
		s = strings.TrimSuffix(s, "0m")
	}
	return s
}

// shortAge: "45m", "5h", "3d".
func shortAge(d time.Duration) string {
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm", max(0, int(d.Minutes())))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
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
	add(failing, "CI falhando", "CI falhando")
	add(changes, "com mudanças", "com mudanças")
	add(conflicts, "com conflito", "com conflito")
	add(approved, "aprovado", "aprovados")
	return strings.Join(parts, " · ")
}

// wrapParts quebra um resumo "a · b · c" em linhas de até width colunas,
// sempre entre os itens, e devolve as linhas já em cinza.
func wrapParts(note string, width int) []string {
	var lines []string
	cur := ""
	for _, part := range strings.Split(note, " · ") {
		switch {
		case cur == "":
			cur = part
		case ansi.StringWidth(cur+" · "+part) <= width:
			cur += " · " + part
		default:
			lines = append(lines, dimStyle.Render(ansi.Truncate(cur, width, "…")))
			cur = part
		}
	}
	if cur != "" {
		lines = append(lines, dimStyle.Render(ansi.Truncate(cur, width, "…")))
	}
	return lines
}

func (m model) viewAgendaBox(width, limit int) string {
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
	lines := m.gap([]string{titleStyle.Render(count) + " Agenda de hoje" + m.spinner(m.agendaLoading)})
	if m.agenda.Err != nil {
		lines = append(lines, errStyle.Render(ansi.Truncate(m.agenda.Err.Error(), inner, "…")))
	}
	if len(m.agenda.Events) == 0 && m.agenda.Err == nil && !m.agenda.At.IsZero() {
		lines = append(lines, dimStyle.Render("agenda livre 🌴"))
	}

	shown := 0
	for _, e := range m.agenda.Events {
		if shown == agendaLimit(limit) {
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
		if e.Link != "" {
			text += " ↗"
		}
		key := ""
		if e.Link != "" {
			key = eventKey(e)
		}
		lines = append(lines, m.mark(key, style.Render(ansi.Truncate(text, inner, "…")), inner))
	}

	return strings.Join(lines, "\n")
}

func (m model) viewJiraBox(width, limit int) string {
	inner := width - 2 - 2
	titleStyle := lipgloss.NewStyle().Foreground(blue).Bold(true)

	count := "…"
	if !m.jira.At.IsZero() {
		count = fmt.Sprint(len(m.jira.Issues))
		if m.jira.More {
			count += "+"
		}
	}
	lines := []string{titleStyle.Render(count) + " Jira" + m.spinner(m.jiraLoading)}

	var doing, todo int
	for _, it := range m.jira.Issues {
		switch it.Category {
		case "indeterminate":
			doing++
		case "new":
			todo++
		}
	}
	var summary []string
	if doing > 0 {
		summary = append(summary, fmt.Sprintf("%d em andamento", doing))
	}
	if todo > 0 {
		summary = append(summary, fmt.Sprintf("%d a fazer", todo))
	}
	if len(summary) > 0 && !m.lay.noSummary {
		lines = append(lines, wrapParts(strings.Join(summary, " · "), inner)...)
	}
	lines = m.gap(lines)

	if m.jira.Err != nil {
		lines = append(lines, errStyle.Render(ansi.Truncate(m.jira.Err.Error(), inner, "…")))
	}
	if len(m.jira.Issues) == 0 && m.jira.Err == nil && !m.jira.At.IsZero() {
		lines = append(lines, dimStyle.Render("nada atribuído ✨"))
	}

	for i, it := range m.jira.Issues {
		if i == limit {
			more := fmt.Sprintf("+%d mais", len(m.jira.Issues)-limit)
			if m.jira.More {
				more = "e mais"
			}
			lines = append(lines, dimStyle.Render(more))
			break
		}
		lines = append(lines, m.mark(issueKey(it), issueLine(it, prFor(it, m.prs.Mine)), inner))
	}
	return strings.Join(lines, "\n")
}

// issueLine: chave colorida pelo andamento, status (se em andamento), como
// está o meu PR do card (se houver) e título.
func issueLine(it issue, mine *pr) string {
	var parts []string
	switch it.Category {
	case "indeterminate":
		parts = append(parts, waitStyle.Render(it.Key), dimStyle.Render(it.Status))
	case "done":
		parts = append(parts, okStyle.Render(it.Key))
	default:
		parts = append(parts, dimStyle.Render(it.Key))
	}
	if mine != nil {
		parts = append(parts, prTag(*mine))
	}
	return strings.Join(append(parts, it.Summary), " ")
}

// prTag resume o PR numa palavra, do mais urgente para o menos.
func prTag(p pr) string {
	switch {
	case p.IsDraft:
		return dimStyle.Render("PR draft")
	case failing(p.CI):
		return failStyle.Render("PR CI falhando")
	case p.Review == "CHANGES_REQUESTED":
		return failStyle.Render("PR mudanças")
	case p.Review == "APPROVED":
		return okStyle.Render("PR aprovado")
	}
	return cardStyle.Render("PR aberto")
}

func (m model) viewFooter() string {
	if m.typing {
		return m.viewReminderInput()
	}
	parts := []string{}
	if m.inputErr != "" {
		parts = append(parts, errStyle.Render(m.inputErr))
	}
	if m.demo {
		parts = append(parts, waitStyle.Render("demo com dados fictícios"))
	}
	// Com pouco espaço, os horários das buscas saem antes dos atalhos.
	if !m.lay.smallClock {
		if !m.prs.At.IsZero() {
			parts = append(parts, "PRs às "+m.prs.At.Format("15:04"))
		}
		if !m.agenda.At.IsZero() {
			parts = append(parts, "agenda às "+m.agenda.At.Format("15:04"))
		}
		if !m.jira.At.IsZero() {
			parts = append(parts, "jira às "+m.jira.At.Format("15:04"))
		}
	}
	switch {
	case strings.HasPrefix(m.selKey, "lembrete:"):
		parts = append(parts, "x conclui", "esc desmarca")
	case m.selKey != "":
		parts = append(parts, "enter abre", "esc desmarca")
	default:
		parts = append(parts, "↑↓ navega")
	}
	if m.lay.tabs {
		parts = append(parts, "tab troca de caixa")
	}
	parts = append(parts, "n lembrete", "r atualiza", "t tema: "+m.themeLabel(), "p abre revisões")
	if m.cfg.jiraEnabled() {
		parts = append(parts, "j abre o Jira")
	}
	parts = append(parts, "q sai")
	return strings.Join(wrapParts(strings.Join(parts, " · "), m.width-4), "\n")
}

func (m model) themeLabel() string {
	if m.theme == "" {
		return defaultTheme
	}
	return m.theme
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
