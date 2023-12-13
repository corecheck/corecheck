package main

import "github.com/corecheck/corecheck/internal/db"

func GroupBenchmarks(benchmarks []db.BenchmarkResult) map[string][]db.BenchmarkResult {
	groupedBenchmarks := make(map[string][]db.BenchmarkResult)

	for _, benchmark := range benchmarks {
		groupedBenchmarks[benchmark.Name] = append(groupedBenchmarks[benchmark.Name], benchmark)
	}

	return groupedBenchmarks
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
