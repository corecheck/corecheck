package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/corecheck/corecheck/internal/db"
)

func GroupBenchmarks(benchmarks []db.BenchmarkResult) map[string][]db.BenchmarkResult {
	groupedBenchmarks := make(map[string][]db.BenchmarkResult)

	for _, benchmark := range benchmarks {
		groupedBenchmarks[benchmark.Name] = append(groupedBenchmarks[benchmark.Name], benchmark)
	}

	return groupedBenchmarks
}

func groupCoverageTypes(coverage []db.CoverageLine) map[string]map[string][]db.CoverageLine {
	groupedCoverage := make(map[string]map[string][]db.CoverageLine)

	for _, line := range coverage {
		if groupedCoverage[line.CoverageType] == nil {
			groupedCoverage[line.CoverageType] = make(map[string][]db.CoverageLine)
		}

		groupedCoverage[line.CoverageType][line.File] = append(groupedCoverage[line.CoverageType][line.File], line)
	}

	return groupedCoverage
}

func getSourceFile(filename string, commit string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/bitcoin/bitcoin/%s/%s", commit, filename))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func groupLinesByGap(lines []db.CoverageLine, maxGap int) [][]db.CoverageLine {
	var groupedLines [][]db.CoverageLine

	currentGroup := []db.CoverageLine{}
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if len(currentGroup) == 0 {
			currentGroup = append(currentGroup, line)
			continue
		}

		lastLine := currentGroup[len(currentGroup)-1]
		if line.OriginalLineNumber-lastLine.OriginalLineNumber <= maxGap {
			currentGroup = append(currentGroup, line)
		} else {
			groupedLines = append(groupedLines, currentGroup)
			currentGroup = []db.CoverageLine{}
		}
	}

	if len(currentGroup) > 0 {
		groupedLines = append(groupedLines, currentGroup)
	}

	return groupedLines
}

// For each coverage type, for each file, fetch the source file and create hunks
func createFileHunks(sourceCodeLines []string, filename string, commit string, lines []db.CoverageLine) []db.CoverageFileHunk {
	var fileHunks []db.CoverageFileHunk

	currentHunk := &db.CoverageFileHunk{
		Filename: filename,
	}

	// Group lines if they are next to each other (max 5 lines apart)
	groupedLines := groupLinesByGap(lines, 5)

	// For each group of lines, create a hunk with context (5 lines before and after)
	for _, group := range groupedLines {
		startLine := group[0].OriginalLineNumber - 5
		if startLine < 0 {
			startLine = 0
		}

		endLine := group[len(group)-1].OriginalLineNumber + 5
		if endLine > len(sourceCodeLines) {
			endLine = len(sourceCodeLines)
		}

		for i := startLine; i < endLine; i++ {
			highlight := false
			if containsLine(group, i+1) {
				highlight = true
			}

			currentHunk.Lines = append(currentHunk.Lines, db.CoverageFileHunkLine{
				LineNumber: i + 1,
				Content:    sourceCodeLines[i],
				Highlight:  highlight,
				Context:    isContextLine(i+1, group),
			})
		}

		fileHunks = append(fileHunks, *currentHunk)
		currentHunk = &db.CoverageFileHunk{
			Filename: filename,
		}
	}

	return fileHunks
}

func containsLine(lines []db.CoverageLine, lineNumber int) bool {
	for _, line := range lines {
		if line.OriginalLineNumber == lineNumber {
			return true
		}
	}
	return false
}

func isContextLine(lineNumber int, lines []db.CoverageLine) bool {
	for _, line := range lines {
		if line.OriginalLineNumber == lineNumber {
			return false
		}
	}
	return true
}

func containsFile(files []string, file string) bool {
	for _, f := range files {
		if f == file {
			return true
		}
	}
	return false
}

func getRequiredFiles(coverage []db.CoverageLine) []string {
	var files []string
	for _, line := range coverage {
		if !containsFile(files, line.File) {
			files = append(files, line.File)
		}
	}
	return files
}

func fetchAllFiles(files []string, commit string) map[string][]string {
	var wg sync.WaitGroup
	var filesMap = make(map[string][]string)

	wg.Add(len(files))
	for _, file := range files {
		go func(file string) {
			defer wg.Done()
			sourceCodeFile, err := getSourceFile(file, commit)
			if err != nil {
				log.Error(err)
				return
			}
			filesMap[file] = strings.Split(sourceCodeFile, "\n")
		}(file)
	}
	wg.Wait()

	return filesMap
}

func CreateCoverageHunks(report *db.CoverageReport) map[string][]db.CoverageFileHunk {
	requiredFiles := getRequiredFiles(report.CoverageLines)
	sourceFiles := fetchAllFiles(requiredFiles, report.Commit)

	groupedCoverage := groupCoverageTypes(report.CoverageLines)
	var coverageHunks = make(map[string][]db.CoverageFileHunk)

	for coverageType, files := range groupedCoverage {
		for filename, lines := range files {
			coverageHunks[coverageType] = append(coverageHunks[coverageType], createFileHunks(sourceFiles[filename], filename, report.Commit, lines)...)
		}
	}

	return coverageHunks
}
