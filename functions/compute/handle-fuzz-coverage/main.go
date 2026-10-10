package main

import (
	"context"
	"time"

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

func handleFuzzCoverageSuccess(job *types.JobParams) error {
	log.Info("Handling fuzz coverage success")
	stepFunctionExecutionARN := job.GetStepFunctionExecutionARN()
	coverageBatchJobID := job.GetCoverageBatchJobID()

	report, err := db.GetOrCreateCoverageReportByCommitMasterFuzz(job.Commit)
	if err != nil {
		log.Error("Error getting fuzz coverage report", err)
		return err
	}

	err = db.UpdateCoverageReportTrace(report.ID, stepFunctionExecutionARN, coverageBatchJobID)
	if err != nil {
		log.Error("Error updating fuzz coverage report trace", err)
		return err
	}

	err = db.UpdateCoverageReport(report.ID, db.COVERAGE_REPORT_STATUS_SUCCESS, report.BenchmarkStatus, report.BaseCommit)
	if err != nil {
		log.Error("Error updating fuzz coverage report", err)
		return err
	}

	err = db.UpdateCoverageReportGeneratedAt(report.ID, time.Now().UTC())
	if err != nil {
		log.Error("Error setting fuzz coverage report generation time", err)
		return err
	}

	log.Infof("Fuzz coverage for master updated")
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

	err = handleFuzzCoverageSuccess(params)
	if err != nil {
		log.Error("Error handling fuzz coverage success", err)
		return "", err
	}

	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}
