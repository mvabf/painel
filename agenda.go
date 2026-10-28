package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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

	body, err := openICS(url)
	if err != nil {
		res.Err = fmt.Errorf("agenda: %w", err)
		return res
	}
	defer body.Close()

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 0, 1)

	// O gocal só esconde uma ocorrência remarcada (RECURRENCE-ID) se a versão
	// remarcada também estiver na janela, e o filtro de janela dele deixa de
	// fora eventos que começam exatamente à meia-noite. Por isso: SkipBounds
	// traz todos os eventos avulsos, a janela das recorrências ganha 1s de
	// folga e o recorte de hoje é feito por overlaps logo abaixo.
	winStart, winEnd := start.Add(-time.Second), end.Add(time.Second)
	c := gocal.NewParser(body)
	c.Start, c.End = &winStart, &winEnd
	c.SkipBounds = true
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
		// Fica se ocupa algum pedaço de hoje; eventos de duração zero, se começam hoje.
		overlaps := ev.Start.Before(end) && ev.End.After(start)
		instant := ev.End.Equal(ev.Start) && !ev.Start.Before(start) && ev.Start.Before(end)
		if !overlaps && !instant {
			continue
		}
		res.Events = append(res.Events, ev)
	}

	res.Events = tidyEvents(res.Events)
	return res
}

// tidyEvents ordena (dia inteiro primeiro, depois por horário) e remove
// eventos idênticos: o feed do Google às vezes traz duas séries iguais que
// a interface dele mostra como uma só.
func tidyEvents(events []event) []event {
	sort.SliceStable(events, func(i, j int) bool {
		a, b := events[i], events[j]
		if a.AllDay != b.AllDay {
			return a.AllDay
		}
		return a.Start.Before(b.Start)
	})

	seen := map[event]bool{}
	out := events[:0]
	for _, e := range events {
		if !seen[e] {
			seen[e] = true
			out = append(out, e)
		}
	}
	return out
}

// openICS aceita uma URL http(s) ou o caminho de um arquivo .ics local
// (útil para testar com uma exportação do calendário).
func openICS(src string) (io.ReadCloser, error) {
	if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
		return os.Open(expandHome(src))
	}

	resp, err := httpClient.Get(src)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp.Body, nil
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
