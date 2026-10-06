# Coverage pins for lines that flicker on corecheck

Sample: 64 successful corecheck reports, one latest report per pull request. A line is counted when its text showed up as a lost or gained baseline highlight. These are lines the pull request did not edit. Lost/gained baseline is a zero-versus-nonzero check, so a branch the suite only sometimes reaches flips between the master run and the pull-request run.

A pin works only if both runs execute the line. The tests below are deterministic: fixed inputs, no sleeps that race the scheduler, except the feeler delay which is bounded by a one-second window and is called directly, and the Tor controller thread which still waits up to a second for the peer to close. Unit tests run, then functional tests. Each suite is parallel inside itself. Profiles are merged afterwards.

The bitcoin diff is `workers/coverage-worker/coverage-pins.patch`, taken from an uncommitted checkout of bitcoin master `09e22fb`. The coverage worker applies it after checkout, on both the master run and the pull-request run, and does not commit it. Unit tests run, then functional tests. Each suite is parallel inside itself. Profiles are merged afterwards.

Coverage line numbers come from that patched source. The diff stored for a pull request is the diff between the patched base and the patched pull request, so the pin patch is not listed as a pull-request change. A pull request that edits the same lines can make `git apply --check` fail; the worker logs that and continues with the unpatched tree.

Upstreaming the same tests cleans reports only after they are on the master base that corecheck rebases onto. Until then, the worker patch is what makes both sides execute the line. Lines no test can execute, and lines whose counter moves between builds, are ignored in `functions/api/get-report/ignored_lines.go`.

`Reports` is how many of the 64 reports showed the line. `Masters` is the execution count on the two patched master profiles, `cf80493` generated 2026-10-05T22:48:39Z and then `acaf322` generated 2026-10-06T08:54:17Z. Both counts greater than zero means this pin is what keeps the line out of later reports: each side executes it, so the report does not mark it lost or gained. `src/net_processing.cpp` changed between those two commits, so those counts were matched by line text. Every other file in this table was identical.

## Tests to add

