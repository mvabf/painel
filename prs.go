package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

type pr struct {
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	IsDraft    bool      `json:"isDraft"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

type prsResult struct {
	ToReview   []pr // pedidos de revisão feitos diretamente a mim
	TeamReview int  // pedidos feitos só a um time do qual faço parte
	Mine       []pr
	Err        error
	At         time.Time
}

// fetchPRs usa o gh CLI, que já está autenticado, em vez de lidar com token.
func fetchPRs() prsResult {
	res := prsResult{At: time.Now()}

	// "review-requested" inclui pedidos aos meus times, que costumam ser
	// muitos; "user-review-requested" é só o que pediram para mim.
	// As buscas rodam em paralelo porque cada uma leva alguns segundos.
	queries := []string{"user-review-requested:@me", "review-requested:@me", "author:@me"}
	results := make([][]pr, len(queries))
	errs := make([]error, len(queries))
	var wg sync.WaitGroup
	for i, q := range queries {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = searchPRs(q)
		}()
	}
	wg.Wait()

	if res.Err = errors.Join(errs...); res.Err != nil {
		return res
	}
	direct, all, mine := results[0], results[1], results[2]
	res.ToReview, res.TeamReview, res.Mine = direct, max(0, len(all)-len(direct)), mine
	return res
}

func searchPRs(query string) ([]pr, error) {
	out, err := exec.Command("gh", "search", "prs", query,
		"--state=open", "--limit=100",
		"--json=number,title,url,isDraft,updatedAt,repository",
	).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			return nil, fmt.Errorf("gh: %s", strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("gh: %w", err)
	}

	var prs []pr
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("gh: resposta inválida: %w", err)
	}
	sort.Slice(prs, func(i, j int) bool { return prs[i].UpdatedAt.After(prs[j].UpdatedAt) })
	return prs, nil
}
