package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	gh "github.com/google/go-github/v66/github"
)

var (
	estimateRe = regexp.MustCompile(`(?i)\bEstimate:\s*(\d+)\s*days\b`)
)

func hasEstimate(body string) bool {
	return estimateRe.FindStringSubmatch(body) != nil
}

func main() {
	port := getenv("PORT", "8080")
	webhookSecret := mustGetenv("GITHUB_WEBHOOK_SECRET")
	appID := mustGetenv("GITHUB_APP_ID")
	pkPath := mustGetenv("GITHUB_PRIVATE_KEY_PATH")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		payload, err := gh.ValidatePayload(r, []byte(webhookSecret))
		if err != nil {
			log.Printf("invalid signature: %v", err)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		defer r.Body.Close()

		eventType := gh.WebHookType(r)
		event, err := gh.ParseWebHook(eventType, payload)
		if err != nil {
			log.Printf("parse webhook: %v", err)
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
				log.Printf("missing installation in event")
				break
			}

			owner := e.GetRepo().GetOwner().GetLogin()
			repo := e.GetRepo().GetName()
			number := issue.GetNumber()

			client, err := clientForInstallation(r.Context(), appID, pkPath, inst.GetID())
			if err != nil {
				log.Printf("client err: %v", err)
				break
			}

			user := issue.GetUser().GetLogin()
			msg := fmt.Sprintf("Thanks for opening this issue, @%s! To help with planning, please add an estimate in the form `Estimate: X days` to the issue description (e.g., `Estimate: 2 days`).", user)
			comment := &gh.IssueComment{Body: gh.String(msg)}

			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			_, _, err = client.Issues.CreateComment(ctx, owner, repo, number, comment)
			if err != nil {
				log.Printf("create comment: %v", err)
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           secureHeaders(mux),
		ReadHeaderTimeout: 10 * time.Second,
		TLSConfig:         &tls.Config{MinVersion: tls.VersionTLS12},
	}

	log.Printf("listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func clientForInstallation(ctx context.Context, appIDStr, pkPath string, installationID int64) (*gh.Client, error) {
	appID, err := parseInt64(appIDStr)
	if err != nil {
		return nil, err
	}

	tr := http.DefaultTransport
	itr, err := ghinstallation.NewKeyFromFile(tr, appID, installationID, pkPath)
	if err != nil {
		return nil, err
	}

	client := gh.NewClient(&http.Client{Transport: itr})

	// Support GHES if env vars present
	if base := os.Getenv("GITHUB_API_BASE_URL"); base != "" {
		client, err = gh.NewEnterpriseClient(base, os.Getenv("GITHUB_UPLOADS_BASE_URL"), &http.Client{Transport: itr})
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env: %s", key)
	}
	return v
}

func parseInt64(s string) (int64, error) {
	var v int64
	_, err := fmt.Sscan(s, &v)
	return v, err
}