| Reports | Code site | Fix | File | Masters |
| ---: | --- | --- | --- | --- |
| 38 | `BaseIndex::Commit` logs a failure when no block has been indexed (`src/index/base.cpp`, the `if (!ok)` log) | `Commit()` is private. `Init()`, `Interrupt()` before any block is indexed, then `StartBackgroundSync()`. The sync thread commits a null best block and logs the failure | `src/test/baseindex_tests.cpp` (`index_commit_before_any_block`) | 1, 2 |
| 38 | `JSONRPCRequest::parse` rejects params that are not an array or object (`src/rpc/request.cpp`) | Parse a request whose `params` is a string, and one whose `params` is a number | `src/test/rpc_tests.cpp` (`rpc_invalid_params_type`) | 2, 2 |
| 35 | `ExecuteHTTPRPC` throws on a top-level JSON value that is not an object or batch (`src/httprpc.cpp`) | `ExecuteHTTPRPC` of `"hello"` and of `1`. Status stays 500. Body is code `-32700` | `src/test/rpc_tests.cpp` (`rpc_http_toplevel_parse_error`) | 2, 2 |
| 35 | `estimateSmartFee` selects `DOUBLE_ESTIMATE`, and `StringForBlockPolicyEstimateReason` returns "Double Target 95% Threshold" (`src/policy/fees/block_policy_estimator.cpp`) | 80 blocks: 9/10 low-fee transactions confirm in 2 blocks, 1/10 is removed after 8 blocks, high-fee confirms in the next block. `estimateSmartFee(4, conservative=false)` | `src/test/blockpolicyestimator_tests.cpp` (`BlockPolicyDoubleEstimateReason`) | 5, 1 |
| 35 | Anchor vectors are resized to 2 (`src/net.cpp`, previously three `resize` sites) | `LimitBlockRelayAnchors` holds the resize. The test passes 3 addresses and checks the size is 2. The three call sites use the helper | `src/net.h`, `src/net.cpp`, `src/test/coverage_pin_tests.cpp` (`anchor_limit_resize`) | 1, 1 |
| 33 | `need_activate_chain` when a block has chain transactions and `BLOCK_VALID_TREE` but not `BLOCK_VALID_SCRIPTS`, and the `ActivateBestChain` call on `getblocks` (`src/net_processing.cpp`) | Clear the tip's script-validity bit, send `getdata` for it, then send `getblocks`. Restore the status word afterwards | `src/test/coverage_pin_tests.cpp` (`net_processing_activation_announcement_and_pong`) | 1, 2 |
| 30 | `HTTPRemoteClient::MaybeDisconnect` returns false while a shutdown still has unsent data (`src/httpserver.cpp`) | Default client is connection-busy. `MaybeDisconnect(now, 0s, disconnect_all=true)` returns false | `src/test/coverage_pin_tests.cpp` (`http_peer_and_disconnect`) | 34, 97 |
| 29 | "Announcing block not on main chain" (`src/net_processing.cpp`) | Mine one block so it is queued for a peer that does not prefer headers, move the active tip back without another announcement, then `SendMessages` | same case as the activation test | 1, 1 |
| 27 | `GenericClusterImpl::Relinearize` marks a cluster `ACCEPTABLE` (`src/txgraph.cpp`) | `MakeTxGraph` with `acceptable_cost` 0, a chain of 8 transactions, `DoWork(1)`. Both master profiles executed `SetClusterQuality(..., ACCEPTABLE)` | `src/test/coverage_pin_tests.cpp` (`txgraph_acceptable_quality`) | 1, 1 |
| 26 | `ReadCompactSize` throws "size too large" (`src/serialize.h`) | Canonical 5-byte encoding of `0x02000001`, with and without the range check | `src/test/serialize_tests.cpp` (`compactsize_too_large`) | 1, 1 |
| 23–24 | Feeler branch: timer update and the short sleep plus log (`src/net.cpp`) | `SelectOutboundConnection` with full outbound slots and a due feeler timer. `FeelerSleep` once interrupted (returns immediately) and once not (finishes within 1 second) | `src/net.h`, `src/net.cpp`, `src/test/coverage_pin_tests.cpp` (`outbound_selection_feeler_and_extra_block_relay`, `feeler_sleep_interrupt_and_complete`) | 1, 1 |
| 21 | `CreateTransaction` returns the empty insufficient-funds error (`src/wallet/spend.cpp`) | Ask for 21,000,000 BTC from a regtest wallet. The test does not reach this return. Line 945 and the later empty `util::Error{}` are both still zero | `src/wallet/test/spend_tests.cpp` (`insufficient_funds_returns_empty_error`) | 0, 0 |
| 21 | `PostLinearize` splices an independent higher-feerate group ahead of a lower one (`src/cluster_linearize.h`) | Two independent transactions, low then high. After `PostLinearize` the order is high then low | `src/test/cluster_linearize_tests.cpp` (`postlinearize_swaps_independent_higher_feerate`) | 1, 1 |
| 21 | `InterruptibleRecv` returns `Interrupted` (`src/netbase.cpp`) | `Socks5` against a local socket that reads the handshake and then sets `g_socks5_interrupt` | `src/test/coverage_pin_tests.cpp` (`socks5_interrupted_and_network_error`) | 5, 4 |
| 20 | `getblocktemplate` rejects a non-string `mode` (`src/rpc/mining.cpp`) | `getblocktemplate` with `{"mode": 1}` | `src/test/rpc_tests.cpp` (`rpc_getblocktemplate_invalid_mode`) | 1, 1 |
| 18 | `LoadExternalBlockFile` continues when the 4-byte magic does not match (`src/validation.cpp`) | A file whose first byte is the network magic and whose next bytes are not | `src/test/coverage_pin_tests.cpp` (`load_external_block_file_magic_mismatch`) | 1, 1 |
| 17 | `stats.presync_height` while a peer has a `HeadersSyncState` (`src/net_processing.cpp`) | Regtest node with `-minimumchainwork` above any regtest chain, then a full 2000-header message forking from genesis, then `GetNodeStateStats`. The unit-test chainstate setup ignores that argument unless it is copied onto the chainman options, which `src/test/util/setup_common.cpp` now does. The existing functional test `p2p_headers_sync_with_minchainwork.py` also reaches this line when it gets that far; this unit test does not depend on the rest of that scenario | `src/test/coverage_pin_tests.cpp` (`headers_presync_height`), `src/test/util/setup_common.cpp` | 1, 2 |
| 15–16 | Extra block-relay-only timer (`src/net.cpp`) | `SelectOutboundConnection` with full slots, `start_extra_block_relay_peers` true, and a due extra-block-relay timer. This branch is ahead of the feeler branch | same outbound-selection test | 2, 2 |
| 16 | `InterruptibleRecv` returns `NetworkError` (`src/netbase.cpp`) | `Socks5` against a local socket that reads the handshake and closes with `SO_LINGER` 0, so the next read is a connection reset. The count is the permanent-error return. The `Wait` failure return is still 0, 0 | same SOCKS5 test | 1, 1 |
| 14 | `getScriptFromDescriptor` takes `scripts.at(1)` for a 2-script combo (`src/rpc/mining.cpp`) | `generatetodescriptor` of an uncompressed `combo()` | `src/test/rpc_tests.cpp` (`rpc_generatetodescriptor_combo_and_give_up`) | 2, 11 |
| 14 | `Sock::EqualSharedPtrSock` (`src/util/sock.h`) | Compare a socket with itself, with another socket, with null, and null with null. The mixed null/non-null `return false` does not keep a stable counter, so that line is ignored. The other comparisons in the function did run | `src/test/sock_tests.cpp` (`equal_shared_ptr_sock`) | ignored |
| 13 | `Sock::IsConnected` reports "closed" on a zero-length peek (`src/util/sock.cpp`) | `socketpair`, close one end, `IsConnected` on the other | `src/test/sock_tests.cpp` (`is_connected_closed`) | 2, 1 |
| 13 | Tor control thread logs "Lost connection to Tor control port" (`src/torcontrol.cpp`) | Point `TorController` at a loopback listener that accepts and then closes. No Tor process | `src/test/coverage_pin_tests.cpp` (`tor_control_lost_connection`) | 2, 1 |
| 12 | `AddrManImpl::AddSingle` returns false for an entry already in the tried table (`src/addrman.cpp`) | `Good` an address, then `Add` it again with `nTime` one second newer | `src/test/addrman_tests.cpp` (`addrman_tried_entry_not_updated`) | 1, 1 |
| 12 | `HTTPRequest::GetPeer` returns an empty service when the client weak pointer is empty (`src/httpserver.cpp`) | Default `HTTPRequest` | same HTTP test | 1, 1 |
| 11 | `fRevertToInv` initial assignment when the peer does not prefer headers (`src/net_processing.cpp`) | The announcement test never sends `sendheaders`, so the first announcement takes this assignment. The later `fRevertToInv = true` after a reorg is still 0, 0 | same net-processing test | 341027, 346801 |
| 11 | `CConnman::DeleteNode` (`src/net.cpp`) | `StopNodes` on a connman that has one initialized test node. That call is the live-node loop. The loop over `m_nodes_disconnected` is still 0, 0 | `src/test/coverage_pin_tests.cpp` (`open_network_connection_early_returns_and_delete_node`) | 770, 771 |
| 10 | `ProcessPong` records "Timing mishap" when the pong timestamp is before the ping (`src/net_processing.cpp`) | Send a ping, move mock time backwards, send the matching pong | same net-processing test | 1, 2 |
| 9 | Block assembly gives up after 1000 consecutive failures once the block is within 4000 weight of full (`src/node/miner.cpp`) | `block_reserved_weight` 4000, `block_max_weight` 5000, 1001 heavy transactions that do not fit, then one light transaction that would. The template stays coinbase-only | `src/test/coverage_pin_tests.cpp` (`miner_gives_up_near_full_block`) | 16, 22 |
| 9 | `GenerateBlock` returns false when `max_tries` is already 0, and `generateBlocks` breaks (`src/rpc/mining.cpp`) | `generatetodescriptor` with `maxtries` 0. The same calls also hit `scripts.at(2)` for a compressed combo (4 scripts) | same generatetodescriptor test | 3, 3 |
| 8 | SHA3 `Write` permutes on the fill path when the buffer completes exactly at the rate (`src/crypto/sha3.cpp`) | Write 129 bytes, then the remaining 7, of a 136-byte buffer. Digest matches one `Write` of all 136 | `src/test/crypto_tests.cpp` (`sha3_256_partial_buffer_rate_boundary`) | 3, 2 |
| 6 | `CCoinsViewDB` destructor logs that it is waiting for compaction (`src/txdb.cpp`) | `SetTestCompactionHold` keeps the background thread parked until the destructor has observed a future that is not ready | `src/txdb.h`, `src/txdb.cpp`, `src/test/coverage_pin_tests.cpp` (`coinsview_destructor_waits_for_compaction`) | 2, 1 |

