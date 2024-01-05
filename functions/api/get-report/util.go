package main

import (
	"github.com/corecheck/corecheck/internal/db"
)

func GroupBenchmarks(benchmarks []db.BenchmarkResult) map[string]*db.BenchmarkResult {
	groupedBenchmarks := make(map[string][]db.BenchmarkResult)

	for _, benchmark := range benchmarks {
		groupedBenchmarks[benchmark.Name] = append(groupedBenchmarks[benchmark.Name], benchmark)
	}

	averageBenchmarks := make(map[string]*db.BenchmarkResult)
	for name, benchmarks := range groupedBenchmarks {
		averageBenchmarks[name] = &db.BenchmarkResult{
			Name:  name,
			Title: benchmarks[0].Title,
			Unit:  benchmarks[0].Unit,
		}

		for _, benchmark := range benchmarks {
			averageBenchmarks[name].Batch += benchmark.Batch
			averageBenchmarks[name].ComplexityN += benchmark.ComplexityN
			averageBenchmarks[name].Epochs += benchmark.Epochs
			averageBenchmarks[name].ClockResolution += benchmark.ClockResolution
			averageBenchmarks[name].ClockResolutionMultiple += benchmark.ClockResolutionMultiple
			averageBenchmarks[name].MaxEpochTime += benchmark.MaxEpochTime
			averageBenchmarks[name].MinEpochTime += benchmark.MinEpochTime
			averageBenchmarks[name].MinEpochIterations += benchmark.MinEpochIterations
			averageBenchmarks[name].EpochIterations += benchmark.EpochIterations
			averageBenchmarks[name].Warmup += benchmark.Warmup
			averageBenchmarks[name].Relative += benchmark.Relative
			averageBenchmarks[name].MedianElapsed += benchmark.MedianElapsed
			averageBenchmarks[name].MedianAbsolutePercentErrorElapsed += benchmark.MedianAbsolutePercentErrorElapsed
			averageBenchmarks[name].MedianInstructions += benchmark.MedianInstructions
			averageBenchmarks[name].MedianAbsolutePercentErrorInstructions += benchmark.MedianAbsolutePercentErrorInstructions
			averageBenchmarks[name].MedianCpucycles += benchmark.MedianCpucycles
			averageBenchmarks[name].MedianContextswitches += benchmark.MedianContextswitches
			averageBenchmarks[name].MedianPagefaults += benchmark.MedianPagefaults
			averageBenchmarks[name].MedianBranchinstructions += benchmark.MedianBranchinstructions
			averageBenchmarks[name].MedianBranchmisses += benchmark.MedianBranchmisses
			averageBenchmarks[name].TotalTime += benchmark.TotalTime
		}

		averageBenchmarks[name].Batch /= float64(len(benchmarks))
		averageBenchmarks[name].ComplexityN /= float64(len(benchmarks))
		averageBenchmarks[name].Epochs /= float64(len(benchmarks))
		averageBenchmarks[name].ClockResolution /= float64(len(benchmarks))
		averageBenchmarks[name].ClockResolutionMultiple /= float64(len(benchmarks))
		averageBenchmarks[name].MaxEpochTime /= float64(len(benchmarks))
		averageBenchmarks[name].MinEpochTime /= float64(len(benchmarks))
		averageBenchmarks[name].MinEpochIterations /= float64(len(benchmarks))
		averageBenchmarks[name].EpochIterations /= float64(len(benchmarks))
		averageBenchmarks[name].Warmup /= float64(len(benchmarks))
		averageBenchmarks[name].Relative /= float64(len(benchmarks))
		averageBenchmarks[name].MedianElapsed /= float64(len(benchmarks))
		averageBenchmarks[name].MedianAbsolutePercentErrorElapsed /= float64(len(benchmarks))
		averageBenchmarks[name].MedianInstructions /= float64(len(benchmarks))
		averageBenchmarks[name].MedianAbsolutePercentErrorInstructions /= float64(len(benchmarks))
		averageBenchmarks[name].MedianCpucycles /= float64(len(benchmarks))
		averageBenchmarks[name].MedianContextswitches /= float64(len(benchmarks))
		averageBenchmarks[name].MedianPagefaults /= float64(len(benchmarks))
		averageBenchmarks[name].MedianBranchinstructions /= float64(len(benchmarks))
		averageBenchmarks[name].MedianBranchmisses /= float64(len(benchmarks))
		averageBenchmarks[name].TotalTime /= float64(len(benchmarks))
	}

	return averageBenchmarks
}

func GroupCoverageHunks(hunks []db.CoverageFileHunk) map[string]map[string][]db.CoverageFileHunk {
	groupedHunks := make(map[string]map[string][]db.CoverageFileHunk)

	for _, hunk := range hunks {
		if groupedHunks[hunk.CoverageType] == nil {
			groupedHunks[hunk.CoverageType] = make(map[string][]db.CoverageFileHunk)
		}

		groupedHunks[hunk.CoverageType][hunk.Filename] = append(groupedHunks[hunk.CoverageType][hunk.Filename], hunk)
	}

	return groupedHunks
}
