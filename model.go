package main

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	tickMsg          time.Time
	prsMsg           prsResult
	agendaMsg        agendaResult
	refreshPRsMsg    struct{ gen int }
	refreshAgendaMsg struct{ gen int }
	jiraMsg          jiraResult
	refreshJiraMsg   struct{ gen int }
)

type model struct {
	demo          bool // dados fictícios do `painel demo`: não busca nada
	cfg           config
	now           time.Time
	width, height int

	prs        prsResult
	prsLoading bool
	prsGen     int

	agenda        agendaResult
	agendaLoading bool
	agendaGen     int

	jira        jiraResult
	jiraLoading bool
	jiraGen     int

	meetingsSent map[string]bool

	selKey string // linha selecionada com as setas; vazia = nenhuma
	theme  string

	lay layout // compactação da tela; vale só durante View
	tab int    // caixa visível no modo de abas

	reminders     []reminder
	remindersFile string // vazio = não salva (demo e testes)
	remindersMod  time.Time
	typing        bool // campo de novo lembrete aberto no rodapé
	input         string
	inputErr      string
}

func newModel(cfg config) model {
	t := loadTheme(cfg.Theme)
	applyTheme(t)
	path := remindersPath()
	return model{
		reminders:     loadReminders(path),
		remindersFile: path,
		remindersMod:  modTime(path),
		theme:         t.Name,
		cfg:           cfg,
		now:           time.Now(),
		prsLoading:    true,
		agendaLoading: cfg.Agenda != "",
		jiraLoading:   cfg.jiraEnabled(),
		meetingsSent:  map[string]bool{},
	}
}

func (m model) Init() tea.Cmd {
	if m.demo {
		return tick()
	}
	return tea.Batch(tick(), m.loadPRs(), m.loadAgenda(), m.loadJira())
}

