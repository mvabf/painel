package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

// Login no Google Calendar para quem não tem o "endereço secreto" do iCal
// (contas Workspace em que o admin desativou esse recurso).
//
// O client OAuth é do tipo "App para computador", criado no Google Cloud
// Console e baixado como JSON em configDir()/google-client.json.

const calendarScope = "https://www.googleapis.com/auth/calendar.events.readonly"

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "painel")
}

func clientPath() string { return filepath.Join(configDir(), "google-client.json") }
func tokenPath() string  { return filepath.Join(configDir(), "google-token.json") }

func hasGoogleToken() bool {
	_, err := os.Stat(tokenPath())
	return err == nil
}

func oauthConfig() (*oauth2.Config, error) {
	raw, err := os.ReadFile(clientPath())
	if err != nil {
		return nil, fmt.Errorf("não achei %s (veja o README): %w", clientPath(), err)
	}
	var f struct {
		Installed struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		} `json:"installed"`
	}
	if err := json.Unmarshal(raw, &f); err != nil || f.Installed.ClientID == "" {
		return nil, errors.New("google-client.json inválido: o client precisa ser do tipo \"App para computador\"")
	}
	return &oauth2.Config{
		ClientID:     f.Installed.ClientID,
		ClientSecret: f.Installed.ClientSecret,
		Endpoint:     endpoints.Google,
		Scopes:       []string{calendarScope},
	}, nil
}

// googleAuth faz o fluxo OAuth com redirect para uma porta local e salva o token.
func googleAuth() error {
	cfg, err := oauthConfig()
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer ln.Close()
	cfg.RedirectURL = "http://" + ln.Addr().String()

	state := oauth2.GenerateVerifier()
	verifier := oauth2.GenerateVerifier()
	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("prompt", "consent"))

	codes := make(chan url.Values, 1)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") == "" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, "Pronto, pode voltar ao terminal.")
		codes <- r.URL.Query()
	})}
	go srv.Serve(ln)
	defer srv.Close()

	fmt.Println("Abrindo o navegador para login no Google. Se não abrir, acesse:")
	fmt.Println(authURL)
	_ = exec.Command("xdg-open", authURL).Start()

	var q url.Values
	select {
	case q = <-codes:
	case <-time.After(5 * time.Minute):
		return errors.New("tempo esgotado esperando o login")
	}
	if e := q.Get("error"); e != "" {
		return fmt.Errorf("o Google recusou o login: %s", e)
	}
	if q.Get("state") != state {
		return errors.New("resposta do login não confere (state)")
	}

	tok, err := cfg.Exchange(context.Background(), q.Get("code"), oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("trocando o código pelo token: %w", err)
	}
	if err := saveToken(tok); err != nil {
		return err
	}
	fmt.Println("Login feito. A agenda agora vem do Google Calendar.")
	return nil
}

func saveToken(tok *oauth2.Token) error {
	raw, err := json.Marshal(tok)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir(), 0o700); err != nil {
		return err
	}
	return os.WriteFile(tokenPath(), raw, 0o600)
}

// fetchGoogleAgenda busca os eventos de hoje do calendário principal.
func fetchGoogleAgenda() agendaResult {
	res := agendaResult{At: time.Now()}

	cfg, err := oauthConfig()
	if err != nil {
		res.Err = fmt.Errorf("agenda: %w", err)
		return res
	}
	raw, err := os.ReadFile(tokenPath())
	if err != nil {
		res.Err = fmt.Errorf("agenda: rode `painel auth`: %w", err)
		return res
	}
	var tok oauth2.Token
	if err := json.Unmarshal(raw, &tok); err != nil {
		res.Err = fmt.Errorf("agenda: token inválido, rode `painel auth` de novo")
		return res
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)
	client := oauth2.NewClient(ctx, cfg.TokenSource(ctx, &tok))

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	q := url.Values{
		"timeMin":      {start.Format(time.RFC3339)},
		"timeMax":      {start.AddDate(0, 0, 1).Format(time.RFC3339)},
		"singleEvents": {"true"}, // expande as recorrências
		"orderBy":      {"startTime"},
		"maxResults":   {"100"},
	}
	resp, err := client.Get("https://www.googleapis.com/calendar/v3/calendars/primary/events?" + q.Encode())
	if err != nil {
		res.Err = fmt.Errorf("agenda: %w", err)
		return res
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		res.Err = fmt.Errorf("agenda: Google Calendar respondeu HTTP %d", resp.StatusCode)
		return res
	}

	type when struct {
		Date     string    `json:"date"`
		DateTime time.Time `json:"dateTime"`
	}
	var body struct {
		Items []struct {
			Summary   string `json:"summary"`
			Status    string `json:"status"`
			Location  string `json:"location"`
			Start     when   `json:"start"`
			End       when   `json:"end"`
			Attendees []struct {
				Self           bool   `json:"self"`
				ResponseStatus string `json:"responseStatus"`
			} `json:"attendees"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		res.Err = fmt.Errorf("agenda: resposta inválida: %w", err)
		return res
	}

	for _, it := range body.Items {
		if it.Status == "cancelled" {
			continue
		}
		declined := false
		for _, a := range it.Attendees {
			if a.Self && a.ResponseStatus == "declined" {
				declined = true
			}
		}
		if declined {
			continue
		}

		ev := event{Title: strings.TrimSpace(it.Summary), Location: it.Location}
		if it.Start.Date != "" {
			ev.AllDay = true
			ev.Start, _ = time.ParseInLocation("2006-01-02", it.Start.Date, time.Local)
			ev.End, _ = time.ParseInLocation("2006-01-02", it.End.Date, time.Local)
		} else {
			ev.Start, ev.End = it.Start.DateTime.Local(), it.End.DateTime.Local()
		}
		res.Events = append(res.Events, ev)
	}

	// Mesma ordem do iCal: eventos de dia inteiro primeiro.
	sortEvents(res.Events)
	return res
}