`OpenNetworkConnection` also returns false when the network is inactive, when the connman is interrupted, and when the address is already connected. Those are the other early returns next to the lines above. Same test.

## Lines the first pin did not hold

Four pull-request reports were built with this patch on both trees against those masters (bitcoin 36445, 35901, 34844, 36446). They had 3, 6, 1, and 10 lost or gained baseline lines. Bitcoin 36425's uploaded tree does not contain the patch, so it is not in that set.

| Seen in those four | Code site | Masters before the follow-up | What changed |
| ---: | --- | --- | --- |
| 3 | `TorControlConnection::IsConnected` logs "Connection check failed" (`src/torcontrol.cpp`) | 1, 0 | `tor_control_connection_check_failed` connects to a loopback listener, accepts, closes the server, and calls `IsConnected` until the peek reports the socket closed. The lost-connection thread test stays; it is what held that other log |
| 3 | `EqualSharedPtrSock` `return false` (`src/util/sock.h`) | 0, 1337 | Ignore. The counter sits on the statement in one build and on the closing brace in the other. This is the only `return false` in the header |
| 1 | `MallocUsage` `} else {` around the dead `assert(0)` (`src/memusage.h`) | 0, 0 | Ignore. One pull request still counted the dead branch. A bare `}` is not ignored, because that text is also a live closing brace |

