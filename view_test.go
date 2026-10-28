package main

import (
	"strings"
	"testing"
	"time"
)

func fakeModel(width int) model {
	now := time.Date(2026, 10, 28, 14, 7, 30, 0, time.Local)
	p := func(repo string, n int, title string) pr {
		x := pr{Number: n, Title: title, UpdatedAt: now}
		x.Repository.Name = repo
		return x
	}
	ev := func(h, min, dur int, title string) event {
		s := time.Date(2026, 10, 28, h, min, 0, 0, time.Local)
		return event{Title: title, Start: s, End: s.Add(time.Duration(dur) * time.Minute)}
	}
	return model{
		cfg:    config{Name: "Marcus", ICSURL: "x"},
		now:    now,
		width:  width,
		height: 40,
		prs: prsResult{
			At:         now,
			ToReview:   []pr{p("tray-native", 4052, "Corrige retry do webhook"), p("melhor-integrador-app", 294, "Nova tela de onboarding")},
			TeamReview: 52,
			Mine:       []pr{},
		},
		agenda: agendaResult{At: now, Events: []event{
			{Title: "Feriado", AllDay: true},
			ev(9, 30, 15, "Daily"),
			ev(14, 0, 30, "Refinamento"),
			ev(16, 0, 60, "1:1"),
		}},
	}
}

func TestViewRenders(t *testing.T) {
	for _, w := range []int{160, 80} {
		out := fakeModel(w).View()
		for _, want := range []string{"Boa tarde, Marcus", "quarta-feira, 28 de outubro", "Revisões pendentes", "tray-native#4052", "nada aqui", "+52 pedidos aos seus times", "Refinamento", "agora: "} {
			if !strings.Contains(out, want) {
				t.Errorf("largura %d: faltou %q na tela:\n%s", w, want, out)
			}
		}
		if testing.Verbose() {
			t.Log("\n" + out)
		}
	}
}

func TestHumanDuration(t *testing.T) {
	cases := map[time.Duration]string{12 * time.Minute: "12 min", time.Hour: "1h", 90 * time.Minute: "1h30"}
	for d, want := range cases {
		if got := humanDuration(d); got != want {
			t.Errorf("humanDuration(%v) = %q, quero %q", d, got, want)
		}
	}
}
