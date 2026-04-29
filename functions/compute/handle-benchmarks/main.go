package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/corecheck/corecheck/internal/config"
	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/logger"
	"github.com/corecheck/corecheck/internal/telemetry"
	"github.com/corecheck/corecheck/internal/types"
)

var (
	cfg = Config{}
	log = logger.New()
)

func handleBenchmarkSuccess(job *types.JobParams) error {
	log.Info("Handling benchmark success")

	var totalBenchmarks []*db.BenchmarkResult
	var report *db.CoverageReport
	var err error

	if job.GetIsMaster() {
		report, err = db.GetOrCreateCoverageReportByCommitMaster(job.Commit)
		if err != nil {
			log.Error("Error getting coverage report", err)
			return err
		}
	} else {
		report, err = db.GetOrCreateCoverageReportByCommitPr(job.Commit, job.GetPRNumber(), job.BaseCommit)
		if err != nil {
			log.Error("Error getting coverage report", err)
			return err
		}
	}

	for n := 0; n < cfg.BenchArraySize; n++ {
		log.Infof("Getting benchmark data for n=%d", n)
		var benchResults []*db.BenchmarkResult
		var err error
		if job.GetIsMaster() {
			log.Info("Getting benchmark data for master")
			benchResults, err = GetBenchDataMaster(job.Commit, n)
			if err != nil {
				log.Error("Error getting benchmark data", err)
				return err
			}
		} else {
			log.Info("Getting benchmark data for PR")
			benchResults, err = GetBenchData(job.GetPRNumber(), job.Commit, n)
			if err != nil {
				log.Error("Error getting benchmark data", err)
				return err
			}
		}

		log.Infof("Grouping benchmarks for n=%d", n)
		totalBenchmarks = append(totalBenchmarks, benchResults...)
	}

	if job.GetIsMaster() {
		metrics := telemetry.Default()
		resultsByBenchmark := make(map[string][]*db.BenchmarkResult)
		for _, result := range totalBenchmarks {
			resultsByBenchmark[result.Name] = append(resultsByBenchmark[result.Name], result)
		}

		for name, results := range resultsByBenchmark {
			log.Infof("Calculating benchmark stats for %s", name)
			avg := db.GetAverageBenchmarkResults(results)

			tags := []telemetry.Tag{
				telemetry.NewTag("benchmark_name", name),
			}

			metrics.Metric("bitcoin.bitcoin.benchmarks.batch", avg.Batch, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.complexity_n", avg.ComplexityN, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.epochs", avg.Epochs, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.clock_resolution", avg.ClockResolution, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.clock_resolution_multiple", avg.ClockResolutionMultiple, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.max_epoch_time", avg.MaxEpochTime, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.min_epoch_time", avg.MinEpochTime, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.min_epoch_iterations", avg.MinEpochIterations, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.epoch_iterations", avg.EpochIterations, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.warmup", avg.Warmup, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.relative", avg.Relative, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.median_elapsed", avg.MedianElapsed, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.median_absolute_percent_error_elapsed", avg.MedianAbsolutePercentErrorElapsed, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.median_instructions", avg.MedianInstructions, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.median_absolute_percent_error_instructions", avg.MedianAbsolutePercentErrorInstructions, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.median_cpucycles", avg.MedianCpucycles, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.median_contextswitches", avg.MedianContextswitches, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.median_pagefaults", avg.MedianPagefaults, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.median_branchinstructions", avg.MedianBranchinstructions, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.median_branchmisses", avg.MedianBranchmisses, tags...)
			metrics.Metric("bitcoin.bitcoin.benchmarks.total_time", avg.TotalTime, tags...)
		}

		metrics.Metric("bitcoin.bitcoin.benchmarks.count", float64(len(resultsByBenchmark)))
	}

	log.Info("Creating benchmark results")
	err = db.CreateBenchmarkResults(report.ID, totalBenchmarks)
	if err != nil {
		log.Error("Error creating benchmark results", err)
		return err
	}
	log.Info("Updating coverage report")

	err = db.UpdateCoverageReport(report.ID, report.Status, db.BENCHMARK_STATUS_SUCCESS, report.BaseCommit)
	if err != nil {
		log.Error("Error updating coverage report", err)
		return err
	}

	return nil
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

	err = handleBenchmarkSuccess(params)
	if err != nil {
		log.Error("Error handling benchmark success", err)
		return "", err
	}

	return "OK", nil
}

func main() {
	if err := telemetry.ConfigureDefaultFromEnv(); err != nil {
		log.Fatalf("Error configuring telemetry: %s", err)
	}

	lambda.Start(HandleRequest)
}
