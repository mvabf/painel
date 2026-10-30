package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// reminder é um lembrete criado com a tecla n ou com `painel lembrar`.
// Fired diz se o aviso já tocou; ele fica na tela até ser concluído com x.
type reminder struct {
	At    time.Time
	Text  string
	Fired bool
}

const reminderLayout = "2006-01-02 15:04"

func remindersPath() string { return filepath.Join(configDir(), "lembretes") }

func reminderKey(r reminder) string { return "lembrete:" + r.At.Format(reminderLayout) + " " + r.Text }

// loadReminders lê o arquivo, uma linha por lembrete:
//
//   - 2026-10-30 15:30 ligar pro fornecedor
//     ! 2026-10-30 09:00 já avisado
//
// Linhas que não seguem o formato são ignoradas.
func loadReminders(path string) []reminder {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var rs []reminder
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if len(line) < 2 || (line[0] != '-' && line[0] != '!') {
			continue
		}
		fields := strings.SplitN(strings.TrimSpace(line[1:]), " ", 3)
		if len(fields) < 3 {
			continue
		}
		at, err := time.ParseInLocation(reminderLayout, fields[0]+" "+fields[1], time.Local)
		if err != nil || strings.TrimSpace(fields[2]) == "" {
			continue
		}
		rs = append(rs, reminder{At: at, Text: strings.TrimSpace(fields[2]), Fired: line[0] == '!'})
	}
	sortReminders(rs)
	return rs
}

