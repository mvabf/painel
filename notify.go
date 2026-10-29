package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	evReview   = "review"
	evApproval = "approval"
	evCI       = "ci"
	evJira     = "jira"
	evMeeting  = "meeting"
)

var allEvents = []string{evReview, evApproval, evCI, evJira, evMeeting}

type notice struct {
	Event string
	Title string
	Body  string
}

// sendNotification é trocada nos testes para não abrir balões de verdade.
var sendNotification = desktopNotify

// parseEvents lê uma lista como "ci,meeting". "on" ou "all" liga tudo;
// vazio ou "off" desliga.
func parseEvents(s string) map[string]bool {
	s = strings.ToLower(strings.TrimSpace(s))
	set := map[string]bool{}
	switch s {
	case "", "off", "none":
		return set
	case "on", "all":
		for _, ev := range allEvents {
			set[ev] = true
		}
		return set
	}
	for _, ev := range strings.Split(s, ",") {
		if ev = strings.TrimSpace(ev); ev != "" {
			set[ev] = true
		}
	}
	return set
}

func (m model) notify(ns []notice) tea.Cmd {
	var cmds []tea.Cmd
	for _, n := range ns {
		if !m.cfg.Notify[n.Event] {
			continue
		}
		n, sound := n, m.cfg.NotifySound[n.Event]
		cmds = append(cmds, func() tea.Msg {
			_ = sendNotification(n.Title, n.Body, sound)
			return nil
		})
	}
	return tea.Batch(cmds...)
}

// prNotices compara duas buscas de PRs e diz o que mudou. A primeira busca
// (ou uma que falhou) não gera aviso, senão tudo pareceria novo ao abrir.
func prNotices(old, cur prsResult) []notice {
	if old.ToReview == nil && old.Mine == nil {
		return nil
	}
	var ns []notice

	seen := map[string]bool{}
	for _, p := range old.ToReview {
		seen[p.URL] = true
	}
	for _, p := range cur.ToReview {
		if !seen[p.URL] {
			ns = append(ns, notice{evReview, "Pedido de revisão", prRef(p)})
		}
	}

	before := map[string]pr{}
	for _, p := range old.Mine {
		before[p.URL] = p
	}
	for _, p := range cur.Mine {
		prev, existed := before[p.URL]
		if failing(p.CI) && !(existed && failing(prev.CI)) {
			ns = append(ns, notice{evCI, "CI falhou", prRef(p)})
		}
		if p.Review != prev.Review {
			switch p.Review {
			case "APPROVED":
				ns = append(ns, notice{evApproval, "PR aprovado", prRef(p)})
			case "CHANGES_REQUESTED":
				ns = append(ns, notice{evApproval, "Pediram mudanças", prRef(p)})
			}
		}
	}
	return ns
}

func failing(ci string) bool { return ci == "FAILURE" || ci == "ERROR" }

func prRef(p pr) string { return fmt.Sprintf("%s#%d %s", p.Repo, p.Number, p.Title) }

func jiraNotices(old, cur jiraResult) []notice {
	if old.Issues == nil {
		return nil
	}
	seen := map[string]bool{}
	for _, it := range old.Issues {
		seen[it.Key] = true
	}
	var ns []notice
	for _, it := range cur.Issues {
		if !seen[it.Key] {
			ns = append(ns, notice{evJira, "Card novo no Jira", it.Key + " " + it.Summary})
		}
	}
	return ns
}

// meetingNotices avisa das reuniões que começam em até `before`. sent guarda
// as já avisadas, para cada uma tocar uma vez só.
func meetingNotices(events []event, now time.Time, before time.Duration, sent map[string]bool) []notice {
	var ns []notice
	for _, e := range events {
		if e.AllDay || !e.Start.After(now) || e.Start.Sub(now) > before {
			continue
		}
		key := e.Start.Format(time.RFC3339) + " " + e.Title
		if sent[key] {
			continue
		}
		sent[key] = true
		body := e.Title
		if e.Location != "" {
			body += " · " + e.Location
		}
		ns = append(ns, notice{evMeeting, "Reunião em " + humanDuration(e.Start.Sub(now)), body})
	}
	return ns
}

// desktopNotify usa o notify-send no Linux e o osascript no macOS. No
// Windows ainda não há aviso.
func desktopNotify(title, body string, sound bool) error {
	switch runtime.GOOS {
	case "linux", "freebsd", "openbsd", "netbsd":
		args := []string{"--app-name=painel", "--icon=dialog-information"}
		if sound {
			args = append(args, "--hint=string:sound-name:message-new-instant")
		}
		return exec.Command("notify-send", append(args, title, body)...).Run()
	case "darwin":
		show := "display notification (item 2 of argv) with title (item 1 of argv)"
		if sound {
			show += ` sound name "Glass"`
		}
		return exec.Command("osascript", "-e", "on run argv", "-e", show, "-e", "end run", title, body).Run()
	}
	return nil
}
