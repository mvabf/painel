package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type config struct {
	Name        string
	ICSURL      string
	Email       string
	Agenda      string // "google", "ics" ou "" (sem agenda)
	PRInterval  time.Duration
	CalInterval time.Duration
}

// loadConfig lê ~/.config/painel/config (formato CHAVE=valor) e deixa
// variáveis de ambiente PAINEL_* sobrescreverem o arquivo.
func loadConfig() config {
	vals := map[string]string{}

	if home, err := os.UserHomeDir(); err == nil {
		if f, err := os.Open(filepath.Join(home, ".config", "painel", "config")); err == nil {
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if k, v, ok := strings.Cut(line, "="); ok {
					vals[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
				}
			}
			f.Close()
		}
	}

	get := func(key string) string {
		if v := os.Getenv("PAINEL_" + key); v != "" {
			return v
		}
		return vals[key]
	}

	cfg := config{
		Name:        get("NAME"),
		ICSURL:      get("ICS_URL"),
		Email:       get("EMAIL"),
		PRInterval:  parseDuration(get("PR_INTERVAL"), 2*time.Minute),
		CalInterval: parseDuration(get("CAL_INTERVAL"), 5*time.Minute),
	}
	if cfg.Name == "" {
		cfg.Name = gitFirstName()
	}

	// Sem AGENDA explícito: o Google ganha se já houve `painel auth`.
	cfg.Agenda = get("AGENDA")
	if cfg.Agenda == "" {
		switch {
		case hasGoogleToken():
			cfg.Agenda = "google"
		case cfg.ICSURL != "":
			cfg.Agenda = "ics"
		}
	}
	return cfg
}

func parseDuration(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	return def
}

func gitFirstName() string {
	out, err := exec.Command("git", "config", "--global", "user.name").Output()
	if err != nil {
		return ""
	}
	if fields := strings.Fields(string(out)); len(fields) > 0 {
		return fields[0]
	}
	return ""
}