func saveReminders(path string, rs []reminder) error {
	var b strings.Builder
	for _, r := range rs {
		mark := "-"
		if r.Fired {
			mark = "!"
		}
		fmt.Fprintf(&b, "%s %s %s\n", mark, r.At.Format(reminderLayout), r.Text)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func sortReminders(rs []reminder) {
	sort.SliceStable(rs, func(i, j int) bool { return rs[i].At.Before(rs[j].At) })
}

var (
	clockRe = regexp.MustCompile(`^(\d{1,2})(?::(\d{2})|h(\d{2})?)$`)
	dateRe  = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})$`)
)

const reminderHelp = "ex.: 15:30 texto · +20m texto · amanhã 9h texto · 25/12 10:00 texto"

// parseReminder entende "15:30 texto", "15h texto", "+20m texto",
// "amanhã 9:00 texto" e "25/12 10:00 texto". Um horário sem dia que já
// passou fica para amanhã.
func parseReminder(input string, now time.Time) (reminder, error) {
	fields := strings.Fields(input)
	if len(fields) == 0 {
		return reminder{}, errors.New("escreva quando e o quê; " + reminderHelp)
	}

	if strings.HasPrefix(fields[0], "+") {
		d, err := time.ParseDuration(strings.Replace(fields[0][1:], "min", "m", 1))
		if err != nil || d <= 0 {
			return reminder{}, fmt.Errorf("não entendi %q; use +20m, +1h ou +1h30m", fields[0])
		}
		return withText(reminder{At: now.Add(d).Truncate(time.Minute)}, fields[1:])
	}

	day, explicitDay := now, false
	switch strings.ToLower(fields[0]) {
	case "hoje":
		fields, explicitDay = fields[1:], true
	case "amanhã", "amanha":
		day, fields, explicitDay = now.AddDate(0, 0, 1), fields[1:], true
	default:
		if m := dateRe.FindStringSubmatch(fields[0]); m != nil {
			d, _ := strconv.Atoi(m[1])
			mo, _ := strconv.Atoi(m[2])
			day = time.Date(now.Year(), time.Month(mo), d, 0, 0, 0, 0, time.Local)
			if day.Day() != d || int(day.Month()) != mo {
				return reminder{}, fmt.Errorf("data inválida: %s", fields[0])
			}
			if day.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)) {
				day = day.AddDate(1, 0, 0)
			}
			fields, explicitDay = fields[1:], true
		}
	}

	if len(fields) == 0 {
		return reminder{}, errors.New("falta o horário; " + reminderHelp)
	}
	m := clockRe.FindStringSubmatch(strings.ToLower(fields[0]))
	if m == nil {
		return reminder{}, fmt.Errorf("não entendi o horário %q; %s", fields[0], reminderHelp)
	}
	h, _ := strconv.Atoi(m[1])
	mins, _ := strconv.Atoi(m[2] + m[3])
	if h > 23 || mins > 59 {
		return reminder{}, fmt.Errorf("horário inválido: %s", fields[0])
	}
	at := time.Date(day.Year(), day.Month(), day.Day(), h, mins, 0, 0, time.Local)
	if !explicitDay && !at.After(now) {
		at = at.AddDate(0, 0, 1)
	}
	if !at.After(now) {
		return reminder{}, errors.New("esse horário já passou")
	}
	return withText(reminder{At: at}, fields[1:])
}

func withText(r reminder, words []string) (reminder, error) {
	r.Text = strings.Join(words, " ")
	if r.Text == "" {
		return reminder{}, errors.New("falta o texto do lembrete; " + reminderHelp)
	}
	return r, nil
}

// dueReminders marca como avisados os lembretes que chegaram na hora e os
// devolve. Os que venceram com o painel fechado avisam ao abrir.
func dueReminders(rs []reminder, now time.Time) []reminder {
	var due []reminder
	for i := range rs {
		if !rs[i].Fired && !rs[i].At.After(now) {
			rs[i].Fired = true
			due = append(due, rs[i])
		}
	}
	return due
}

// alarm é trocada nos testes para não abrir balões nem tocar som.
var alarm = desktopAlarm

// desktopAlarm mostra o lembrete e toca o som à parte, para tocar mesmo com
// o "não perturbe" ligado.
func desktopAlarm(title, body string) error {
	_ = sendNotification(title, body, false)
	switch runtime.GOOS {
	case "linux", "freebsd", "openbsd", "netbsd":
		if err := exec.Command("canberra-gtk-play", "-i", "alarm-clock-elapsed").Run(); err == nil {
			return nil
		}
		return exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/alarm-clock-elapsed.oga").Run()
	case "darwin":
		return exec.Command("afplay", "/System/Library/Sounds/Glass.aiff").Run()
	}
	return nil
}

func reminderAlarms(due []reminder) tea.Cmd {
	var cmds []tea.Cmd
	for _, r := range due {
		r := r
		cmds = append(cmds, func() tea.Msg {
			_ = alarm("Lembrete", r.Text)
			return nil
		})
	}
	return tea.Batch(cmds...)
}

// reminderWhen é a hora do lembrete, com o dia quando não é hoje.
func reminderWhen(at, now time.Time) string {
	y1, m1, d1 := at.Date()
	y2, m2, d2 := now.Date()
	today := time.Date(y2, m2, d2, 0, 0, 0, 0, time.Local)
	switch day := time.Date(y1, m1, d1, 0, 0, 0, 0, time.Local); {
	case day.Equal(today):
		return at.Format("15:04")
	case day.Equal(today.AddDate(0, 0, 1)):
		return "amanhã " + at.Format("15:04")
	case day.After(today) && day.Before(today.AddDate(0, 0, 7)):
		return string([]rune(weekdays[at.Weekday()])[:3]) + " " + at.Format("15:04")
	}
	return at.Format("02/01 15:04")
}

// addReminderCLI é o `painel lembrar`: sem argumentos lista os pendentes.
func addReminderCLI(args []string) error {
	path := remindersPath()
	rs := loadReminders(path)
	now := time.Now()
	if len(args) == 0 {
		if len(rs) == 0 {
			fmt.Println("nenhum lembrete")
		}
		for _, r := range rs {
			mark := " "
			if !r.At.After(now) {
				mark = "!"
			}
			fmt.Printf("%s %-16s %s\n", mark, reminderWhen(r.At, now), r.Text)
		}
		return nil
	}
	r, err := parseReminder(strings.Join(args, " "), now)
	if err != nil {
		return err
	}
	rs = append(rs, r)
	sortReminders(rs)
	if err := saveReminders(path, rs); err != nil {
		return err
	}
	fmt.Printf("lembrete para %s (daqui a %s): %s\n", reminderWhen(r.At, now), humanDuration(r.At.Sub(now)), r.Text)
	return nil
}

func modTime(path string) time.Time {
	if st, err := os.Stat(path); err == nil {
		return st.ModTime()
	}
	return time.Time{}
}

// reloadReminders relê o arquivo se ele mudou por fora (editado à mão ou
// pelo `painel lembrar` em outro terminal).
func (m model) reloadReminders() model {
	if m.remindersFile == "" {
		return m
	}
	if mod := modTime(m.remindersFile); !mod.Equal(m.remindersMod) {
		m.reminders, m.remindersMod = loadReminders(m.remindersFile), mod
	}
	return m
}

func (m *model) persistReminders() {
	if m.remindersFile == "" {
		return
	}
	if err := saveReminders(m.remindersFile, m.reminders); err != nil {
		m.inputErr = "não deu para salvar os lembretes: " + err.Error()
		return
	}
	m.remindersMod = modTime(m.remindersFile)
}

func (m model) removeReminder(key string) model {
	rs := make([]reminder, 0, len(m.reminders))
	for _, r := range m.reminders {
		if reminderKey(r) != key {
			rs = append(rs, r)
		}
	}
	m.reminders = rs
	m.selKey = ""
	m.persistReminders()
	return m
}

// typeReminder cuida das teclas enquanto o campo de novo lembrete está aberto.
func (m model) typeReminder(msg tea.KeyMsg) model {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		m.typing, m.input, m.inputErr = false, "", ""
	case tea.KeyEnter:
		r, err := parseReminder(m.input, m.now)
		if err != nil {
			m.inputErr = err.Error()
			return m
		}
		m.reminders = append(append([]reminder{}, m.reminders...), r)
		sortReminders(m.reminders)
		m.typing, m.input, m.inputErr = false, "", ""
		m.persistReminders()
	case tea.KeyBackspace:
		if rs := []rune(m.input); len(rs) > 0 {
			m.input = string(rs[:len(rs)-1])
		}
	case tea.KeyCtrlU:
		m.input = ""
	case tea.KeySpace:
		m.input += " "
	case tea.KeyRunes:
		m.input += string(msg.Runes)
	}
	return m
}

func (m model) viewRemindersBox(width, limit int) string {
	inner := width - 2 - 2
	titleStyle := lipgloss.NewStyle().Foreground(accent).Bold(true)

	var fired int
	for _, r := range m.reminders {
		if r.Fired {
			fired++
		}
	}
	lines := []string{titleStyle.Render(fmt.Sprint(len(m.reminders))) + " Lembretes"}
	if fired > 0 && !m.lay.noSummary {
		lines = append(lines, failStyle.Render(fmt.Sprintf("%d na hora · x conclui", fired)))
	}
	lines = m.gap(lines)
	if len(m.reminders) == 0 {
		lines = append(lines, dimStyle.Render(ansi.Wrap("nenhum lembrete · aperte n para criar", inner, "")))
	}

	for i, r := range m.reminders {
		if i == limit {
			lines = append(lines, dimStyle.Render(fmt.Sprintf("+%d mais", len(m.reminders)-limit)))
			break
		}
		when := reminderWhen(r.At, m.now)
		var line string
		switch {
		case r.Fired:
			line = failStyle.Render("! "+when) + " " + r.Text
		case r.At.Sub(m.now) <= time.Hour:
			line = waitStyle.Render(when+" em "+humanDuration(r.At.Sub(m.now))) + " " + r.Text
		default:
			line = dimStyle.Render(when) + " " + r.Text
		}
		lines = append(lines, m.mark(reminderKey(r), line, inner))
	}
	return strings.Join(lines, "\n")
}

// viewReminderInput é o rodapé enquanto se digita um lembrete novo.
func (m model) viewReminderInput() string {
	prompt := lipgloss.NewStyle().Foreground(accent).Bold(true).Render("novo lembrete: ") + m.input + selStyle.Render(" ")
	hint := dimStyle.Render(ansi.Truncate("enter salva · esc cancela · "+reminderHelp, m.width-4, "…"))
	if m.inputErr != "" {
		hint = errStyle.Render(ansi.Truncate(m.inputErr, m.width-4, "…"))
	}
	return prompt + "\n" + hint
}
