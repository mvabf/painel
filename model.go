package main

import (
	"os/exec"
	"runtime"
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
		case "x":
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

func openURL(url string) tea.Cmd {
	return func() tea.Msg {
		_ = openBrowser(url)
		return nil
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
