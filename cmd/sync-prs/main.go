// sync-prs fetches open PRs from bitcoin/bitcoin on GitHub and writes them to the local DB.
// Run with DATABASE_* and GITHUB_ACCESS_TOKEN set (e.g. from scripts/start-local.sh).
package main

import (
	"context"
	"log"
	"os"

	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type Config struct {
	config.DatabaseConfig
	Github struct {
		AccessToken string `env:"ACCESS_TOKEN"`
	} `env-prefix:"GITHUB_"`
}

func main() {
	var cfg Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.Github.AccessToken == "" {
		log.Fatal("GITHUB_ACCESS_TOKEN is required. Set it to a GitHub personal access token (no scope needed for public repo).")
	}

	os.Setenv("AUTO_MIGRATE", "true")
	if err := db.Connect(cfg.DatabaseConfig); err != nil {
		log.Fatalf("db connect: %v", err)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.Github.AccessToken})
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	log.Println("Syncing open PRs from bitcoin/bitcoin...")
	page := 0
	maxPages := 5
	synced := 0
	for page < maxPages {
		prs, _, err := client.PullRequests.List(ctx, "bitcoin", "bitcoin", &github.PullRequestListOptions{
			State:     "open",
			Sort:      "updated",
			Direction: "desc",
			ListOptions: github.ListOptions{
				PerPage: 30,
				Page:    page,
			},
		})
		if err != nil {
			log.Fatalf("list PRs: %v", err)
		}
		if len(prs) == 0 {
			break
		}
		for _, pr := range prs {
			if err := db.UpdateOrCreatePR(pr); err != nil {
				log.Printf("update PR %d: %v", pr.GetNumber(), err)
				continue
			}
			synced++
		}
		page++
	}
	log.Printf("Synced %d PRs. You can refresh the frontend.", synced)
}
