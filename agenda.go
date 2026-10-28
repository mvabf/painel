package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/apognu/gocal"
)

type event struct {
	Title    string
	Start    time.Time
	End      time.Time
	AllDay   bool
	Location string
}

type agendaResult struct {
	Events []event
	Err    error
	At     time.Time
}

var httpClient = &http.Client{Timeout: 20 * time.Second}

// fetchAgenda baixa o calendário no formato iCal (no Google Calendar:
// Configurações > seu calendário > "Endereço secreto no formato iCal")
// e devolve os eventos de hoje, com recorrências já expandidas.
func fetchAgenda(url, email string) agendaResult {
	res := agendaResult{At: time.Now()}
	if url == "" {
		return res
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		res.Err = fmt.Errorf("agenda: %w", err)
		return res
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		res.Err = fmt.Errorf("agenda: HTTP %d", resp.StatusCode)
		return res
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 0, 1)

	c := gocal.NewParser(resp.Body)
	c.Start, c.End = &start, &end
	c.AllDayEventsTZ = time.Local
	c.Strict.Mode = gocal.StrictModeFailEvent
	c.Duplicate.Mode = gocal.DuplicateModeKeepLast
	if err := c.Parse(); err != nil {
		res.Err = fmt.Errorf("agenda: %w", err)
		return res
	}

	for _, e := range c.Events {
		if e.Start == nil || strings.EqualFold(e.Status, "CANCELLED") || declined(e, email) {
			continue
		}
		ev := event{
			Title:    strings.TrimSpace(e.Summary),
			Start:    e.Start.Local(),
			AllDay:   e.RawStart.Params["VALUE"] == "DATE",
			Location: e.Location,
		}
		if e.End != nil {
			ev.End = e.End.Local()
		} else {
			ev.End = ev.Start
		}
		res.Events = append(res.Events, ev)
	}

	sort.Slice(res.Events, func(i, j int) bool {
		a, b := res.Events[i], res.Events[j]
		if a.AllDay != b.AllDay {
			return a.AllDay
		}
		return a.Start.Before(b.Start)
	})
	return res
}

// declined esconde reuniões que eu recusei, achando meu e-mail entre os convidados.
func declined(e gocal.Event, email string) bool {
	if email == "" {
		return false
	}
	for _, a := range e.Attendees {
		if strings.EqualFold(strings.TrimPrefix(a.Value, "mailto:"), email) {
			return strings.EqualFold(a.Status, "DECLINED")
		}
	}
	return false
}
