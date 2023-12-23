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
