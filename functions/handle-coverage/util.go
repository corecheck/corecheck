package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/corecheck/corecheck/internal/db"
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

func GetCoverageData(prNum int, commit string) (*CoverageData, error) {
	return getCoverageData("https://bitcoin-coverage-data.s3.eu-west-3.amazonaws.com/" + strconv.Itoa(prNum) + "/" + commit + "/coverage.json")
}

func GetCoverageDataMaster(commit string) (*CoverageData, error) {
	return getCoverageData("https://bitcoin-coverage-data.s3.eu-west-3.amazonaws.com/master/" + commit + "/coverage.json")
}

func GetBaseCommit(prNum int, commit string) (string, error) {
	resp, err := http.Get("https://bitcoin-coverage-data.s3.eu-west-3.amazonaws.com/" + strconv.Itoa(prNum) + "/" + commit + "/.base_commit")
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

func ComputeCoverageRatio(lines []*db.CoverageLine, mustChange bool) *float64 {
	var covered, total int
	for _, line := range lines {
		if line.Testable && ((mustChange && line.Changed) || !mustChange) {
			if line.Covered {
				covered++
			}
			total++
		}
	}

	if total == 0 {
		return nil
	}

	r := float64(covered) / float64(total)
	return &r
}

func CreateCoverageFileHunks(lines []db.CoverageLine) []*db.CoverageFile {
	filesMap := map[string][]*db.CoverageHunk{}

	for _, line := range lines {
		if _, ok := filesMap[line.File]; !ok {
			filesMap[line.File] = []*db.CoverageHunk{}
		}

		if len(filesMap[line.File]) == 0 {
			filesMap[line.File] = append(filesMap[line.File], &db.CoverageHunk{})
		}

		// Create hunks
		currentHunk := filesMap[line.File][len(filesMap[line.File])-1]
		if len(currentHunk.Lines) > 0 && line.LineNumber != currentHunk.Lines[len(currentHunk.Lines)-1].LineNumber+1 {
			filesMap[line.File] = append(filesMap[line.File], &db.CoverageHunk{})
			currentHunk = filesMap[line.File][len(filesMap[line.File])-1]
		}

		currentHunk.Lines = append(currentHunk.Lines, line)
	}

	var files = make([]*db.CoverageFile, 0)
	for name, hunks := range filesMap {
		// remove chunks where all lines are untestable
		var newHunks []*db.CoverageHunk
		testableCount := 0
		testedCount := 0
		for _, hunk := range hunks {
			for _, line := range hunk.Lines {
				if line.Testable && line.Changed {
					testableCount++
				}
				if line.Testable && line.Changed && line.Covered {
					testedCount++
				}
			}

			newHunks = append(newHunks, hunk)
		}

		// Do not add file if all chunks are untestable
		if len(newHunks) == 0 {
			continue
		}

		var ratio float64 = 1
		if testableCount > 0 {
			ratio = float64(testedCount) / float64(testableCount)
		}

		files = append(files, &db.CoverageFile{
			Name:        name,
			Hunks:       newHunks,
			TestedRatio: ratio,
		})
	}

	// sort files by ratio and name
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].TestedRatio > files[j].TestedRatio {
				files[i], files[j] = files[j], files[i]
			} else if files[i].TestedRatio == files[j].TestedRatio && files[i].Name > files[j].Name {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	return files
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
