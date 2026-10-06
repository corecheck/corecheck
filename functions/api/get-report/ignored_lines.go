package main

import (
	"strings"

	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/types"
)

// ignoredBaselineLine drops a lost or gained baseline highlight whose text
// cannot be executed on this coverage build, or whose counter does not stay
// on that line from one build to the next. The key is the filename plus
// the line text, so a later edit that moves the line does not keep a stale
// line number and does not hide a different line that lands on the old number.
type ignoredBaselineLine struct {
	filename string
	content  string
}

var ignoredBaselineLines = []ignoredBaselineLine{
	// Macro definition. LLVM still attributes a count to it on non-Windows
	// builds, and that count flips between the master run and the pull run.
	{filename: "src/compat/compat.h", content: "#define WSAEINVAL           EINVAL"},
	// Dead on 64-bit. The branch is sizeof(void*) neither 4 nor 8.
	// The opening brace of that branch is the only "} else {" in the header.
	// A bare "}" is left alone: that text is also the closing brace of live code.
	{filename: "src/memusage.h", content: "assert(0);"},
	{filename: "src/memusage.h", content: "} else {"},
	// EqualSharedPtrSock's return when one pointer is null. Builds disagree
	// on whether the counter sits on this statement or on the closing brace.
	{filename: "src/util/sock.h", content: "return false;"},
}

func isIgnoredBaselineLine(filename, content string) bool {
	trimmed := strings.TrimSpace(content)
	for _, line := range ignoredBaselineLines {
		if filename == line.filename && trimmed == line.content {
			return true
		}
	}
	return false
}

func hasHighlightedLine(lines []db.CoverageFileHunkLine) bool {
	for _, line := range lines {
		if line.Highlight {
			return true
		}
	}
	return false
}

func dropIgnoredHighlights(filename string, lines []db.CoverageFileHunkLine) []db.CoverageFileHunkLine {
	kept := make([]db.CoverageFileHunkLine, 0, len(lines))
	for _, line := range lines {
		if line.Highlight && isIgnoredBaselineLine(filename, line.Content) {
			continue
		}
		kept = append(kept, line)
	}
	return kept
}

// FilterIgnoredBaselineLines removes ignored highlights from lost and gained
// baseline hunks. A hunk with no highlight left is dropped. Other coverage
// types are left as they are, including a real edit of one of these lines.
func FilterIgnoredBaselineLines(coverage map[string]map[string][]db.CoverageFileHunk) map[string]map[string][]db.CoverageFileHunk {
	if coverage == nil {
		return coverage
	}

	for _, coverageType := range []string{
		types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE,
		types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE,
	} {
		files := coverage[coverageType]
		for filename, hunks := range files {
			keptHunks := make([]db.CoverageFileHunk, 0, len(hunks))
			for _, hunk := range hunks {
				hunk.Lines = dropIgnoredHighlights(filename, hunk.Lines)
				if !hasHighlightedLine(hunk.Lines) {
					continue
				}
				keptHunks = append(keptHunks, hunk)
			}
			if len(keptHunks) == 0 {
				delete(files, filename)
			} else {
				files[filename] = keptHunks
			}
		}
		if len(files) == 0 {
			delete(coverage, coverageType)
		}
	}

	return coverage
}
