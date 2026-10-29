package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestParseEvents(t *testing.T) {
	if got := parseEvents(""); len(got) != 0 {
		t.Errorf("vazio deveria desligar tudo, veio %v", got)
	}
	if got := parseEvents("off"); len(got) != 0 {
		t.Errorf("off deveria desligar tudo, veio %v", got)
	}
	if got := parseEvents("on"); len(got) != len(allEvents) {
		t.Errorf("on deveria ligar tudo, veio %v", got)
	}
	got := parseEvents(" CI, meeting ")
	if !got[evCI] || !got[evMeeting] || got[evReview] {
		t.Errorf("lista mal lida: %v", got)
	}
}

func TestPRNotices(t *testing.T) {
	a := pr{URL: "a", Repo: "web-app", Number: 1, Title: "A", CI: "SUCCESS"}
	b := pr{URL: "b", Repo: "web-app", Number: 2, Title: "B", CI: "PENDING"}
	c := pr{URL: "c", Repo: "api", Number: 3, Title: "C"}

	if ns := prNotices(prsResult{}, prsResult{ToReview: []pr{c}, Mine: []pr{a}}); ns != nil {
		t.Errorf("a primeira busca não deveria avisar: %v", ns)
	}

	old := prsResult{ToReview: []pr{}, Mine: []pr{a, b}}
	failedB, approvedA := b, a
	failedB.CI = "FAILURE"
	approvedA.Review = "APPROVED"
	cur := prsResult{ToReview: []pr{c}, Mine: []pr{approvedA, failedB}}

	got := map[string]string{}
	for _, n := range prNotices(old, cur) {
		got[n.Event] = n.Body
	}
	want := map[string]string{evReview: "api#3 C", evCI: "web-app#2 B", evApproval: "web-app#1 A"}
	for ev, body := range want {
		if got[ev] != body {
			t.Errorf("%s: veio %q, esperado %q", ev, got[ev], body)
		}
	}
	if len(got) != len(want) {
		t.Errorf("avisos a mais: %v", got)
	}

	if ns := prNotices(cur, cur); len(ns) != 0 {
		t.Errorf("sem mudança não deveria avisar: %v", ns)
	}
}

func TestJiraNotices(t *testing.T) {
	one := issue{Key: "APP-1", Summary: "Um"}
	two := issue{Key: "APP-2", Summary: "Dois"}
	if ns := jiraNotices(jiraResult{}, jiraResult{Issues: []issue{one}}); ns != nil {
		t.Errorf("a primeira busca não deveria avisar: %v", ns)
	}
	ns := jiraNotices(jiraResult{Issues: []issue{}}, jiraResult{Issues: []issue{one}})
	if len(ns) != 1 || ns[0].Body != "APP-1 Um" {
		t.Errorf("faltou o card novo: %v", ns)
	}
	if ns := jiraNotices(jiraResult{Issues: []issue{one, two}}, jiraResult{Issues: []issue{two}}); len(ns) != 0 {
		t.Errorf("card que saiu não deveria avisar: %v", ns)
	}
}

func TestMeetingNotices(t *testing.T) {
	now := time.Date(2026, 10, 28, 13, 56, 0, 0, time.Local)
	at := func(h, m int) time.Time { return time.Date(2026, 10, 28, h, m, 0, 0, time.Local) }
	events := []event{
		{Title: "Feriado", AllDay: true, Start: at(0, 0), End: at(24, 0)},
		{Title: "Daily", Start: at(9, 30), End: at(9, 45)},
		{Title: "Refinamento", Start: at(14, 0), End: at(15, 0), Location: "Sala 2"},
		{Title: "1:1", Start: at(16, 0), End: at(16, 30)},
	}
	sent := map[string]bool{}

	ns := meetingNotices(events, now, 5*time.Minute, sent)
	if len(ns) != 1 || ns[0].Title != "Reunião em 4 min" || ns[0].Body != "Refinamento · Sala 2" {
		t.Fatalf("aviso errado: %v", ns)
	}
	if ns := meetingNotices(events, now.Add(time.Minute), 5*time.Minute, sent); len(ns) != 0 {
		t.Errorf("a mesma reunião não deveria avisar de novo: %v", ns)
	}
}

func TestNotifyRespeitaConfig(t *testing.T) {
	type call struct {
		title string
		sound bool
	}
	var calls []call
	orig := sendNotification
	sendNotification = func(title, body string, sound bool) error {
		calls = append(calls, call{title, sound})
		return nil
	}
	defer func() { sendNotification = orig }()

	m := model{cfg: config{Notify: parseEvents("ci,review"), NotifySound: parseEvents("ci")}}
	cmd := m.notify([]notice{
		{Event: evCI, Title: "ci"},
		{Event: evReview, Title: "review"},
		{Event: evJira, Title: "jira"},
	})
	runCmd(cmd)

	if len(calls) != 2 || calls[0] != (call{"ci", true}) || calls[1] != (call{"review", false}) {
		t.Errorf("chamadas erradas: %v", calls)
	}
}

func runCmd(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	if batch, ok := cmd().(tea.BatchMsg); ok {
		for _, c := range batch {
			runCmd(c)
		}
	}
}
