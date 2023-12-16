package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/corecheck/corecheck/internal/db"
)

func GetBenchData(prNum int, commit string, n int) ([]*db.BenchmarkResult, error) {
	return getBenchData("https://bitcoin-coverage-data-default.s3.eu-west-3.amazonaws.com/" + strconv.Itoa(prNum) + "/" + commit + "/bench/bench-" + strconv.Itoa(n) + ".json")
}

func GetBenchDataMaster(commit string, n int) ([]*db.BenchmarkResult, error) {
	return getBenchData("https://bitcoin-coverage-data-default.s3.eu-west-3.amazonaws.com/master/" + commit + "/bench/bench-" + strconv.Itoa(n) + ".json")
}

func getBenchData(url string) ([]*db.BenchmarkResult, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var benchResults []*db.BenchmarkResult
	err = json.NewDecoder(resp.Body).Decode(&benchResults)
	if err != nil {
		return nil, err
	}

	return benchResults, nil
}

func GroupBenchmarks(benchmarks []db.BenchmarkResult) map[string][]db.BenchmarkResult {
	groupedBenchmarks := make(map[string][]db.BenchmarkResult)

	for _, benchmark := range benchmarks {
		groupedBenchmarks[benchmark.Name] = append(groupedBenchmarks[benchmark.Name], benchmark)
	}

	return groupedBenchmarks
}
