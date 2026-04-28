package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/corecheck/corecheck/internal/types"
)

const defaultFailureReason = "The coverage workflow failed."

type coverageFailure struct {
	Error string `json:"Error"`
	Cause string `json:"Cause"`
}

var (
	cfg = Config{}
	log = logger.New()
)

func getCoverageReport(job *types.JobParams) (*db.CoverageReport, error) {
	if job.GetIsMaster() {
		return db.GetOrCreateCoverageReportByCommitMaster(job.Commit)
	}

	return db.GetOrCreateCoverageReportByCommitPr(job.Commit, job.GetPRNumber(), job.BaseCommit)
}

func getCoverageFailure(event map[string]interface{}) *coverageFailure {
	rawFailure, ok := event["coverage_error"]
	if !ok {
		return &coverageFailure{}
	}

	data, err := json.Marshal(rawFailure)
	if err != nil {
		log.Error("Error marshalling coverage failure payload", err)
		return &coverageFailure{}
	}

	failure := &coverageFailure{}
	if err := json.Unmarshal(data, failure); err != nil {
		log.Error("Error unmarshalling coverage failure payload", err)
		return &coverageFailure{}
	}

	return failure
}

func summarizeFailureReason(failure *coverageFailure) string {
	reason := extractReason(failure.Cause)
	if reason != "" {
		return reason
	}

	reason = extractReason(failure.Error)
	if reason != "" {
		return reason
	}

	return defaultFailureReason
}

func extractReason(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var payload interface{}
	if err := json.Unmarshal([]byte(value), &payload); err == nil {
		if reason := findReason(payload); reason != "" {
			return normalizeReason(reason)
		}
	}

	if mapped := mapStateError(value); mapped != "" {
		return mapped
	}

	return normalizeReason(value)
}

func findReason(payload interface{}) string {
	switch typed := payload.(type) {
	case map[string]interface{}:
		for _, key := range []string{
			"StatusReason",
			"statusReason",
			"errorMessage",
			"ErrorMessage",
			"Message",
			"message",
			"Cause",
			"cause",
		} {
			if value, ok := typed[key].(string); ok && strings.TrimSpace(value) != "" {
				return value
			}
		}

		for _, value := range typed {
			if reason := findReason(value); reason != "" {
				return reason
			}
		}
	case []interface{}:
		for _, value := range typed {
			if reason := findReason(value); reason != "" {
				return reason
			}
		}
	case string:
		return typed
	}

	return ""
}

func mapStateError(value string) string {
	switch value {
	case "States.Timeout":
		return "The coverage workflow timed out."
	case "States.TaskFailed":
		return "The coverage task failed."
	case "States.Permissions":
		return "The coverage workflow did not have permission to continue."
	default:
		return ""
	}
}

func normalizeReason(reason string) string {
	reason = strings.TrimSpace(reason)
	reason = strings.Trim(reason, "\"")
	reason = strings.Join(strings.Fields(reason), " ")
	if reason == "" {
		return defaultFailureReason
	}

	const maxReasonLength = 160
	if len(reason) > maxReasonLength {
		return strings.TrimSpace(reason[:maxReasonLength-1]) + "..."
	}

	return reason
}

func HandleRequest(ctx context.Context, event map[string]interface{}) (string, error) {
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

	report, err := getCoverageReport(params)
	if err != nil {
		log.Error("Error getting coverage report", err)
		return "", err
	}

	if err := db.UpdateCoverageReportTrace(report.ID, params.GetStepFunctionExecutionARN(), params.GetCoverageBatchJobID()); err != nil {
		log.Error("Error updating coverage report trace", err)
		return "", err
	}

	failureReason := summarizeFailureReason(getCoverageFailure(event))
	if err := db.UpdateCoverageReportFailure(report.ID, failureReason); err != nil {
		log.Error("Error updating coverage report failure", err)
		return "", err
	}

	log.Infof("Coverage report %d marked as failed: %s", report.ID, failureReason)
	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}
