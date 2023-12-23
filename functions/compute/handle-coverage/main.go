package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/corecheck/corecheck/internal/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

var (
	cfg = Config{}
	log = logger.New()
)

func handleCodeCoverageSuccess(job *types.JobParams) error {
	log.Info("Handling code coverage success")

	if job.GetIsMaster() {
		report, err := db.GetOrCreateCoverageReportByCommitMaster(job.Commit)
		if err != nil {
			log.Error("Error getting coverage report", err)
			return err
		}

		err = db.UpdateCoverageReport(report.ID, db.COVERAGE_REPORT_STATUS_SUCCESS, report.BenchmarkStatus, report.CoverageRatio, report.BaseCommit)
		if err != nil {
			log.Error("Error updating coverage report", err)
			return err
		}

		log.Infof("Coverage for master updated")
		return nil
	}

	report, err := db.GetOrCreateCoverageReportByCommitPr(job.Commit, job.GetPRNumber())
	if err != nil {
		log.Error("Error getting coverage report", err)
		return err
	}

	coverage, err := GetCoverageData(job.GetPRNumber(), job.Commit)
	if err != nil {
		log.Error("Error getting coverage data", err)
		return err
	}
	log.Debugf("Getting diff for PR %d", job.PRNumber)
	diff, err := GetPullDiff(job.GetPRNumber())
	if err != nil {
		return err
	}

	coverageMaster, err := GetCoverageDataMaster(job.BaseCommit)
	if err != nil {
		log.Error("Error getting master coverage data", err)
		return err
	}

	differentialCoverage := ComputeDifferentialCoverage(coverageMaster, coverage, diff)

	log.Debugf("Updating coverage data for PR %d", job.PRNumber)
	err = db.StoreDifferentialCoverage(report.ID, differentialCoverage)
	if err != nil {
		return err
	}

	report.Status = db.COVERAGE_REPORT_STATUS_SUCCESS

	err = db.UpdateCoverageReport(report.ID, report.Status, report.BenchmarkStatus, report.CoverageRatio, report.BaseCommit)
	if err != nil {
		return err
	}

	log.Infof("Coverage for PR %d updated", job.PRNumber)
	return nil
}

var excludedFolders = []string{
	"src/test",
	"src/qt/test",
	"src/wallet/test",
	"test",
	"src/bench",
}

var allowedFileExtensions = []string{
	".cpp",
	".h",
	".c",
}

func isFileExcluded(file string) bool {
	for _, folder := range excludedFolders {
		if strings.HasPrefix(file, folder) {
			return true
		}
	}

	for _, extension := range allowedFileExtensions {
		if strings.HasSuffix(file, extension) {
			return false
		}
	}

	return true
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

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Github.AccessToken},
	)
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	fmt.Println(client)

	params, err := types.GetJobParams(event)
	if err != nil {
		log.Error("Error getting job params", err)
		return "", err
	}

	err = handleCodeCoverageSuccess(params)
	if err != nil {
		log.Error("Error handling code coverage success", err)
		return "", err
	}

	return "OK", nil
}

func main() {
	lambda.Start(HandleRequest)
}