func tick() tea.Cmd {
	// Alinha no início de cada segundo para o relógio virar junto com o do sistema.
	return tea.Tick(time.Until(time.Now().Truncate(time.Second).Add(time.Second)), func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) loadPRs() tea.Cmd {
	return func() tea.Msg { return prsMsg(fetchPRs()) }
}

func (m model) loadAgenda() tea.Cmd {
	switch m.cfg.Agenda {
	case "google":
		return func() tea.Msg { return agendaMsg(fetchGoogleAgenda()) }
	case "ics":
		url, email := m.cfg.ICSURL, m.cfg.Email
		return func() tea.Msg { return agendaMsg(fetchAgenda(url, email)) }
	}
	return nil
}

func (m model) loadJira() tea.Cmd {
	if !m.cfg.jiraEnabled() {
		return nil
	}
	cfg := m.cfg
	return func() tea.Msg { return jiraMsg(fetchJira(cfg)) }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.typing {
			return m.typeReminder(msg), nil
		}
		m.inputErr = ""
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if _, i := m.selected(); i < 0 {
				return m, tea.Quit
			}
			m.selKey = ""
		case "down":
			return m.move(1), nil
		case "up":
			return m.move(-1), nil
		case "tab", "right":
			return m.nextBox(1), nil
		case "shift+tab", "left":
			return m.nextBox(-1), nil
		case "enter":
			if it, i := m.selected(); i >= 0 && it.URL != "" && !m.demo {
				return m, openURL(it.URL)
			}
		case "r":
			if m.demo {
				return m, nil
			}
			// Invalida os refreshes agendados para não buscar duas vezes.
			m.prsGen++
			m.agendaGen++
			m.jiraGen++
			m.prsLoading = true
			m.agendaLoading = m.cfg.Agenda != ""
			m.jiraLoading = m.cfg.jiraEnabled()
			return m, tea.Batch(m.loadPRs(), m.loadAgenda(), m.loadJira())
		case "n":
			m.typing, m.input, m.inputErr = true, "", ""
		case "x", "d", "delete":
			if it, i := m.selected(); i >= 0 && it.Box == boxReminders {
				m = m.removeReminder(it.Key)
			}
		case "t":
			t := loadTheme(nextTheme(m.theme))
			applyTheme(t)
			m.theme = t.Name
		case "p":
			return m, openURL("https://github.com/pulls/review-requested")
		case "j":
			if m.cfg.jiraEnabled() {
				return m, openURL(jiraBrowseURL(m.cfg))
			}
		}

	case openErrMsg:
		m.inputErr = msg.err.Error()

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tickMsg:
		m.now = time.Time(msg)
		var notes []notice
		if m.cfg.Notify[evMeeting] && m.meetingsSent != nil {
			notes = meetingNotices(m.agenda.Events, m.now, m.cfg.NotifyBefore, m.meetingsSent)
		}
		m = m.reloadReminders()
		var alarms tea.Cmd
		if due := dueReminders(m.reminders, m.now); len(due) > 0 && !m.demo {
			m.persistReminders()
			alarms = reminderAlarms(due)
		}
		return m, tea.Batch(tick(), m.notify(notes), alarms)

	case prsMsg:
		m.prsLoading = false
		res := prsResult(msg)
		var notes []notice
		if res.Err != nil {
			// Mantém os últimos números na tela e só sinaliza o erro.
			m.prs.Err, m.prs.At = res.Err, res.At
		} else {
			notes = prNotices(m.prs, res)
			m.prs = res
		}
		gen := m.prsGen
		return m, tea.Batch(m.notify(notes), tea.Tick(m.cfg.PRInterval, func(time.Time) tea.Msg { return refreshPRsMsg{gen} }))

	case agendaMsg:
		m.agendaLoading = false
		res := agendaResult(msg)
		if res.Err != nil {
			m.agenda.Err, m.agenda.At = res.Err, res.At
		} else {
			m.agenda = res
		}
		gen := m.agendaGen
		return m, tea.Tick(m.cfg.CalInterval, func(time.Time) tea.Msg { return refreshAgendaMsg{gen} })

	case jiraMsg:
		m.jiraLoading = false
		res := jiraResult(msg)
		var notes []notice
		if res.Err != nil {
			m.jira.Err, m.jira.At = res.Err, res.At
		} else {
			notes = jiraNotices(m.jira, res)
			m.jira = res
		}
		gen := m.jiraGen
		return m, tea.Batch(m.notify(notes), tea.Tick(m.cfg.JiraInterval, func(time.Time) tea.Msg { return refreshJiraMsg{gen} }))

	case refreshJiraMsg:
		if msg.gen == m.jiraGen {
			m.jiraLoading = true
			return m, m.loadJira()
		}

	case refreshPRsMsg:
		if msg.gen == m.prsGen {
			m.prsLoading = true
			return m, m.loadPRs()
		}

	case refreshAgendaMsg:
		if msg.gen == m.agendaGen {
			m.agendaLoading = true
			return m, m.loadAgenda()
		}
	}
	return m, nil
}

type openErrMsg struct{ err error }

func openURL(url string) tea.Cmd {
	return func() tea.Msg {
		if err := openBrowser(url); err != nil {
			return openErrMsg{err}
		}
		return nil
	}
}

func openBrowser(url string) error {
	for _, c := range browserCommands(url, os.Getenv("BROWSER"), runtime.GOOS, isWSL()) {
		if _, err := exec.LookPath(c[0]); err != nil {
			continue
		}
		if err := exec.Command(c[0], c[1:]...).Start(); err == nil {
			return nil
		}
	}
	return errors.New("não achei como abrir o navegador; defina a variável BROWSER")
}

// browserCommands são os jeitos de abrir um link, na ordem em que são
// tentados. No WSL o xdg-open costuma não existir ou não ter navegador, então
// o link vai para o navegador do Windows.
func browserCommands(url, browser, goos string, wsl bool) [][]string {
	var cmds [][]string
	if browser != "" {
		cmds = append(cmds, []string{browser, url})
	}
	switch {
	case goos == "darwin":
		cmds = append(cmds, []string{"open", url})
	case goos == "windows":
		cmds = append(cmds, []string{"rundll32", "url.dll,FileProtocolHandler", url})
	case wsl:
		cmds = append(cmds,
			[]string{"wslview", url},
			[]string{"rundll32.exe", "url.dll,FileProtocolHandler", url},
			[]string{"/mnt/c/Windows/System32/rundll32.exe", "url.dll,FileProtocolHandler", url},
			[]string{"xdg-open", url})
	default:
		cmds = append(cmds, []string{"xdg-open", url})
	}
	return cmds
}

func isWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}
	b, err := os.ReadFile("/proc/sys/kernel/osrelease")
	return err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft")
}
