# Coverage pins for lines that flicker on corecheck

Sample: 64 successful corecheck reports, one latest report per pull request. A line is counted when its text showed up as a lost or gained baseline highlight. These are lines the pull request did not edit. Lost/gained baseline is a zero-versus-nonzero check, so a branch the suite only sometimes reaches flips between the master run and the pull-request run.

A pin works only if both runs execute the line. The tests below are deterministic: fixed inputs, no sleeps that race the scheduler, except the feeler delay which is bounded by a one-second window and is called directly. Unit tests and functional tests run one after the other. Each suite is parallel inside itself. Profiles are merged afterwards.

The bitcoin diff is `workers/coverage-worker/coverage-pins.patch`, taken from an uncommitted checkout of bitcoin master `09e22fb`. The coverage worker applies it after checkout, on both the master run and the pull-request run, and does not commit it. Unit tests run, then functional tests. Each suite is parallel inside itself. Profiles are merged afterwards.

Coverage line numbers come from that patched source. The diff stored for a pull request is the diff between the patched base and the patched pull request, so the pin patch is not listed as a pull-request change. A pull request that edits the same lines can make `git apply --check` fail; the worker logs that and continues with the unpatched tree.

Upstreaming the same tests cleans reports only after they are on the master base that corecheck rebases onto. Until then, the worker patch is what makes both sides execute the line. The two lines no test can execute are ignored in `functions/api/get-report/ignored_lines.go`.

Counts are "in how many of the 64 reports".

## Tests to add