These showed up once and are not pinned. The disconnected-list `DeleteNode` is 0, 0. `ProcessMessages` returns false when `fDisconnect` is set on 0, then 4. The reorg `fRevertToInv = true` is 0, 0. `CheckMinimalPush`'s final `return true` is 0, 0. An addrman `nTime` update is 9, 18 on the masters and was missing on one pull request; the tried-table return this pin added is the 1, 1 row above.

## Do not pin with a test

| Reports | Code site | What to do | File | Masters |
| ---: | --- | --- | --- | --- |
| 36 | `read_atmost_n` has `return rbytes` after `while (1)`, and every path inside the loop already returns (`src/util/subprocess.h`) | Delete the unreachable return. A test cannot execute it. Leaving it in the file keeps an uncovered line in every report that looks at that header | `src/util/subprocess.h` | removed |
| 17 | `#define WSAEINVAL EINVAL` (`src/compat/compat.h`) | Ignore. On this platform the line is a macro definition. Coverage attributes it as a gained line even though no test "runs" it. An ignore entry should key on the filename plus the exact line text, not a line number | `src/compat/compat.h` | ignored |
| 9 | `MallocUsage` `assert(0)` when `sizeof(void*)` is neither 4 nor 8 (`src/memusage.h`) | Ignore. This build is 64-bit, so the branch is dead. A test cannot take it without a different ABI. The `} else {` above that assert is ignored too | `src/memusage.h` | 0, 0 |

Permanent HTTP read/send errors showed up in 3 reports. That is below the recurrence used for this list, so there is no test for them.

## Where this lives in corecheck

| Piece | File |
| --- | --- |
| Bitcoin diff, applied uncommitted inside the worker | `workers/coverage-worker/coverage-pins.patch` |
| Image copies the patch to `/coverage-pins.patch` | `workers/coverage-worker/Dockerfile` |
| Apply on both trees before cmake; record the patched-tree diff | `workers/coverage-worker/entrypoint.sh` |
| Drop baseline highlights that are dead or whose counter moves between builds | `functions/api/get-report/ignored_lines.go` |
| This table, plus the accuracy snapshot from 2026-10-05 | `coverage-pins/` |

`SetTestCompactionHold` does nothing unless a test sets it. `SelectOutboundConnection` and `FeelerSleep` are the same decisions the open-connection thread already made; the thread calls them instead of inlining the branches. `src/net.cpp` was still the same bytes on `acaf322` as on `09e22fb`, so that hunk still applies.

A new master coverage run has to be generated before pull requests are compared against it. An old unpatched master report and a new patched pull-request report do not share line numbers in the files the patch edits.
