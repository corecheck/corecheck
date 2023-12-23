package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/waigani/diffparser"
)

type CoverageData struct {
	Files []struct {
		File      string `json:"file"`
		Functions []struct {
			ExecutionCount int    `json:"execution_count"`
			Lineno         int    `json:"lineno"`
			Name           string `json:"name"`
		} `json:"functions"`
		Lines []struct {
			Branches   []any `json:"branches"`
			Count      int   `json:"count"`
			LineNumber int   `json:"line_number"`
		} `json:"lines"`
	} `json:"files"`
}

type LineCoverage struct {
	LineNumber int
	Count      int
}

type CoverageMap map[string]map[int]LineCoverage

func (c CoverageData) ToMap() CoverageMap {
	m := make(CoverageMap)

	for _, file := range c.Files {
		m[file.File] = make(map[int]LineCoverage)
		for _, l := range file.Lines {
			m[file.File][l.LineNumber] = LineCoverage{
				LineNumber: l.LineNumber,
				Count:      l.Count,
			}
		}
	}

	return m
}

func GetCoverageData(prNum int, commit string) (*CoverageData, error) {
	return getCoverageData("https://bitcoin-coverage-data-default.s3.eu-west-3.amazonaws.com/" + strconv.Itoa(prNum) + "/" + commit + "/coverage.json")
}

func GetCoverageDataMaster(commit string) (*CoverageData, error) {
	return getCoverageData("https://bitcoin-coverage-data-default.s3.eu-west-3.amazonaws.com/master/" + commit + "/coverage.json")
}

func GetBaseCommit(prNum int, commit string) (string, error) {
	resp, err := http.Get("https://bitcoin-coverage-data-default.s3.eu-west-3.amazonaws.com/" + strconv.Itoa(prNum) + "/" + commit + "/.base_commit")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	baseCommit, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(string(baseCommit), "\n", ""), nil
}

func getCoverageData(url string) (*CoverageData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	coverageData := CoverageData{}
	err = json.NewDecoder(resp.Body).Decode(&coverageData)
	if err != nil {
		return nil, err
	}

	return &coverageData, nil
}

func GetPullDiff(prNum int) (*diffparser.Diff, error) {
	resp, err := http.Get("https://patch-diff.githubusercontent.com/raw/bitcoin/bitcoin/pull/" + strconv.Itoa(prNum) + ".diff")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return diffparser.Parse(string(data))
}
