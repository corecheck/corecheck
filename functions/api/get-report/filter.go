package main

import (
	"strings"

	"github.com/corecheck/corecheck/internal/db"
	"github.com/corecheck/corecheck/internal/types"
)

type HunkFilter struct {
	File    string
	Content string
}

func normalizeHunkContent(content string) string {
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, "\t", "")

	return content
}

var hunkFilters = []HunkFilter{
	{
		File: "src/net.cpp",
		Content: normalizeHunkContent(`CAddress addr(CService(), NODE_NONE);
            OpenNetworkConnection(addr, false, std::move(grant), info.m_params.m_added_node.c_str(), ConnectionType::MANUAL, info.m_params.m_use_v2transport);
            if (!interruptNet.sleep_for(std::chrono::milliseconds(500))) return;
            grant = CSemaphoreGrant(*semAddnode, /*fTry=*/true);
        }
        // Retry every 60 seconds if a connection was attempted, otherwise two seconds
        if (!interruptNet.sleep_for(std::chrono::seconds(tried ? 60 : 2)))`),
	},
	{
		File: "src/policy/fees.cpp",
		Content: normalizeHunkContent(`}
    double doubleEst = estimateCombinedFee(2 * confTarget, DOUBLE_SUCCESS_PCT, !conservative, &tempResult);
    if (doubleEst > median) {
        median = doubleEst;
        if (feeCalc) {
            feeCalc->est = tempResult;
            feeCalc->reason = FeeReason::DOUBLE_ESTIMATE;
        }
    }`),
	},
	{
		File: "src/pubkey.cpp",
		Content: normalizeHunkContent(`/* Integer tag byte for S */
    if (pos == inputlen || input[pos] != 0x02) {
        return 0;
    }
    pos++;
`),
	},
	{
		File: "src/bitcoin-cli.cpp",
		Content: normalizeHunkContent(`if (fWait) {
                const UniValue& error = response.find_value("error");
                if (!error.isNull() && error["code"].getInt<int>() == RPC_IN_WARMUP) {
                    throw CConnectionFailed("server in warmup");
                }
            }
            break; // Connection succeeded, no need to retry.`),
	},
	{
		File: "src/net.cpp",
		Content: normalizeHunkContent(`m_msgproc->SendMessages(pnode);
                if (flagInterruptMsgProc)
                    return;
            }
        }`),
	},
	{
		File: "src/node/blockstorage.cpp",
		Content: normalizeHunkContent(`{
    if (!m_block_tree_db->LoadBlockIndexGuts(
            GetConsensus(), [this](const uint256& hash) EXCLUSIVE_LOCKS_REQUIRED(cs_main) { return this->InsertBlockIndex(hash); }, m_interrupt)) {
        return false;
    }
    if (snapshot_blockhash) {`),
	},
	{
		File: "src/script/interpreter.cpp",
		Content: normalizeHunkContent(`opcode == OP_MOD ||
                opcode == OP_LSHIFT ||
                opcode == OP_RSHIFT)
                return set_error(serror, SCRIPT_ERR_DISABLED_OPCODE); // Disabled opcodes (CVE-2010-5137).
            // With SCRIPT_VERIFY_CONST_SCRIPTCODE, OP_CODESEPARATOR in non-segwit script is rejected even in an unexecuted branch
            if (opcode == OP_CODESEPARATOR && sigversion == SigVersion::BASE && (flags & SCRIPT_VERIFY_CONST_SCRIPTCODE))`),
	},
}

func FilterFlakyCoverageHunks(coverage map[string]map[string][]db.CoverageFileHunk) map[string]map[string][]db.CoverageFileHunk {
	filterCoverage := func(hunks []db.CoverageFileHunk) []db.CoverageFileHunk {
		var newLineCoverage []db.CoverageFileHunk

		for _, hunk := range hunks {
			if !shouldRemoveHunk(hunk) {
				newLineCoverage = append(newLineCoverage, hunk)
			}
		}

		return newLineCoverage
	}

	for _, hunkType := range []string{types.COVERAGE_TYPE_GAINED_BASELINE_COVERAGE, types.COVERAGE_TYPE_LOST_BASELINE_COVERAGE} {
		for file, hunks := range coverage[hunkType] {
			coverage[hunkType][file] = filterCoverage(hunks)
		}
	}

	return coverage
}

func shouldRemoveHunk(hunk db.CoverageFileHunk) bool {
	hunkLinesConcat := ""

	for _, line := range hunk.Lines {
		hunkLinesConcat += line.Content
	}

	for _, filter := range hunkFilters {
		if hunk.Filename == filter.File && normalizeHunkContent(hunkLinesConcat) == filter.Content {
			return true
		}
	}

	return false
}
