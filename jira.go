package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type issue struct {
	Key      string
	Summary  string
	Status   string
	Category string // "new" (a fazer), "indeterminate" (em andamento) ou "done"
}

type jiraResult struct {
	Issues []issue
	More   bool // a busca tem mais resultados do que vieram
	Err    error
	At     time.Time
}

const jiraMaxResults = 50

// fetchJira busca as issues da JQL configurada no Jira Cloud (API v3).
func fetchJira(cfg config) jiraResult {
	res := jiraResult{At: time.Now()}

	q := url.Values{
		"jql":        {cfg.JiraJQL},
		"fields":     {"summary,status"},
		"maxResults": {fmt.Sprint(jiraMaxResults)},
	}
	req, err := http.NewRequest(http.MethodGet, cfg.JiraURL+"/rest/api/3/search/jql?"+q.Encode(), nil)
	if err != nil {
		res.Err = fmt.Errorf("jira: %w", err)
		return res
	}
	req.SetBasicAuth(cfg.JiraEmail, cfg.JiraToken)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		res.Err = fmt.Errorf("jira: %w", err)
		return res
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		res.Err = fmt.Errorf("jira: HTTP %d%s", resp.StatusCode, jiraErrorDetail(resp.Body))
		return res
	}
	// Com e-mail ou token errado o Jira Cloud não devolve 401: responde 200
	// como anônimo, com a lista vazia, e só avisa neste cabeçalho.
	if strings.Contains(resp.Header.Get("X-Seraph-LoginReason"), "FAILED") {
		res.Err = fmt.Errorf("jira: login recusado, confira o e-mail e o token")
		return res
	}

	var body struct {
		IsLast bool `json:"isLast"`
		Issues []struct {
			Key    string `json:"key"`
			Fields struct {
				Summary string `json:"summary"`
				Status  struct {
					Name           string `json:"name"`
					StatusCategory struct {
						Key string `json:"key"`
					} `json:"statusCategory"`
				} `json:"status"`
			} `json:"fields"`
		} `json:"issues"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		res.Err = fmt.Errorf("jira: resposta inválida: %w", err)
		return res
	}

	for _, it := range body.Issues {
		res.Issues = append(res.Issues, issue{
			Key:      it.Key,
			Summary:  strings.TrimSpace(it.Fields.Summary),
			Status:   it.Fields.Status.Name,
			Category: it.Fields.Status.StatusCategory.Key,
		})
	}
	res.More = !body.IsLast

	// Em andamento primeiro; dentro de cada grupo, a ordem da JQL.
	sort.SliceStable(res.Issues, func(i, j int) bool {
		return categoryRank(res.Issues[i].Category) < categoryRank(res.Issues[j].Category)
	})
	return res
}

func categoryRank(category string) int {
	switch category {
	case "indeterminate":
		return 0
	case "new":
		return 1
	}
	return 2
}

// jiraErrorDetail extrai a primeira mensagem de erro do Jira (ex.: JQL inválida).
func jiraErrorDetail(r io.Reader) string {
	var body struct {
		ErrorMessages []string `json:"errorMessages"`
	}
	if err := json.NewDecoder(io.LimitReader(r, 64<<10)).Decode(&body); err == nil && len(body.ErrorMessages) > 0 {
		return ": " + body.ErrorMessages[0]
	}
	return ""
}

// jiraBrowseURL abre a mesma busca no navegador.
func jiraBrowseURL(cfg config) string {
	return cfg.JiraURL + "/issues/?jql=" + url.QueryEscape(cfg.JiraJQL)
}
