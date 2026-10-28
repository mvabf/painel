package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type pr struct {
	Number    int
	Title     string
	URL       string
	IsDraft   bool
	UpdatedAt time.Time
	Repo      string

	CI       string // "SUCCESS", "FAILURE", "ERROR", "PENDING", "EXPECTED" ou "" sem checks
	Review   string // "APPROVED", "CHANGES_REQUESTED", "REVIEW_REQUIRED" ou ""
	Conflict bool
}

type prsResult struct {
	ToReview      []pr // pedidos de revisão feitos diretamente a mim
	ToReviewTotal int  // total real; a lista para em 50
	TeamReview    int  // pedidos feitos só a um time do qual faço parte
	Mine          []pr
	MineTotal     int
	Err           error
	At            time.Time
}

// Uma consulta GraphQL traz as três listas, com status de CI e de revisão.
// "review-requested" inclui pedidos aos meus times, que costumam ser muitos;
// "user-review-requested" é só o que pediram para mim. O issueCount é o total
// real, mesmo quando a lista vem cortada em 50.
const prsQuery = `
fragment PR on PullRequest {
  number title url isDraft updatedAt
  repository { name }
  reviewDecision mergeable
  commits(last: 1) { nodes { commit { statusCheckRollup { state } } } }
}
query {
  direct: search(query: "is:pr is:open user-review-requested:@me", type: ISSUE, first: 50) {
    issueCount
    nodes { ...PR }
  }
  all: search(query: "is:pr is:open review-requested:@me", type: ISSUE, first: 1) {
    issueCount
  }
  mine: search(query: "is:pr is:open author:@me", type: ISSUE, first: 50) {
    issueCount
    nodes { ...PR }
  }
}`

type gqlPR struct {
	Number     int       `json:"number"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	IsDraft    bool      `json:"isDraft"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	ReviewDecision string `json:"reviewDecision"`
	Mergeable      string `json:"mergeable"`
	Commits        struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *struct {
					State string `json:"state"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
}

type gqlSearch struct {
	IssueCount int     `json:"issueCount"`
	Nodes      []gqlPR `json:"nodes"`
}

// fetchPRs usa o gh CLI, que já está autenticado, em vez de lidar com token.
func fetchPRs() prsResult {
	res := prsResult{At: time.Now()}

	out, err := exec.Command("gh", "api", "graphql", "-f", "query="+prsQuery).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && len(ee.Stderr) > 0 {
			res.Err = fmt.Errorf("gh: %s", strings.TrimSpace(string(ee.Stderr)))
		} else {
			res.Err = fmt.Errorf("gh: %w", err)
		}
		return res
	}

	var body struct {
		Data struct {
			Direct gqlSearch `json:"direct"`
			All    gqlSearch `json:"all"`
			Mine   gqlSearch `json:"mine"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(out, &body); err != nil {
		res.Err = fmt.Errorf("gh: resposta inválida: %w", err)
		return res
	}
	if len(body.Errors) > 0 {
		res.Err = fmt.Errorf("gh: %s", body.Errors[0].Message)
		return res
	}

	d := body.Data
	res.ToReview = convertPRs(d.Direct.Nodes)
	res.Mine = convertPRs(d.Mine.Nodes)
	res.ToReviewTotal, res.MineTotal = d.Direct.IssueCount, d.Mine.IssueCount
	res.TeamReview = max(0, d.All.IssueCount-d.Direct.IssueCount)
	return res
}

func convertPRs(nodes []gqlPR) []pr {
	prs := make([]pr, 0, len(nodes))
	for _, n := range nodes {
		p := pr{
			Number:    n.Number,
			Title:     n.Title,
			URL:       n.URL,
			IsDraft:   n.IsDraft,
			UpdatedAt: n.UpdatedAt,
			Repo:      n.Repository.Name,
			Review:    n.ReviewDecision,
			Conflict:  n.Mergeable == "CONFLICTING",
		}
		if c := n.Commits.Nodes; len(c) > 0 && c[0].Commit.StatusCheckRollup != nil {
			p.CI = c[0].Commit.StatusCheckRollup.State
		}
		prs = append(prs, p)
	}
	sort.Slice(prs, func(i, j int) bool { return prs[i].UpdatedAt.After(prs[j].UpdatedAt) })
	return prs
}
