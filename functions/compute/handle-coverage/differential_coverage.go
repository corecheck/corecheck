package main

import (
	"fmt"

	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/waigani/diffparser"
)

const (
	CONTEXT_LINES = 3
	MAX_GAP_LINES = 5
)

type CoverageLine struct {
	File               string
	OriginalLineNumber int
	NewLineNumber      int
}

type CoverageByFile map[string][]CoverageLine

type DifferentialCoverage struct {
	Coverage     CoverageMap
	BaseCoverage CoverageMap

	Results map[string]CoverageByFile
}

func isLineModifiedByDiff(filename string, lineNumber int, diff *diffparser.Diff) bool {
	for _, file := range diff.Files {
		if file.OrigName == filename {
			if isLineModifiedInHunks(lineNumber, file.Hunks) {
				return true
			}
		}
	}

	return false
}

func isLineModifiedInHunks(lineNumber int, hunks []*diffparser.DiffHunk) bool {
	for _, hunk := range hunks {
		for _, line := range hunk.WholeRange.Lines {
			if line.Number == lineNumber && line.Mode != diffparser.UNCHANGED {
				return true
			}
		}
	}

	return false
}

func (pullCoverage *RawCoverageData) Diff(masterCoverage *RawCoverageData, diff *diffparser.Diff) *DifferentialCoverage {
	masterCoverageMap := masterCoverage.ToMap()
	pullCoverageMap := pullCoverage.ToMap()

	diffCoverage := DifferentialCoverage{
		Coverage:     pullCoverageMap,
		BaseCoverage: masterCoverageMap,
	}

	diffCoverage.Results = make(map[string]CoverageByFile)
	for _, coverageType := range types.COVERAGE_TYPES {
		diffCoverage.Results[coverageType] = make(CoverageByFile)
	}

	for _, file := range masterCoverage.Files {
		for _, l := range file.Lines {
			// Master previously had coverage
			if l.Count > 0 {
				r, err := diff.TranslateOriginalToNew(file.File, l.LineNumber)
				if err != nil {
					continue
				}

				lineCoverage, ok := pullCoverageMap[file.File][r]
				if !ok {
					continue
				}

				// Pull has no coverage
				if lineCoverage.Count == 0 {
					diffCoverage.Results[types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE][file.File] = append(diffCoverage.Results[types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE][file.File], CoverageLine{
						OriginalLineNumber: l.LineNumber,
						NewLineNumber:      r,
						File:               file.File,
					})
				} else {
					// Pull still has coverage
					// Do nothing
				}
			} else { // Master previously had no coverage
				r, err := diff.TranslateOriginalToNew(file.File, l.LineNumber)
				if err != nil {
					continue
				}

				lineCoverage, ok := pullCoverageMap[file.File][r]
				if !ok {
					continue
				}

				// Now there is coverage
				if lineCoverage.Count > 0 {
					diffCoverage.Results[types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE][file.File] = append(diffCoverage.Results[types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE][file.File], CoverageLine{
						OriginalLineNumber: l.LineNumber,
						NewLineNumber:      r,
						File:               file.File,
					})
				} else {
					// Still no coverage
					// Do nothing
				}
			}
		}
	}

	for _, file := range diff.Files {
		for _, hunk := range file.Hunks {
			for _, line := range hunk.WholeRange.Lines {
				// New code
				if line.Mode == diffparser.ADDED {
					lineCoverage, ok := pullCoverageMap[file.NewName][line.Number]
					if !ok {
						continue
					}

					// New code is covered
					if lineCoverage.Count > 0 {
						diffCoverage.Results[types.COVERAGE_TYPE_GAINED_COVERAGE_NEW_CODE][file.NewName] = append(diffCoverage.Results[types.COVERAGE_TYPE_GAINED_COVERAGE_NEW_CODE][file.NewName], CoverageLine{
							OriginalLineNumber: -1,
							NewLineNumber:      line.Number,
							File:               file.NewName,
						})
					} else {
						// New code is not covered
						diffCoverage.Results[types.COVERAGE_TYPE_UNCOVERED_NEW_CODE][file.NewName] = append(diffCoverage.Results[types.COVERAGE_TYPE_UNCOVERED_NEW_CODE][file.NewName], CoverageLine{
							OriginalLineNumber: -1,
							NewLineNumber:      line.Number,
							File:               file.NewName,
						})
					}
				} else if line.Mode == diffparser.REMOVED {
					// Deleted code
					lineCoverage, ok := masterCoverageMap[file.OrigName][line.Number]
					if !ok {
						continue
					}

					// Deleted code was covered
					if lineCoverage.Count > 0 {
						diffCoverage.Results[types.COVERAGE_TYPE_DELETED_COVERED_BASELINE_CODE][file.OrigName] = append(diffCoverage.Results[types.COVERAGE_TYPE_DELETED_COVERED_BASELINE_CODE][file.OrigName], CoverageLine{
							OriginalLineNumber: line.Number,
							NewLineNumber:      -1,
							File:               file.OrigName,
						})
					} else {
						// Deleted code was not covered
						diffCoverage.Results[types.COVERAGE_TYPE_DELETED_UNCOVERED_BASELINE_CODE][file.OrigName] = append(diffCoverage.Results[types.COVERAGE_TYPE_DELETED_UNCOVERED_BASELINE_CODE][file.OrigName], CoverageLine{
							OriginalLineNumber: line.Number,
							NewLineNumber:      -1,
							File:               file.OrigName,
						})
					}
				}
			}
		}
	}

	for _, file := range pullCoverage.Files {
		for _, l := range file.Lines {
			// Pull has coverage
			if l.Count > 0 {
				r, err := diff.TranslateNewToOriginal(file.File, l.LineNumber)
				if err != nil {
					continue
				}

				// Check that the line was not in the diff
				if isLineModifiedByDiff(file.File, l.LineNumber, diff) {
					continue
				}

				_, ok := masterCoverageMap[file.File][r]
				if ok {
					continue // Line is not detected as code in master coverage
				}

				// Line was not in master, so it is new
				diffCoverage.Results[types.COVERAGE_TYPE_GAINED_COVERAGE_INCLUDED_CODE][file.File] = append(diffCoverage.Results[types.COVERAGE_TYPE_GAINED_COVERAGE_INCLUDED_CODE][file.File], CoverageLine{
					OriginalLineNumber: r,
					NewLineNumber:      l.LineNumber,
					File:               file.File,
				})
			} else { // Pull has no coverage
				r, err := diff.TranslateNewToOriginal(file.File, l.LineNumber)
				if err != nil {
					continue
				}

				// Check that the line was not in the diff
				if isLineModifiedByDiff(file.File, l.LineNumber, diff) {
					continue
				}

				_, ok := masterCoverageMap[file.File][r]
				if ok {
					continue // Line is not detected as code in master coverage
				}

				// Line was not in master, so it is new
				diffCoverage.Results[types.COVERAGE_TYPE_UNCOVERED_INCLUDED_CODE][file.File] = append(diffCoverage.Results[types.COVERAGE_TYPE_UNCOVERED_INCLUDED_CODE][file.File], CoverageLine{
					OriginalLineNumber: r,
					NewLineNumber:      l.LineNumber,
					File:               file.File,
				})
			}
		}
	}

	for _, file := range masterCoverage.Files {
		for _, l := range file.Lines {
			// Master has coverage
			if l.Count > 0 {
				r, err := diff.TranslateOriginalToNew(file.File, l.LineNumber)
				if err != nil {
					continue
				}

				// Check that the line was not in the diff
				if isLineModifiedByDiff(file.File, l.LineNumber, diff) {
					continue
				}

				_, ok := pullCoverageMap[file.File][r]
				if ok {
					continue // Line is not detected as code in pull coverage
				}

				// Line was not in pull, so it is new
				diffCoverage.Results[types.COVERAGE_TYPE_EXCLUDED_UNCOVERED_BASELINE_CODE][file.File] = append(diffCoverage.Results[types.COVERAGE_TYPE_EXCLUDED_UNCOVERED_BASELINE_CODE][file.File], CoverageLine{
					OriginalLineNumber: l.LineNumber,
					NewLineNumber:      r,
					File:               file.File,
				})
			} else { // Master has no coverage
				r, err := diff.TranslateOriginalToNew(file.File, l.LineNumber)
				if err != nil {
					continue
				}

				// Check that the line was not in the diff
				if isLineModifiedByDiff(file.File, l.LineNumber, diff) {
					continue
				}

				_, ok := pullCoverageMap[file.File][r]
				if ok {
					continue // Line is not detected as code in pull coverage
				}

				// Line was not in pull, so it is new
				diffCoverage.Results[types.COVERAGE_TYPE_EXCLUDED_COVERED_BASELINE_CODE][file.File] = append(diffCoverage.Results[types.COVERAGE_TYPE_EXCLUDED_COVERED_BASELINE_CODE][file.File], CoverageLine{
					OriginalLineNumber: l.LineNumber,
					NewLineNumber:      r,
					File:               file.File,
				})
			}
		}
	}

	return &diffCoverage
}