| Reports | Code site | Fix | File |
| ---: | --- | --- | --- |
| 38 | `BaseIndex::Commit` logs a failure when no block has been indexed (`src/index/base.cpp`, the `if (!ok)` log) | `Commit()` is private. `Init()`, `Interrupt()` before any block is indexed, then `StartBackgroundSync()`. The sync thread commits a null best block and logs the failure | `src/test/baseindex_tests.cpp` (`index_commit_before_any_block`) |
| 38 | `JSONRPCRequest::parse` rejects params that are not an array or object (`src/rpc/request.cpp`) | Parse a request whose `params` is a string, and one whose `params` is a number | `src/test/rpc_tests.cpp` (`rpc_invalid_params_type`) |
| 35 | `ExecuteHTTPRPC` throws on a top-level JSON value that is not an object or batch (`src/httprpc.cpp`) | `ExecuteHTTPRPC` of `"hello"` and of `1`. Status stays 500. Body is code `-32700` | `src/test/rpc_tests.cpp` (`rpc_http_toplevel_parse_error`) |
| 35 | `estimateSmartFee` selects `DOUBLE_ESTIMATE`, and `StringForBlockPolicyEstimateReason` returns "Double Target 95% Threshold" (`src/policy/fees/block_policy_estimator.cpp`) | 80 blocks: 9/10 low-fee transactions confirm in 2 blocks, 1/10 is removed after 8 blocks, high-fee confirms in the next block. `estimateSmartFee(4, conservative=false)` | `src/test/blockpolicyestimator_tests.cpp` (`BlockPolicyDoubleEstimateReason`) |
| 35 | Anchor vectors are resized to 2 (`src/net.cpp`, previously three `resize` sites) | `LimitBlockRelayAnchors` holds the resize. The test passes 3 addresses and checks the size is 2. The three call sites use the helper | `src/net.h`, `src/net.cpp`, `src/test/coverage_pin_tests.cpp` (`anchor_limit_resize`) |
| 33 | `need_activate_chain` when a block has chain transactions and `BLOCK_VALID_TREE` but not `BLOCK_VALID_SCRIPTS`, and the `ActivateBestChain` call on `getblocks` (`src/net_processing.cpp`) | Clear the tip's script-validity bit, send `getdata` for it, then send `getblocks`. Restore the status word afterwards | `src/test/coverage_pin_tests.cpp` (`net_processing_activation_announcement_and_pong`) |
| 30 | `HTTPRemoteClient::MaybeDisconnect` returns false while a shutdown still has unsent data (`src/httpserver.cpp`) | Default client is connection-busy. `MaybeDisconnect(now, 0s, disconnect_all=true)` returns false | `src/test/coverage_pin_tests.cpp` (`http_peer_and_disconnect`) |
| 29 | "Announcing block not on main chain" (`src/net_processing.cpp`) | Mine one block so it is queued for a peer that does not prefer headers, move the active tip back without another announcement, then `SendMessages` | same case as the activation test |
| 27 | `GenericClusterImpl::Relinearize` marks a cluster `ACCEPTABLE` (`src/txgraph.cpp`) | `MakeTxGraph` with `acceptable_cost` 0, a chain of 8 transactions, `DoWork(1)` | `src/test/coverage_pin_tests.cpp` (`txgraph_acceptable_quality`) |
| 26 | `ReadCompactSize` throws "size too large" (`src/serialize.h`) | Canonical 5-byte encoding of `0x02000001`, with and without the range check | `src/test/serialize_tests.cpp` (`compactsize_too_large`) |
| 23–24 | Feeler branch: timer update and the short sleep plus log (`src/net.cpp`) | `SelectOutboundConnection` with full outbound slots and a due feeler timer. `FeelerSleep` once interrupted (returns immediately) and once not (finishes within 1 second) | `src/net.h`, `src/net.cpp`, `src/test/coverage_pin_tests.cpp` (`outbound_selection_feeler_and_extra_block_relay`, `feeler_sleep_interrupt_and_complete`) |
| 21 | `CreateTransaction` returns the empty insufficient-funds error (`src/wallet/spend.cpp`) | Ask for 21,000,000 BTC from a regtest wallet | `src/wallet/test/spend_tests.cpp` (`insufficient_funds_returns_empty_error`) |
| 21 | `PostLinearize` splices an independent higher-feerate group ahead of a lower one (`src/cluster_linearize.h`) | Two independent transactions, low then high. After `PostLinearize` the order is high then low | `src/test/cluster_linearize_tests.cpp` (`postlinearize_swaps_independent_higher_feerate`) |
| 21 | `InterruptibleRecv` returns `Interrupted` (`src/netbase.cpp`) | `Socks5` against a local socket that reads the handshake and then sets `g_socks5_interrupt` | `src/test/coverage_pin_tests.cpp` (`socks5_interrupted_and_network_error`) |
| 20 | `getblocktemplate` rejects a non-string `mode` (`src/rpc/mining.cpp`) | `getblocktemplate` with `{"mode": 1}` | `src/test/rpc_tests.cpp` (`rpc_getblocktemplate_invalid_mode`) |
| 18 | `LoadExternalBlockFile` continues when the 4-byte magic does not match (`src/validation.cpp`) | A file whose first byte is the network magic and whose next bytes are not | `src/test/coverage_pin_tests.cpp` (`load_external_block_file_magic_mismatch`) |
| 17 | `stats.presync_height` while a peer has a `HeadersSyncState` (`src/net_processing.cpp`) | Regtest node with `-minimumchainwork` above any regtest chain, then a full 2000-header message forking from genesis, then `GetNodeStateStats`. The unit-test chainstate setup ignores that argument unless it is copied onto the chainman options, which `src/test/util/setup_common.cpp` now does. The existing functional test `p2p_headers_sync_with_minchainwork.py` also reaches this line when it gets that far; this unit test does not depend on the rest of that scenario | `src/test/coverage_pin_tests.cpp` (`headers_presync_height`), `src/test/util/setup_common.cpp` |
| 15–16 | Extra block-relay-only timer (`src/net.cpp`) | `SelectOutboundConnection` with full slots, `start_extra_block_relay_peers` true, and a due extra-block-relay timer. This branch is ahead of the feeler branch | same outbound-selection test |
| 16 | `InterruptibleRecv` returns `NetworkError` (`src/netbase.cpp`) | `Socks5` against a local socket that reads the handshake and closes with `SO_LINGER` 0, so the next read is a connection reset | same SOCKS5 test |
| 14 | `getScriptFromDescriptor` takes `scripts.at(1)` for a 2-script combo (`src/rpc/mining.cpp`) | `generatetodescriptor` of an uncompressed `combo()` | `src/test/rpc_tests.cpp` (`rpc_generatetodescriptor_combo_and_give_up`) |
| 14 | `Sock::EqualSharedPtrSock` (`src/util/sock.h`) | Compare a socket with itself, with another socket, with null, and null with null | `src/test/sock_tests.cpp` (`equal_shared_ptr_sock`) |
| 13 | `Sock::IsConnected` reports "closed" on a zero-length peek (`src/util/sock.cpp`) | `socketpair`, close one end, `IsConnected` on the other | `src/test/sock_tests.cpp` (`is_connected_closed`) |
| 13 | Tor control thread logs "Lost connection to Tor control port" (`src/torcontrol.cpp`) | Point `TorController` at a loopback listener that accepts and then closes. No Tor process | `src/test/coverage_pin_tests.cpp` (`tor_control_lost_connection`) |
| 12 | `AddrManImpl::AddSingle` returns false for an entry already in the tried table (`src/addrman.cpp`) | `Good` an address, then `Add` it again with `nTime` one second newer | `src/test/addrman_tests.cpp` (`addrman_tried_entry_not_updated`) |
| 12 | `HTTPRequest::GetPeer` returns an empty service when the client weak pointer is empty (`src/httpserver.cpp`) | Default `HTTPRequest` | same HTTP test |
| 11 | `fRevertToInv` initial assignment when the peer does not prefer headers (`src/net_processing.cpp`) | The announcement test never sends `sendheaders`, so the first announcement takes this assignment | same net-processing test |
| 11 | `CConnman::DeleteNode` (`src/net.cpp`) | `StopNodes` on a connman that has one initialized test node | `src/test/coverage_pin_tests.cpp` (`open_network_connection_early_returns_and_delete_node`) |
| 10 | `ProcessPong` records "Timing mishap" when the pong timestamp is before the ping (`src/net_processing.cpp`) | Send a ping, move mock time backwards, send the matching pong | same net-processing test |
| 9 | Block assembly gives up after 1000 consecutive failures once the block is within 4000 weight of full (`src/node/miner.cpp`) | `block_reserved_weight` 4000, `block_max_weight` 5000, 1001 heavy transactions that do not fit, then one light transaction that would. The template stays coinbase-only | `src/test/coverage_pin_tests.cpp` (`miner_gives_up_near_full_block`) |
| 9 | `GenerateBlock` returns false when `max_tries` is already 0, and `generateBlocks` breaks (`src/rpc/mining.cpp`) | `generatetodescriptor` with `maxtries` 0. The same calls also hit `scripts.at(2)` for a compressed combo (4 scripts) | same generatetodescriptor test |
| 8 | SHA3 `Write` permutes on the fill path when the buffer completes exactly at the rate (`src/crypto/sha3.cpp`) | Write 129 bytes, then the remaining 7, of a 136-byte buffer. Digest matches one `Write` of all 136 | `src/test/crypto_tests.cpp` (`sha3_256_partial_buffer_rate_boundary`) |
| 6 | `CCoinsViewDB` destructor logs that it is waiting for compaction (`src/txdb.cpp`) | `SetTestCompactionHold` keeps the background thread parked until the destructor has observed a future that is not ready | `src/txdb.h`, `src/txdb.cpp`, `src/test/coverage_pin_tests.cpp` (`coinsview_destructor_waits_for_compaction`) |

