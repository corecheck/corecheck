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
	{
		File: "src/net_processing.cpp",
		Content: normalizeHunkContent(`LOCK(cs_main);
        const CNodeState* state = State(nodeid);
        if (state == nullptr)
            return false;
        stats.nSyncHeight = state->pindexBestKnownBlock ? state->pindexBestKnownBlock->nHeight : -1;
        stats.nCommonHeight = state->pindexLastCommonBlock ? state->pindexLastCommonBlock->nHeight : -1;
        for (const QueuedBlock& queue : state->vBlocksInFlight) {`),
	},
	{
		File: "src/policy/fees.cpp",
		Content: normalizeHunkContent(`if (doubleEst > median) {
        median = doubleEst;
        if (feeCalc) {
            feeCalc->est = tempResult;
            feeCalc->reason = FeeReason::DOUBLE_ESTIMATE;
        }
    }`),
	},
	{
		File: "src/rpc/mining.cpp",
		Content: normalizeHunkContent(`--max_tries;
    }
    if (max_tries == 0 || chainman.m_interrupt) {
        return false;
    }
    if (block.nNonce == std::numeric_limits<uint32_t>::max()) {
        return true;`),
	},
	{
		File: "src/rpc/mining.cpp",
		Content: normalizeHunkContent(`std::shared_ptr<const CBlock> block_out;
        if (!GenerateBlock(chainman, pblocktemplate->block, nMaxTries, block_out, /*process_new_block=*/true)) {
            break;
        }
        if (block_out) {`),
	},
	{
		File: "src/rpc/net.cpp",
		Content: normalizeHunkContent(`// peer got disconnected in between the GetNodeStats() and the GetNodeStateStats()
        // calls. In this case, the peer doesn't need to be reported here.
        if (!fStateStats) {
            continue;
        }
        obj.pushKV("id", stats.nodeid);
        obj.pushKV("addr", stats.m_addr_name);`),
	},
	{
		File: "src/sync.h",
		Content: normalizeHunkContent(`if (Base::try_lock()) {
            return true;
        }
        LeaveCritical();
        return false;
    }
public:`),
	},
	{
		File: "src/wallet/wallet.cpp",
		Content: normalizeHunkContent(`{
        WAIT_LOCK(g_wallet_release_mutex, lock);
        while (g_unloading_wallet_set.count(name) == 1) {
            g_wallet_release_cv.wait(lock);
        }
    }
}`),
	},
	{
		File: "src/init.cpp",
		Content: normalizeHunkContent(`}
    if (ShutdownRequested(node)) {
        return false;
    }
    // ********************************************************* Step 12: start node`),
	},
	{
		File: "src/net_processing.cpp",
		Content: normalizeHunkContent(`const bool processed_orphan = ProcessOrphanTx(*peer);
    if (pfrom->fDisconnect)
        return false;
    if (processed_orphan) return true;`),
	},
	{
		File: "src/net_processing.cpp",
		Content: normalizeHunkContent(`if (msg_type == NetMsgType::VERACK) {
        if (pfrom.fSuccessfullyConnected) {
            LogPrint(BCLog::NET, "ignoring redundant verack message from peer=%d\n", pfrom.GetId());
            return;
        }
        // Log successful connections unconditionally for outbound, but not for inbound as those`),
	},
	{
		File: "src/net_processing.cpp",
		Content: normalizeHunkContent(`// This should be very rare and could be optimized out.
                    // Just log for now.
                    if (m_chainman.ActiveChain()[pindex->nHeight] != pindex) {
                        LogPrint(BCLog::NET, "Announcing block %s not on main chain (tip=%s)\n",
                            hashToAnnounce.ToString(), m_chainman.ActiveChain().Tip()->GetBlockHash().ToString());
                    }`),
	},
	{
		File: "src/wallet/walletdb.cpp",
		Content: normalizeHunkContent(`it++;
        }
        if (it == vTxHashIn.end()) {
            break;
        }
        else if ((*it) == hash) {
            if(!EraseTx(hash)) {`),
	},
	{
		File: "src/serialize.h",
		Content: normalizeHunkContent(`throw std::ios_base::failure("non-canonical ReadCompactSize()");
    }
    if (range_check && nSizeRet > MAX_SIZE) {
        throw std::ios_base::failure("ReadCompactSize(): size too large");
    }
    return nSizeRet;
}`),
	},
	{
		File: "src/tinyformat.h",
		Content: normalizeHunkContent(`if (value > 0 && value <= numArgs)
                argIndex = value - 1;
            else
                TINYFORMAT_ERROR("tinyformat: Positional argument out of range");
            ++c;
            positionalMode = true;
        }`),
	},
	{
		File: "src/wallet/wallet.cpp",
		Content: normalizeHunkContent(`pMasterKey.second.nDeriveIterations = (pMasterKey.second.nDeriveIterations + static_cast<unsigned int>(pMasterKey.second.nDeriveIterations * target / (SteadyClock::now() - start))) / 2;
                if (pMasterKey.second.nDeriveIterations < 25000)
                    pMasterKey.second.nDeriveIterations = 25000;
                WalletLogPrintf("Wallet passphrase changed to an nDeriveIterations of %i\n", pMasterKey.second.nDeriveIterations);`),
	},
	{
		File: "src/node/miner.cpp",
		Content: normalizeHunkContent(`++nConsecutiveFailed;
            if (nConsecutiveFailed > MAX_CONSECUTIVE_FAILURES && nBlockWeight >
                    m_options.nBlockMaxWeight - 4000) {
                // Give up if we're close to full and haven't succeeded in a while
                break;
            }
            continue;
        }`),
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
