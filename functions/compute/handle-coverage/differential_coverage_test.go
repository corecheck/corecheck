package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/types"
	"github.com/waigani/diffparser"
)

type coverageSectionExpectation struct {
	file       string
	lineNumber int
	content    string
	covered    bool
	tested     bool
}

func TestDiffAndCreateHunksCoverReportSections(t *testing.T) {
	masterCoverage := mustLoadCoverageFixture(t, filepath.Join("testdata", "master", "coverage.json"))
	pullCoverage := mustLoadCoverageFixture(t, filepath.Join("testdata", "pr", "coverage.json"))
	diff := mustLoadDiffFixture(t, filepath.Join("testdata", "diff.patch"))

	differentialCoverage := pullCoverage.Diff(masterCoverage, diff)

	expectedSections := map[string]coverageSectionExpectation{
		types.COVERAGE_TYPE_UNCOVERED_NEW_CODE: {
			file:       "src/uncovered_new_code.cpp",
			lineNumber: 2,
			content:    "int added_uncovered() { return 10; }",
			covered:    false,
			tested:     true,
		},
		types.COVERAGE_TYPE_GAINED_COVERAGE_NEW_CODE: {
			file:       "src/covered_new_code.cpp",
			lineNumber: 2,
			content:    "int added_covered() { return 20; }",
			covered:    true,
			tested:     true,
		},
		types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE: {
			file:       "src/lost_baseline.cpp",
			lineNumber: 2,
			content:    "    int stable = 1;",
			covered:    false,
			tested:     true,
		},
		types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE: {
			file:       "src/gained_baseline.cpp",
			lineNumber: 2,
			content:    "    int stable = 2;",
			covered:    true,
			tested:     true,
		},
		types.COVERAGE_TYPE_UNCOVERED_INCLUDED_CODE: {
			file:       "src/uncovered_included.cpp",
			lineNumber: 2,
			content:    "    int newly_tracked = 4;",
			covered:    false,
			tested:     true,
		},
		types.COVERAGE_TYPE_GAINED_COVERAGE_INCLUDED_CODE: {
			file:       "src/covered_included.cpp",
			lineNumber: 2,
			content:    "    int newly_tracked = 5;",
			covered:    true,
			tested:     true,
		},
	}

	assertExpectedCoverageResults(t, differentialCoverage.Results, expectedSections)

	masterSources := mustLoadSourceTree(t, filepath.Join("testdata", "master"))
	pullSources := mustLoadSourceTree(t, filepath.Join("testdata", "pr"))

	originalPullLoader := fetchPullSourceFiles
	originalMasterLoader := fetchMasterSourceFiles
	fetchPullSourceFiles = func(_ int, files []string, _ string) map[string][]string {
		return filterSourceTree(pullSources, files)
	}
	fetchMasterSourceFiles = func(files []string, _ string) map[string][]string {
		return filterSourceTree(masterSources, files)
	}
	t.Cleanup(func() {
		fetchPullSourceFiles = originalPullLoader
		fetchMasterSourceFiles = originalMasterLoader
	})

	report := &db.CoverageReport{
		ID:         1,
		PRNumber:   123,
		Commit:     "pull-commit",
		BaseCommit: "master-commit",
	}

	groupedHunks := groupHunksByCoverageType(differentialCoverage.CreateHunks(report))
	if len(groupedHunks) != len(expectedSections) {
		t.Fatalf("expected %d coverage sections with hunks, got %d", len(expectedSections), len(groupedHunks))
	}

	for coverageType, expectation := range expectedSections {
		hunks := groupedHunks[coverageType]
		if len(hunks) != 1 {
			t.Fatalf("expected 1 hunk for %s, got %d", coverageType, len(hunks))
		}

		hunk := hunks[0]
		if hunk.Filename != expectation.file {
			t.Fatalf("expected hunk for %s, got %s", expectation.file, hunk.Filename)
		}

		var highlighted *db.CoverageFileHunkLine
		contextLines := 0
		for i := range hunk.Lines {
			line := &hunk.Lines[i]
			if line.Highlight {
				highlighted = line
			} else if line.Context {
				contextLines++
			}
		}

		if highlighted == nil {
			t.Fatalf("expected highlighted line for %s", coverageType)
		}
		if highlighted.LineNumber != expectation.lineNumber {
			t.Fatalf("expected highlighted line %d for %s, got %d", expectation.lineNumber, coverageType, highlighted.LineNumber)
		}
		if highlighted.Content != expectation.content {
			t.Fatalf("expected highlighted content %q for %s, got %q", expectation.content, coverageType, highlighted.Content)
		}
		if highlighted.Covered != expectation.covered {
			t.Fatalf("expected covered=%t for %s, got %t", expectation.covered, coverageType, highlighted.Covered)
		}
		if highlighted.Tested != expectation.tested {
			t.Fatalf("expected tested=%t for %s, got %t", expectation.tested, coverageType, highlighted.Tested)
		}
		if highlighted.Context {
			t.Fatalf("expected highlighted line for %s to not be context", coverageType)
		}
		if contextLines == 0 {
			t.Fatalf("expected context lines for %s", coverageType)
		}
	}
}

