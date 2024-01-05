package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/corecheck/corecheck/internal/db"
)

func GetBenchData(prNum int, commit string, n int) ([]*db.BenchmarkResult, error) {
	return getBenchData(os.Getenv("BUCKET_DATA_URL") + "/" + strconv.Itoa(prNum) + "/" + commit + "/bench/bench-" + strconv.Itoa(n) + ".json")
}

func GetBenchDataMaster(commit string, n int) ([]*db.BenchmarkResult, error) {
	return getBenchData(os.Getenv("BUCKET_DATA_URL") + "/master/" + commit + "/bench/bench-" + strconv.Itoa(n) + ".json")
}

func getBenchData(url string) ([]*db.BenchmarkResult, error) {
	type benchData struct {
		Results []*db.BenchmarkResult `json:"results"`
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var benchResults benchData
	err = json.NewDecoder(resp.Body).Decode(&benchResults)
	if err != nil {
		return nil, err
	}

	return benchResults.Results, nil
}