`OpenNetworkConnection` also returns false when the network is inactive, when the connman is interrupted, and when the address is already connected. Those are the other early returns next to the lines above. Same test.

## Do not pin with a test

| Reports | Code site | What to do | File |
| ---: | --- | --- | --- |
| 36 | `read_atmost_n` has `return rbytes` after `while (1)`, and every path inside the loop already returns (`src/util/subprocess.h`) | Delete the unreachable return. A test cannot execute it. Leaving it in the file keeps an uncovered line in every report that looks at that header | `src/util/subprocess.h` |
| 17 | `#define WSAEINVAL EINVAL` (`src/compat/compat.h`) | Ignore. On this platform the line is a macro definition. Coverage attributes it as a gained line even though no test "runs" it. An ignore entry should key on the filename plus the exact line text, not a line number | `src/compat/compat.h` |
| 9 | `MallocUsage` `assert(0)` when `sizeof(void*)` is neither 4 nor 8 (`src/memusage.h`) | Ignore. This build is 64-bit, so the branch is dead. A test cannot take it without a different ABI | `src/memusage.h` |

Permanent HTTP read/send errors showed up in 3 reports. That is below the recurrence used for this list, so there is no test for them.

## Where this lives in corecheck

| Piece | File |
| --- | --- |
| Bitcoin diff, applied uncommitted inside the worker | `workers/coverage-worker/coverage-pins.patch` |
| Image copies the patch to `/coverage-pins.patch` | `workers/coverage-worker/Dockerfile` |
| Apply on both trees before cmake; record the patched-tree diff | `workers/coverage-worker/entrypoint.sh` |
| Drop the two unexecutable baseline highlights | `functions/api/get-report/ignored_lines.go` |
| This table, plus the accuracy snapshot from 2026-10-05 | `coverage-pins/` |

`SetTestCompactionHold` does nothing unless a test sets it. `SelectOutboundConnection` and `FeelerSleep` are the same decisions the open-connection thread already made; the thread calls them instead of inlining the branches.

A new master coverage run has to be generated before pull requests are compared against it. An old unpatched master report and a new patched pull-request report do not share line numbers in the files the patch edits.
