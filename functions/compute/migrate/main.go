package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
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

	err := db.DB.AutoMigrate(&db.CoverageReport{}, &db.CoverageFileHunkLine{}, &db.CoverageFileHunk{}, &db.BenchmarkResult{}, &db.PR{})
	if err != nil {
		log.Fatalf("Error migrating database: %s", err)
	}

	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}
