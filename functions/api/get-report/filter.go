package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/waigani/diffparser"
)

type HunkFilter struct {
	File   string
	Line   int
	Commit string
}

var hunkFilters = []HunkFilter{
	{
		File:   "src/addrman.cpp",
		Line:   568,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/addrman.cpp",
		Line:   74,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/crypto/sha3.cpp",
		Line:   120,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/crypto/sha3.cpp",
		Line:   121,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/init.cpp",
		Line:   1751,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/kernel/mempool_removal_reason.cpp",
		Line:   20,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   1630,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   1631,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   1632,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   1633,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   2116,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   2867,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   2925,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   3041,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   3042,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net.cpp",
		Line:   3043,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   1679,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   3581,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   3582,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4401,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4402,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4403,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4539,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4540,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4541,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4545,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4546,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4547,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4548,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4550,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4585,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4586,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4587,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4597,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4598,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4599,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4604,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4606,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   4813,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   5023,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   5611,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   5612,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   5700,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/net_processing.cpp",
		Line:   5971,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/node/miner.cpp",
		Line:   391,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/node/miner.cpp",
		Line:   393,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/policy/fees.cpp",
		Line:   922,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/policy/fees.cpp",
		Line:   923,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/policy/fees.cpp",
		Line:   924,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/policy/fees.cpp",
		Line:   925,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/pubkey.cpp",
		Line:   112,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/rpc/mining.cpp",
		Line:   136,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/rpc/mining.cpp",
		Line:   163,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/rpc/net.cpp",
		Line:   217,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/script/interpreter.cpp",
		Line:   1199,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/script/interpreter.cpp",
		Line:   471,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/script/script.h",
		Line:   455,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/serialize.h",
		Line:   375,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/tinyformat.h",
		Line:   691,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/tinyformat.h",
		Line:   735,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   124,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   126,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   127,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   130,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   131,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   132,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   136,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   137,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   139,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   178,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   180,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   181,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   182,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   183,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   621,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   624,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   626,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   627,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   630,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   633,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   634,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   635,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   636,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/torcontrol.cpp",
		Line:   76,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/util/time.cpp",
		Line:   125,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/util/time.cpp",
		Line:   128,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/util/time.cpp",
		Line:   129,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/util/time.cpp",
		Line:   130,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/wallet/spend.cpp",
		Line:   862,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/wallet/sqlite.cpp",
		Line:   577,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/wallet/sqlite.cpp",
		Line:   578,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/wallet/test/coinselector_tests.cpp",
		Line:   774,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/wallet/wallet.cpp",
		Line:   258,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/wallet/wallet.cpp",
		Line:   611,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/wallet/wallet.cpp",
		Line:   815,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/wallet/walletdb.cpp",
		Line:   1297,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/wallet/walletdb.cpp",
		Line:   1300,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/sync.h",
		Line:   172,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/sync.h",
		Line:   173,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/bitcoin-cli.cpp",
		Line:   876,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/blockencodings.cpp",
		Line:   152,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/blockencodings.cpp",
		Line:   153,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
	{
		File:   "src/tinyformat.h",
		Line:   898,
		Commit: "4b1196a9855dcd188a24f393aa2fa21e2d61f061",
	},
}

func FilterFlakyCoverageHunks(commit string, coverage map[string]map[string][]db.CoverageFileHunk) map[string]map[string][]db.CoverageFileHunk {
	var diffs = make(map[string]*diffparser.Diff)
	for _, filter := range hunkFilters {
		_, ok := diffs[filter.Commit]
		if !ok {
			diffData, err := http.Get(fmt.Sprintf("https://github.com/bitcoin/bitcoin/compare/%s..%s.diff", filter.Commit, commit))
			if err != nil {
				log.Errorf("Error getting diff: %s", err)
				continue
			}

			defer diffData.Body.Close()

			bodyBytes, err := io.ReadAll(diffData.Body)
			if err != nil {
				log.Errorf("Error reading diff: %s", err)
				continue
			}

			diffs[filter.Commit], err = diffparser.Parse(string(bodyBytes))
			if err != nil {
				log.Errorf("Error parsing diff: %s", err)
				continue
			}
		}
	}

	for _, hunkType := range []string{types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE, types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE} {
		for file, hunks := range coverage[hunkType] {
			coverage[hunkType][file] = filterCoverage(hunks, diffs)
		}

		for file, hunks := range coverage[hunkType] {
			if len(hunks) == 0 {
				delete(coverage[hunkType], file)
			}
		}
	}

	// delete empty coverage types
	for hunkType, files := range coverage {
		if len(files) == 0 {
			delete(coverage, hunkType)
		}
	}

	return coverage
}
func filterCoverage(hunks []db.CoverageFileHunk, diffs map[string]*diffparser.Diff) []db.CoverageFileHunk {
	var newLineCoverage []db.CoverageFileHunk

	for _, hunk := range hunks {
		if !shouldRemoveHunk(hunk, diffs) {
			newLineCoverage = append(newLineCoverage, hunk)
			continue
		}

		log.Infof("Removed hunk")
		spew.Dump(hunk)
	}

	return newLineCoverage
}
func shouldRemoveHunk(hunk db.CoverageFileHunk, diffs map[string]*diffparser.Diff) bool {
	for _, line := range hunk.Lines {
		if !line.Highlight {
			continue
		}

		for _, filter := range hunkFilters {
			translatedLine, err := diffs[filter.Commit].TranslateOriginalToNew(hunk.Filename, line.LineNumber)
			if err != nil {
				log.Warnf("Error translating line: %s", err)
				continue
			}
			if hunk.Filename == filter.File && translatedLine == filter.Line {
				return true
			}
		}
	}

	return false
}
