package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/corecheck/corecheck/internal/types"
	"github.com/davecgh/go-spew/spew"
)

var (
	cfg = Config{}
	log = logger.New()
)

func HandleRequest(ctx context.Context, event map[string]interface{}) (string, error) {
	spew.Dump(event)
	log.Debug("Loading config...")
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	log.Debug("Connecting to database...")
	if err := db.Connect(cfg.DatabaseConfig); err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	params, err := types.GetJobParams(event)
	if err != nil {
		log.Error("Error getting job params", err)
		return "", err
	}

	spew.Dump(params)

	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}
