package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/corecheck/corecheck/internal/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-github/v57/github"
	"github.com/waigani/diffparser"
	"golang.org/x/oauth2"
)

var (
	cfg = Config{}
	log = logger.New()
)

func handleCodeCoverageSuccess(job *types.JobParams) error {
	log.Info("Handling code coverage success")

	var report *db.CoverageReport
	var coverage *CoverageData
	var lines []*db.CoverageLine
	var baseCommit string
	var err error
	if job.IsMaster {
		report, err = db.GetCoverageReportByCommitMaster(job.Commit)
		if err != nil {
			// if not exist, create a new report
			if err.Error() == "record not found" {
				log.Debugf("Creating new coverage report for master commit %s", job.Commit)
				report = &db.CoverageReport{
					Commit:   job.Commit,
					IsMaster: true,
				}

				err = db.CreateCoverageReport(report)
				if err != nil {
					return err
				}
			} else {
				log.Error("Error getting coverage report", err)
				return err
			}
		}

		baseCommit = job.Commit
		log.Debugf("Getting coverage data for master commit %s", job.Commit)
		coverage, err = GetCoverageDataMaster(job.Commit)
		if err != nil {
			log.Error("Error getting coverage data", err)
			return err
		}

		log.Debugf("Computing coverage for master commit %s", job.Commit)

		lines = computeMasterCoverage(report.ID, job, coverage)
	} else {
		report, err = db.GetCoverageReportByCommitPr(job.Commit, job.PRNumber)
		if err != nil {
			// if not exist, create a new report
			if err.Error() == "record not found" {
				log.Debugf("Creating new coverage report for PR %d", job.PRNumber)
				report = &db.CoverageReport{
					PRNumber: job.PRNumber,
					Commit:   job.Commit,
					IsMaster: false,
				}

				err = db.CreateCoverageReport(report)
				if err != nil {
					return err
				}
			} else {
				log.Error("Error getting coverage report", err)
				return err
			}
		}

		coverage, err = GetCoverageData(job.PRNumber, job.Commit)
		if err != nil {
			return err
		}
		log.Debugf("Getting diff for PR %d", job.PRNumber)
		diff, err := GetPullDiff(job.PRNumber)
		if err != nil {
			return err
		}

		lines = computeDiffCoverage(report.ID, coverage, diff)

		log.Debugf("Getting base commit for PR %d", job.PRNumber)
		baseCommit, err = GetBaseCommit(job.PRNumber, job.Commit)
		if err != nil {
			return err
		}
	}

	log.Debugf("Updating coverage data for PR %d", job.PRNumber)
	err = db.CreateLinesCoverage(report.ID, lines)
	if err != nil {
		return err
	}

	report.CoverageRatio = ComputeCoverageRatio(lines, !report.IsMaster)
	report.Status = db.COVERAGE_REPORT_STATUS_SUCCESS
	report.BaseCommit = baseCommit

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

type lineCoveredResult struct {
	Covered  bool
	Testable bool
}

func isLineCovered(coverage *CoverageData, file string, line int) lineCoveredResult {
	for _, f := range coverage.Files {
		if f.File == file {
			for _, l := range f.Lines {
				if l.LineNumber == line {
					return lineCoveredResult{
						Covered:  l.Count > 0,
						Testable: true,
					}
				}
			}

			return lineCoveredResult{
				Covered:  false,
				Testable: false,
			}
		}
	}
	return lineCoveredResult{
		Covered:  false,
		Testable: false,
	}
}

func computeDiffCoverage(reportID int, coverage *CoverageData, diff *diffparser.Diff) []*db.CoverageLine {
	var lines []*db.CoverageLine

	hunks := 0
	for _, file := range diff.Files {
		hunks += len(file.Hunks)
	}

	for _, file := range diff.Files {
		if isFileExcluded(file.NewName) {
			continue
		}

		for _, hunk := range file.Hunks {
			for _, diffLine := range hunk.NewRange.Lines {
				coveredResult := isLineCovered(coverage, file.NewName, diffLine.Number)
				lines = append(lines, &db.CoverageLine{
					CoverageReportID: reportID,
					File:             file.NewName,
					Line:             diffLine.Content,
					LineNumber:       diffLine.Number,
					Covered:          coveredResult.Covered,
					Testable:         coveredResult.Testable,
					Changed:          diffLine.Mode == diffparser.ADDED,
				})
			}
		}
	}

	return lines
}

func downloadSourceFile(commit string, file string, cache map[string]string) (string, error) {
	log.Debug("Downloading source file", file)
	if cache != nil && cache[file] != "" {
		log.Debug("Using cached file", file)
		return cache[file], nil
	}

	resp, err := http.Get("https://raw.githubusercontent.com/bitcoin/bitcoin/" + commit + "/" + file)
	if err != nil {
		log.Error("Error downloading source file", file, err)
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error("Error downloading source file: ", file, resp.StatusCode)
		return "", err
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error downloading source file: ", file, err)
		return "", err
	}

	if cache != nil {
		cache[file] = string(content)
	}
	log.Debug("Downloaded source file: ", file)
	return string(content), nil
}

func computeMasterCoverage(reportID int, job *types.JobParams, coverage *CoverageData) []*db.CoverageLine {
	var lines []*db.CoverageLine

	for _, file := range coverage.Files {
		if isFileExcluded(file.File) {
			continue
		}

		fileContent, err := downloadSourceFile(job.Commit, file.File, nil)
		if err != nil {
			log.Error("Error downloading source file", file.File, err)
			continue
		}

		contentSplit := strings.Split(fileContent, "\n")

		for _, line := range file.Lines {
			lines = append(lines, &db.CoverageLine{
				CoverageReportID: reportID,
				File:             file.File,
				Line:             contentSplit[line.LineNumber-1],
				LineNumber:       line.LineNumber,
				Covered:          line.Count > 0,
				Testable:         true,
				Changed:          false,
			})
		}
	}

	return lines
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
