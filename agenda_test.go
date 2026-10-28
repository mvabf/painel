package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeICS grava um calendário de teste com horários relativos a hoje.
func writeICS(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cal.ics")
	// DTSTAMP é obrigatório pela RFC 5545; sem ele o gocal descarta o evento.
	body = strings.ReplaceAll(strings.TrimSpace(body), "BEGIN:VEVENT", "BEGIN:VEVENT\nDTSTAMP:20260101T000000Z")
	ics := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\n" + strings.ReplaceAll(body, "\n", "\r\n") + "\r\nEND:VCALENDAR\r\n"
	if err := os.WriteFile(path, []byte(ics), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func titles(evs []event) []string {
	var out []string
	for _, e := range evs {
		out = append(out, e.Start.Format("15:04")+" "+e.Title)
	}
	return out
}

func TestFetchAgendaICS(t *testing.T) {
	now := time.Now()
	day := func(offset, h int) time.Time {
		return time.Date(now.Year(), now.Month(), now.Day()+offset, h, 0, 0, 0, time.Local)
	}
	utc := func(tm time.Time) string { return tm.UTC().Format("20060102T150405Z") }

	path := writeICS(t, `
BEGIN:VEVENT
UID:daily
SUMMARY:Daily
DTSTART:`+utc(day(-3, 10))+`
DTEND:`+utc(day(-3, 10).Add(15*time.Minute))+`
RRULE:FREQ=DAILY
END:VEVENT
BEGIN:VEVENT
UID:daily
RECURRENCE-ID:`+utc(day(0, 10))+`
SUMMARY:Daily (remarcada)
DTSTART:`+utc(day(1, 15))+`
DTEND:`+utc(day(1, 15).Add(15*time.Minute))+`
END:VEVENT
BEGIN:VEVENT
UID:sync
SUMMARY:Sync
DTSTART:`+utc(day(-7, 14))+`
DTEND:`+utc(day(-7, 15))+`
RRULE:FREQ=DAILY
END:VEVENT
BEGIN:VEVENT
UID:sync
RECURRENCE-ID:`+utc(day(0, 14))+`
SUMMARY:Sync (mais tarde)
DTSTART:`+utc(day(0, 16))+`
DTEND:`+utc(day(0, 17))+`
END:VEVENT
BEGIN:VEVENT
UID:cancelada
SUMMARY:Cancelada
STATUS:CANCELLED
DTSTART:`+utc(day(0, 9))+`
DTEND:`+utc(day(0, 10))+`
END:VEVENT
BEGIN:VEVENT
UID:recusada
SUMMARY:Recusada
DTSTART:`+utc(day(0, 11))+`
DTEND:`+utc(day(0, 12))+`
ATTENDEE;PARTSTAT=DECLINED:mailto:eu@exemplo.com
END:VEVENT
BEGIN:VEVENT
UID:meia-noite
SUMMARY:Meia-noite
DTSTART:`+utc(day(0, 0))+`
DTEND:`+utc(day(0, 1))+`
END:VEVENT
BEGIN:VEVENT
UID:duplicada-1
SUMMARY:Almoço
DTSTART:`+utc(day(0, 12))+`
DTEND:`+utc(day(0, 13))+`
END:VEVENT
BEGIN:VEVENT
UID:duplicada-2
SUMMARY:Almoço
DTSTART:`+utc(day(0, 12))+`
DTEND:`+utc(day(0, 13))+`
END:VEVENT
BEGIN:VEVENT
UID:ontem
SUMMARY:Ontem
DTSTART:`+utc(day(-1, 9))+`
DTEND:`+utc(day(-1, 10))+`
END:VEVENT`)

	res := fetchAgenda(path, "eu@exemplo.com")
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	got := strings.Join(titles(res.Events), " | ")
	want := "00:00 Meia-noite | 12:00 Almoço | 16:00 Sync (mais tarde)"
	if got != want {
		t.Errorf("eventos de hoje:\n got: %s\nwant: %s", got, want)
	}
}