func groupLinesByGap(baseline bool, lines []CoverageLine, maxGap int) [][]CoverageLine {
	var groupedLines [][]CoverageLine
	currentGroup := []CoverageLine{}

	for _, line := range lines {
		if len(currentGroup) == 0 || isWithinGap(baseline, line, currentGroup[len(currentGroup)-1], maxGap) {
			currentGroup = append(currentGroup, line)
		} else {
			groupedLines = append(groupedLines, currentGroup)
			currentGroup = []CoverageLine{line}
		}
	}

	if len(currentGroup) > 0 {
		groupedLines = append(groupedLines, currentGroup)
	}

	return groupedLines
}

func isWithinGap(baseline bool, currentLine, lastLine CoverageLine, maxGap int) bool {
	if baseline {
		return currentLine.OriginalLineNumber-lastLine.OriginalLineNumber <= maxGap
	}
	return currentLine.NewLineNumber-lastLine.NewLineNumber <= maxGap
}

// For each coverage type, for each file, fetch the source file and create hunks
func (diffCoverage *DifferentialCoverage) createFileHunks(baseline bool, sourceCodeLines []string, filename string, commit string, lines []CoverageLine) []*db.CoverageFileHunk {
	var fileHunks []*db.CoverageFileHunk

	currentHunk := &db.CoverageFileHunk{
		Filename: filename,
	}

	// Group lines if they are next to each other (max 5 lines apart)
	groupedLines := groupLinesByGap(baseline, lines, MAX_GAP_LINES)

	spew.Dump("grouped lines", groupedLines)

	// For each group of lines, create a hunk with context
	for _, group := range groupedLines {
		startLine := group[0].OriginalLineNumber - (CONTEXT_LINES + 1)
		if !baseline {
			startLine = group[0].NewLineNumber - (CONTEXT_LINES + 1)
		}
		if startLine < 0 {
			startLine = 0
		}

		endLine := group[len(group)-1].OriginalLineNumber + CONTEXT_LINES
		if !baseline {
			endLine = group[len(group)-1].NewLineNumber + CONTEXT_LINES
		}
		if endLine > len(sourceCodeLines) {
			endLine = len(sourceCodeLines)
		}

		for i := startLine; i < endLine; i++ {
			highlight := false
			if containsLine(baseline, group, i+1) {
				highlight = true
			}

			currentHunk.Lines = append(currentHunk.Lines, db.CoverageFileHunkLine{
				LineNumber: i + 1,
				Content:    sourceCodeLines[i],
				Highlight:  highlight,
				Context:    isContextLine(baseline, i+1, group),
				Covered:    diffCoverage.Coverage.IsCovered(filename, i+1),
				Tested:     diffCoverage.Coverage.IsTested(filename, i+1),
			})
		}

		if len(currentHunk.Lines) > 0 {
			fileHunks = append(fileHunks, currentHunk)
			currentHunk = &db.CoverageFileHunk{
				Filename: filename,
			}
		}
	}

	return fileHunks
}

