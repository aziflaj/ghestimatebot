package ghestimatebot

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	gh "github.com/google/go-github/v66/github"
)

var (
	EstimateRe = regexp.MustCompile(`(?i)\bEstimate:\s*(\d+)\s*days\b`)

	commentTpl = `
		Thanks for opening this issue, @%s!

		To help with planning, please add an estimate in the form **Estimate: X days**
		to the issue description (e.g., **Estimate: 2 days**).
	`
)

type EventHandler struct {
	cfg *Config
}

func NewEventHandler(cfg *Config) *EventHandler {
	return &EventHandler{
		cfg: cfg,
	}
}

func (h *EventHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := gh.ValidatePayload(r, []byte(h.cfg.WebhookSecret))
	if err != nil {
		slog.Error("Invalid signature", "error", err)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	defer r.Body.Close()

	eventType := gh.WebHookType(r)
	event, err := gh.ParseWebHook(eventType, payload)
	if err != nil {
		slog.Error("Failed to parse webhook", "error", err)
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}

	switch e := event.(type) {
	case *gh.IssuesEvent:
		if action := e.GetAction(); action != "opened" {
			break
		}

		issue := e.GetIssue()
		if issue == nil {
			break
		}

		body := issue.GetBody()
		if hasEstimate(body) {
			break
		}

		inst := e.GetInstallation()
		if inst == nil {
			slog.Error("Missing installation in event")
			break
		}

		owner := e.GetRepo().GetOwner().GetLogin()
		repo := e.GetRepo().GetName()
		number := issue.GetNumber()

		client, err := clientForInstallation(h.cfg.AppId, h.cfg.PrivateKeyPath, inst.GetID())
		if err != nil {
			slog.Error("Failed to create client for installation", "error", err)
			break
		}

		user := issue.GetUser().GetLogin()
		msg := fmt.Sprintf(commentTpl, user)
		comment := &gh.IssueComment{Body: gh.String(msg)}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_, _, err = client.Issues.CreateComment(ctx, owner, repo, number, comment)
		if err != nil {
			slog.Error("Failed to create comment", "error", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func clientForInstallation(appIDStr, pkPath string, installationID int64) (*gh.Client, error) {
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		return nil, err
	}

	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, appID, installationID, pkPath)
	if err != nil {
		return nil, err
	}

	client := gh.NewClient(&http.Client{Transport: itr})
	return client, nil
}

func hasEstimate(body string) bool {
	return EstimateRe.FindStringSubmatch(body) != nil
}
