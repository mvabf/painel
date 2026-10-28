package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFetchJira(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "ana@empresa.com" || pass != "segredo" {
			// O Jira Cloud de verdade responde 200 vazio, só com este cabeçalho.
			w.Header().Set("X-Seraph-LoginReason", "AUTHENTICATED_FAILED")
			w.Write([]byte(`{"isLast": true, "issues": []}`))
			return
		}
		if r.URL.Path != "/rest/api/3/search/jql" {
			t.Errorf("caminho inesperado %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("jql"); got != defaultJiraJQL {
			t.Errorf("jql = %q", got)
		}
		w.Write([]byte(`{"isLast": false, "issues": [
			{"key": "APP-3", "fields": {"summary": "Tela de login", "status": {"name": "A fazer", "statusCategory": {"key": "new"}}}},
			{"key": "APP-1", "fields": {"summary": " Corrige frete ", "status": {"name": "Em revisão", "statusCategory": {"key": "indeterminate"}}}},
			{"key": "APP-2", "fields": {"summary": "Cache", "status": {"name": "Fazendo", "statusCategory": {"key": "indeterminate"}}}}
		]}`))
	}))
	defer srv.Close()

	cfg := config{JiraURL: srv.URL, JiraEmail: "ana@empresa.com", JiraToken: "segredo", JiraJQL: defaultJiraJQL}
	res := fetchJira(cfg)
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	var keys []string
	for _, it := range res.Issues {
		keys = append(keys, it.Key)
	}
	// Em andamento primeiro, mantendo a ordem da JQL dentro de cada grupo.
	if got := strings.Join(keys, ","); got != "APP-1,APP-2,APP-3" {
		t.Errorf("ordem = %s", got)
	}
	if res.Issues[0].Summary != "Corrige frete" || res.Issues[0].Status != "Em revisão" {
		t.Errorf("issue mal lida: %+v", res.Issues[0])
	}
	if !res.More {
		t.Error("isLast=false deveria marcar More")
	}

	cfg.JiraToken = "errado"
	if res := fetchJira(cfg); res.Err == nil || !strings.Contains(res.Err.Error(), "login recusado") {
		t.Errorf("token errado deveria virar erro, veio %v", res.Err)
	}
}

func TestFetchJiraErroDeJQL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errorMessages": ["Erro na consulta JQL: 'INT' é uma palavra reservada."]}`))
	}))
	defer srv.Close()

	res := fetchJira(config{JiraURL: srv.URL, JiraEmail: "a", JiraToken: "b", JiraJQL: "project = INT"})
	want := "jira: HTTP 400: Erro na consulta JQL: 'INT' é uma palavra reservada."
	if res.Err == nil || res.Err.Error() != want {
		t.Errorf("got %v\nwant %s", res.Err, want)
	}
}

func TestJiraConfigPorEnvFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, k := range []string{"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN", "PAINEL_ENV_FILE", "PAINEL_JIRA_URL", "PAINEL_JIRA_EMAIL", "PAINEL_JIRA_TOKEN", "PAINEL_JIRA_JQL"} {
		t.Setenv(k, "")
	}

	write := func(path, content string) {
		os.MkdirAll(filepath.Dir(path), 0o700)
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// JIRA_API_TOKEN aponta para outra variável, como em arquivos de shell.
	write(filepath.Join(home, "segredos.env"), "# tokens\nTOKEN_JIRA=\"segredo\"\nexport JIRA_BASE_URL=https://empresa.atlassian.net/\nJIRA_EMAIL=ana@empresa.com\nJIRA_API_TOKEN=${TOKEN_JIRA}\nLITERAL='$TOKEN_JIRA'\n")
	write(filepath.Join(home, ".config", "painel", "config"), "ENV_FILE=~/segredos.env\nJIRA_JQL=project = \"APP\"\n")

	cfg := loadConfig()
	if !cfg.jiraEnabled() {
		t.Fatalf("Jira deveria estar ligado: %+v", cfg)
	}
	if cfg.JiraURL != "https://empresa.atlassian.net" || cfg.JiraEmail != "ana@empresa.com" || cfg.JiraToken != "segredo" {
		t.Errorf("valores do ENV_FILE não chegaram: url=%q email=%q token ok=%v", cfg.JiraURL, cfg.JiraEmail, cfg.JiraToken == "segredo")
	}
	if cfg.JiraJQL != `project = "APP"` {
		t.Errorf("JIRA_JQL = %q", cfg.JiraJQL)
	}

	// Chave do próprio config ganha do arquivo de segredos.
	t.Setenv("PAINEL_JIRA_EMAIL", "outra@empresa.com")
	if got := loadConfig().JiraEmail; got != "outra@empresa.com" {
		t.Errorf("PAINEL_JIRA_EMAIL deveria ganhar, veio %q", got)
	}
}

func TestGridLayout(t *testing.T) {
	cases := []struct{ width, n, cols int }{
		{196, 4, 4}, // cabem as quatro
		{150, 4, 4},
		{140, 4, 2}, // cabem três: vira 2×2, não 3+1
		{110, 3, 3},
		{80, 3, 2},
		{60, 4, 1},
	}
	for _, c := range cases {
		cols, w := gridLayout(c.width, c.n)
		if cols != c.cols {
			t.Errorf("largura %d com %d caixas: %d colunas, quero %d", c.width, c.n, cols, c.cols)
		}
		if cols > 1 && w < minBoxWidth {
			t.Errorf("largura %d: caixa de %d colunas, menor que o mínimo", c.width, w)
		}
	}
}

func TestReadKVExpansao(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.env")
	t.Setenv("DO_AMBIENTE", "amb")
	os.WriteFile(path, []byte("A=1\nB=${A}-$A\nC='$A'\nD=\"$DO_AMBIENTE\"\nE=project = \"INT\"\n"), 0o600)
	got := readKV(path)
	want := map[string]string{"A": "1", "B": "1-1", "C": "$A", "D": "amb", "E": `project = "INT"`}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, quero %q", k, got[k], v)
		}
	}
}
