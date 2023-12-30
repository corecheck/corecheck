package main

import (
	"strconv"

	"github.com/corecheck/corecheck/internal/api"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/labstack/echo/v4"
)

var (
	cfg = Config{}
	log = logger.New()
)

func getReport(c echo.Context) error {
	reportID := c.QueryParam("id")
	pullNumber := c.Param("number")
	pullNumberInt, err := strconv.Atoi(pullNumber)
	if err != nil {
		log.Errorf("Error converting pull number to int: %s", err)
		return err
	}

	var report *db.CoverageReport

	if reportID == "" {
		// get latest
		report, err = db.GetLatestPullCoverageReport(pullNumberInt)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		i, err := strconv.Atoi(reportID)
		if err != nil {
			log.Errorf("Error converting report id to int: %s", err)
			return err
		}
		report, err = db.GetCoverageReport(i)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	report.BaseReport, err = db.GetMasterCoverageReport(report.BaseCommit)
	if err != nil {
		log.Errorf("Error getting base report: %s", err)
		return err
	}

	report.BenchmarksGrouped = GroupBenchmarks(report.Benchmarks)
	report.BaseReport.BenchmarksGrouped = GroupBenchmarks(report.BaseReport.Benchmarks)
	report.Coverage = GroupCoverageHunks(report.Hunks)
	report.Coverage = FilterFlakyCoverageHunks(report.Coverage)

	return c.JSON(200, report)
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
	e.GET("/pulls/:number/report", getReport)
	api.Start(e)
}
