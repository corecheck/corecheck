package main

import (
	"github.com/corecheck/corecheck/internal/types"
	"github.com/waigani/diffparser"
)

func isLineModifiedByDiff(filename string, lineNumber int, diff *diffparser.Diff) bool {
	for _, file := range diff.Files {
		if file.OrigName == filename {
			for _, hunk := range file.Hunks {
				for _, line := range hunk.WholeRange.Lines {
					if line.Number == lineNumber && line.Mode != diffparser.UNCHANGED {
						return true
					}
				}
			}
		}
	}

	return false
}

func ComputeDifferentialCoverage(masterCoverage, pullCoverage *CoverageData, diff *diffparser.Diff) *types.DifferentialCoverage {
	masterCoverageMap := masterCoverage.ToMap()
	pullCoverageMap := pullCoverage.ToMap()

	var diffCoverage types.DifferentialCoverage
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
					diffCoverage.LostBaselineCoverage = append(diffCoverage.LostBaselineCoverage, types.CoverageLine{
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
					diffCoverage.GainedBaselineCoverage = append(diffCoverage.GainedBaselineCoverage, types.CoverageLine{
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
						diffCoverage.GainedCoverageNewCode = append(diffCoverage.GainedCoverageNewCode, types.CoverageLine{
							OriginalLineNumber: -1,
							NewLineNumber:      line.Number,
							File:               file.NewName,
						})
					} else {
						// New code is not covered
						diffCoverage.UncoveredNewCode = append(diffCoverage.UncoveredNewCode, types.CoverageLine{
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
						diffCoverage.DeletedCoveredBaselineCode = append(diffCoverage.DeletedCoveredBaselineCode, types.CoverageLine{
							OriginalLineNumber: line.Number,
							NewLineNumber:      -1,
							File:               file.OrigName,
						})
					} else {
						// Deleted code was not covered
						diffCoverage.DeletedUncoveredBaselineCode = append(diffCoverage.DeletedUncoveredBaselineCode, types.CoverageLine{
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
				diffCoverage.GainedCoverageIncludedCode = append(diffCoverage.GainedCoverageIncludedCode, types.CoverageLine{
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
				diffCoverage.UncoveredIncludedCode = append(diffCoverage.UncoveredIncludedCode, types.CoverageLine{
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
				diffCoverage.ExcludedUncoveredBaselineCode = append(diffCoverage.ExcludedUncoveredBaselineCode, types.CoverageLine{
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
				diffCoverage.ExcludedCoveredBaselineCode = append(diffCoverage.ExcludedCoveredBaselineCode, types.CoverageLine{
					OriginalLineNumber: l.LineNumber,
					NewLineNumber:      r,
					File:               file.File,
				})
			}
		}
	}

	return &diffCoverage
}