func containsLine(baseline bool, lines []CoverageLine, lineNumber int) bool {
	for _, line := range lines {
		if baseline && line.OriginalLineNumber == lineNumber {
			return true
		}
		if !baseline && line.NewLineNumber == lineNumber {
			return true
		}
	}
	return false
}

func isContextLine(baseline bool, lineNumber int, lines []CoverageLine) bool {
	for _, line := range lines {
		if baseline && line.OriginalLineNumber == lineNumber {
			return false
		}
		if !baseline && line.NewLineNumber == lineNumber {
			return false
		}
	}
	return true
}

func (diffCoverage *DifferentialCoverage) CreateHunks(report *db.CoverageReport) []*db.CoverageFileHunk {
	pullSourceFiles := fetchAllFiles(diffCoverage.Coverage.ListFiles(), report.Commit)
	masterSourceFiles := fetchAllFiles(diffCoverage.BaseCoverage.ListFiles(), report.BaseCommit)

	var coverageHunks []*db.CoverageFileHunk
	for coverageType, files := range diffCoverage.Results {
		isBaseLine := isBaselineHunk(coverageType)
		for filename, lines := range files {
			var hunks []*db.CoverageFileHunk
			fmt.Println(coverageType, isBaseLine)
			if isBaseLine {
				hunks = diffCoverage.createFileHunks(isBaseLine, masterSourceFiles[filename], filename, report.BaseCommit, lines)
			} else {
				hunks = diffCoverage.createFileHunks(isBaseLine, pullSourceFiles[filename], filename, report.Commit, lines)
			}
			for _, hunk := range hunks {
				hunk.CoverageType = coverageType
				hunk.CoverageReportID = report.ID
			}
			coverageHunks = append(coverageHunks, hunks...)
		}
	}

	return coverageHunks
}

func isBaselineHunk(coverageType string) bool {
	switch coverageType {
	case types.COVERAGE_TYPE_DELETED_COVERED_BASELINE_CODE, types.COVERAGE_TYPE_DELETED_UNCOVERED_BASELINE_CODE:
		return true
	}

	return false
}
