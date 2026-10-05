package main

import (
	"testing"

	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/types"
)

func line(content string, highlight bool) db.CoverageFileHunkLine {
	return db.CoverageFileHunkLine{Content: content, Highlight: highlight, Context: !highlight}
}

func TestFilterIgnoredBaselineLinesDropsUnexecutableHighlights(t *testing.T) {
	coverage := map[string]map[string][]db.CoverageFileHunk{
		types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE: {
			"src/compat/compat.h": {{
				Filename: "src/compat/compat.h",
				Lines: []db.CoverageFileHunkLine{
					line("#define WSAGetLastError()   errno", false),
					line("#define WSAEINVAL           EINVAL", true),
					line("#define WSAEWOULDBLOCK      EWOULDBLOCK", false),
				},
			}},
			"src/memusage.h": {{
				Filename: "src/memusage.h",
				Lines: []db.CoverageFileHunkLine{
					line("    } else {", false),
					line("        assert(0);", true),
				},
			}},
			"src/net.cpp": {{
				Filename: "src/net.cpp",
				Lines: []db.CoverageFileHunkLine{
					line("    int real = 1;", true),
				},
			}},
		},
		types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE: {
			"src/memusage.h": {{
				Filename: "src/memusage.h",
				Lines: []db.CoverageFileHunkLine{
					line("        assert(0);", true),
					line("    int still_real = 2;", true),
				},
			}},
		},
		types.COVERAGE_TYPE_GAINED_COVERAGE_NEW_CODE: {
			"src/compat/compat.h": {{
				Filename: "src/compat/compat.h",
				Lines: []db.CoverageFileHunkLine{
					line("#define WSAEINVAL           EINVAL", true),
				},
			}},
		},
	}

	filtered := FilterIgnoredBaselineLines(coverage)

	compat := filtered[types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE]["src/compat/compat.h"]
	if len(compat) != 0 {
		t.Fatalf("expected the WSAEINVAL hunk to be dropped, got %#v", compat)
	}
	if _, ok := filtered[types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE]["src/memusage.h"]; ok {
		t.Fatal("expected the memusage gained hunk to be dropped")
	}
	net := filtered[types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE]["src/net.cpp"]
	if len(net) != 1 || !net[0].Lines[0].Highlight {
		t.Fatalf("expected an unrelated gained baseline hunk to stay, got %#v", net)
	}

	lost := filtered[types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE]["src/memusage.h"]
	if len(lost) != 1 {
		t.Fatalf("expected the lost hunk to stay for the real highlight, got %#v", lost)
	}
	if len(lost[0].Lines) != 1 || lost[0].Lines[0].Content != "    int still_real = 2;" {
		t.Fatalf("expected only the real highlight to remain, got %#v", lost[0].Lines)
	}

	edited := filtered[types.COVERAGE_TYPE_GAINED_COVERAGE_NEW_CODE]["src/compat/compat.h"]
	if len(edited) != 1 || len(edited[0].Lines) != 1 {
		t.Fatalf("expected a real edit of the macro to stay, got %#v", edited)
	}
}

func TestFilterIgnoredBaselineLinesLeavesOtherTypesWhenBaselineIsEmpty(t *testing.T) {
	coverage := map[string]map[string][]db.CoverageFileHunk{
		types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE: {
			"src/compat/compat.h": {{
				Lines: []db.CoverageFileHunkLine{line("#define WSAEINVAL           EINVAL", true)},
			}},
		},
	}

	filtered := FilterIgnoredBaselineLines(coverage)
	if _, ok := filtered[types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE]; ok {
		t.Fatalf("expected an empty baseline type to be removed, got %#v", filtered)
	}
}