func mustLoadCoverageFixture(t *testing.T, path string) *RawCoverageData {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read coverage fixture %s: %v", path, err)
	}

	var coverage RawCoverageData
	if err := json.Unmarshal(data, &coverage); err != nil {
		t.Fatalf("unmarshal coverage fixture %s: %v", path, err)
	}

	return &coverage
}

func mustLoadDiffFixture(t *testing.T, path string) *diffparser.Diff {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read diff fixture %s: %v", path, err)
	}

	diff, err := diffparser.Parse(string(data))
	if err != nil {
		t.Fatalf("parse diff fixture %s: %v", path, err)
	}

	return diff
}

func mustLoadSourceTree(t *testing.T, root string) map[string][]string {
	t.Helper()

	sourceTree := make(map[string][]string)
	err := filepath.WalkDir(filepath.Join(root, "src"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sourceTree[filepath.ToSlash(relativePath)] = strings.Split(string(data), "\n")
		return nil
	})
	if err != nil {
		t.Fatalf("walk source tree %s: %v", root, err)
	}

	return sourceTree
}

func filterSourceTree(sourceTree map[string][]string, files []string) map[string][]string {
	filtered := make(map[string][]string, len(files))
	for _, file := range files {
		if lines, ok := sourceTree[file]; ok {
			filtered[file] = lines
		}
	}
	return filtered
}

func assertExpectedCoverageResults(t *testing.T, results map[string]CoverageByFile, expected map[string]coverageSectionExpectation) {
	t.Helper()

	nonEmptyTypes := 0
	for coverageType, files := range results {
		if len(files) == 0 {
			continue
		}

		nonEmptyTypes++
		expectation, ok := expected[coverageType]
		if !ok {
			t.Fatalf("unexpected non-empty coverage type %s", coverageType)
		}
		if len(files) != 1 {
			t.Fatalf("expected 1 file for %s, got %d", coverageType, len(files))
		}

		lines := files[expectation.file]
		if len(lines) != 1 {
			t.Fatalf("expected 1 line for %s in %s, got %d", coverageType, expectation.file, len(lines))
		}

		line := lines[0]
		if line.File != expectation.file {
			t.Fatalf("expected result file %s for %s, got %s", expectation.file, coverageType, line.File)
		}
		if line.NewLineNumber != expectation.lineNumber {
			t.Fatalf("expected new line %d for %s, got %d", expectation.lineNumber, coverageType, line.NewLineNumber)
		}
	}

	if nonEmptyTypes != len(expected) {
		t.Fatalf("expected %d non-empty coverage types, got %d", len(expected), nonEmptyTypes)
	}
}

func groupHunksByCoverageType(hunks []*db.CoverageFileHunk) map[string][]*db.CoverageFileHunk {
	grouped := make(map[string][]*db.CoverageFileHunk)
	for _, hunk := range hunks {
		grouped[hunk.CoverageType] = append(grouped[hunk.CoverageType], hunk)
	}
	return grouped
}
