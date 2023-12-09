package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

var (
	cfg = Config{}
	log = logger.New()
)

func HandleRequest(ctx context.Context, event interface{}) (string, error) {
	log.Debug("Loading config...")
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	log.Debug("Connecting to database...")
	if err := db.Connect(cfg.DatabaseConfig); err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Github.AccessToken},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	fmt.Println(client)

	// parse event as json
	eventParsed := event.(map[string]interface{})
	spew.Dump(eventParsed)

	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}
