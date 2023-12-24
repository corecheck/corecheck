package main

import (
	"github.com/corecheck/corecheck/internal/db"
)

func GroupBenchmarks(benchmarks []db.BenchmarkResult) map[string][]db.BenchmarkResult {
	groupedBenchmarks := make(map[string][]db.BenchmarkResult)

	for _, benchmark := range benchmarks {
		groupedBenchmarks[benchmark.Name] = append(groupedBenchmarks[benchmark.Name], benchmark)
	}

	return groupedBenchmarks
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
