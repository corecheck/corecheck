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

func handleMutationSuccess(job *types.JobParams) error {
	log.Info("Handling mutation testing success")

	mutation := db.MutationResult{
		Commit: job.Commit,
		State:  "completed",
	}

	err := db.CreateMutationResult(&mutation)
	if err != nil {
		log.Error("Error creating mutation result", err)
		return err
	}

	log.Infof("Mutation result created for %s", job.Commit)
	return nil
}

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

	err = handleMutationSuccess(params)
	if err != nil {
		log.Error("Error handling mutation success", err)
		return "", err
	}

	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}
