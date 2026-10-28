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
)

type model struct {
	cfg           config
	now           time.Time
	width, height int

	prs        prsResult
	prsLoading bool
	prsGen     int

	agenda        agendaResult
	agendaLoading bool
	agendaGen     int
}

func newModel(cfg config) model {
	return model{cfg: cfg, now: time.Now(), prsLoading: true, agendaLoading: cfg.Agenda != ""}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), m.loadPRs(), m.loadAgenda())
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "r":
			// Invalida os refreshes agendados para não buscar duas vezes.
			m.prsGen++
			m.agendaGen++
			m.prsLoading = true
			m.agendaLoading = m.cfg.Agenda != ""
			return m, tea.Batch(m.loadPRs(), m.loadAgenda())
		case "p":
			return m, openURL("https://github.com/pulls/review-requested")
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height

	case tickMsg:
		m.now = time.Time(msg)
		return m, tick()

	case prsMsg:
		m.prsLoading = false
		res := prsResult(msg)
		if res.Err != nil {
			// Mantém os últimos números na tela e só sinaliza o erro.
			m.prs.Err, m.prs.At = res.Err, res.At
		} else {
			m.prs = res
		}
		gen := m.prsGen
		return m, tea.Tick(m.cfg.PRInterval, func(time.Time) tea.Msg { return refreshPRsMsg{gen} })

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
