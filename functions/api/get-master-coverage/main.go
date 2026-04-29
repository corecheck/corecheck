package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/corecheck/corecheck/internal/api"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type MasterCoverageResponse struct {
	*db.CoverageReport
	ReportURL string `json:"report_url,omitempty"`
}

var (
	cfg = Config{}
	log = logger.New()
)

func getLatestMasterCoverage(c echo.Context) error {
	report, err := db.GetLatestMasterCoverageReport()
	if err != nil {
		log.Error(err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(404, "report not found")
		}
		return err
	}

	response := MasterCoverageResponse{
		CoverageReport: report,
	}
	if report.Status == db.COVERAGE_REPORT_STATUS_SUCCESS && report.GeneratedAt != nil {
		response.ReportURL = fmt.Sprintf("%s/master/%s/coverage-report/index.html", strings.TrimRight(cfg.BucketDataURL, "/"), report.Commit)
	}

	return c.JSON(200, response)
}

func main() {
	log.Debug("Loading config...")
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	log.Debug("Connecting to database...")
	if err := db.Connect(cfg.DatabaseConfig); err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	e := api.New()
	e.GET("/master-coverage", getLatestMasterCoverage)
	api.Start(e)
}
