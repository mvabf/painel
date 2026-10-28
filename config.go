package main

import (
	"bufio"
	"fmt"
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
	// ReviewMaxAge esconde pedidos de revisão mais antigos que isso (0 = todos).
	ReviewMaxAge time.Duration

	JiraURL      string // ex.: https://empresa.atlassian.net
	JiraEmail    string
	JiraToken    string
	JiraJQL      string
	JiraInterval time.Duration
}

const defaultJiraJQL = "assignee = currentUser() AND statusCategory != Done ORDER BY updated DESC"

// jiraEnabled diz se há o mínimo para consultar o Jira.
func (c config) jiraEnabled() bool {
	return c.JiraURL != "" && c.JiraEmail != "" && c.JiraToken != ""
}

// loadConfig lê ~/.config/painel/config (formato CHAVE=valor) e deixa
// variáveis de ambiente PAINEL_* sobrescreverem o arquivo.
func loadConfig() config {
	vals := map[string]string{}
	if home, err := os.UserHomeDir(); err == nil {
		vals = readKV(filepath.Join(home, ".config", "painel", "config"))
	}

	get := func(key string) string {
		if v := os.Getenv("PAINEL_" + key); v != "" {
			return v
		}
		return vals[key]
	}

	// ENV_FILE aponta para um arquivo de segredos que já existe (no mesmo
	// formato), para não copiar tokens para o config. Dele e do ambiente vêm
	// os nomes padrão do Jira (JIRA_BASE_URL, JIRA_EMAIL, JIRA_API_TOKEN).
	extra := map[string]string{}
	if path := get("ENV_FILE"); path != "" {
		extra = readKV(expandHome(path))
	}
	std := func(key string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return extra[key]
	}
	first := func(vs ...string) string {
		for _, v := range vs {
			if v != "" {
				return v
			}
		}
		return ""
	}

	cfg := config{
		Name:         get("NAME"),
		ICSURL:       get("ICS_URL"),
		Email:        get("EMAIL"),
		PRInterval:   parseDuration(get("PR_INTERVAL"), 2*time.Minute),
		CalInterval:  parseDuration(get("CAL_INTERVAL"), 5*time.Minute),
		ReviewMaxAge: parseAge(get("REVIEW_MAX_AGE")),

		JiraURL:      strings.TrimRight(first(get("JIRA_URL"), std("JIRA_BASE_URL")), "/"),
		JiraEmail:    first(get("JIRA_EMAIL"), std("JIRA_EMAIL")),
		JiraToken:    first(get("JIRA_TOKEN"), std("JIRA_API_TOKEN")),
		JiraJQL:      first(get("JIRA_JQL"), defaultJiraJQL),
		JiraInterval: parseDuration(get("JIRA_INTERVAL"), 5*time.Minute),
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

// readKV lê um arquivo CHAVE=valor. Ignora linhas vazias, comentários (#) e
// um "export " no começo, e expande $VAR/${VAR} (de linhas anteriores ou do
// ambiente), para aceitar arquivos de shell simples como
// JIRA_API_TOKEN=${TOKEN_JIRA}. Aspas simples desligam a expansão, como no shell.
func readKV(path string) map[string]string {
	vals := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		return vals
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		v = strings.TrimSpace(v)
		literal := strings.HasPrefix(v, "'")
		v = unquote(v)
		if !literal {
			v = os.Expand(v, func(name string) string {
				if val, ok := vals[name]; ok {
					return val
				}
				return os.Getenv(name)
			})
		}
		vals[strings.TrimSpace(k)] = v
	}
	return vals
}

// unquote tira aspas só quando envolvem o valor inteiro, para não estragar
// valores com aspas no meio, como JIRA_JQL=project = "INT".
func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}

func expandHome(path string) string {
	if rest, ok := strings.CutPrefix(path, "~/"); ok {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, rest)
		}
	}
	return path
}

func parseDuration(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return d
	}
	return def
}

// parseAge aceita dias ("7d") ou uma duração do Go ("36h"). Vazio ou
// inválido é 0, que significa sem limite.
func parseAge(s string) time.Duration {
	s = strings.TrimSpace(s)
	if days, ok := strings.CutSuffix(s, "d"); ok {
		var n int
		if _, err := fmt.Sscanf(days, "%d", &n); err == nil && n > 0 && fmt.Sprint(n) == days {
			return time.Duration(n) * 24 * time.Hour
		}
		return 0
	}
	return parseDuration(s, 0)
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
