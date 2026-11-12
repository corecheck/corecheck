#!/usr/bin/env bash

# Proves apply_test_patches skips a bad patch, applies a pristine set,
# reports patches whose requires name a missing helper, and reports a
# test-to-test dependency when 21-...patch is absent. Also checks the JSON
# record and the pull-request diff's base_missing case on a tiny repo.
# Temp directories come only from mktemp under ${TMPDIR:-/tmp}.

FAILS=0
TMP_PARENT="${TMPDIR:-/tmp}"
WORK=$(mktemp -d "${TMP_PARENT}/coverage-apply-test.XXXXXX")

cleanup() {
    if [ -n "${WORK:-}" ] && [ -d "$WORK" ]; then
        case "$WORK" in
            "${TMP_PARENT}"/*) rm -rf "$WORK" ;;
        esac
    fi
}
trap cleanup EXIT

mkdir -p "$WORK/tmp"
export TMPDIR="$WORK/tmp"
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_PREFIX
set -u

pass() { echo "PASS: $1"; }
fail() { echo "FAIL: $1"; FAILS=$((FAILS + 1)); }

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
WORKER_DIR=$(cd "$SCRIPT_DIR/.." && pwd)
PATCHES="$WORKER_DIR/patches"
LIB="$WORKER_DIR/apply-test-patches.sh"
ENTRY="$WORKER_DIR/entrypoint.sh"
BITCOIN_WT="${BITCOIN_WT:-/home/max/source/wt-t_807957db/bitcoin-master}"

if [ ! -d "$BITCOIN_WT" ]; then
    echo "FAIL: BITCOIN_WT is not a directory: $BITCOIN_WT"
    exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
    echo "FAIL: jq is not on PATH"
    exit 1
fi
if ! command -v git >/dev/null 2>&1; then
    echo "FAIL: git is not on PATH"
    exit 1
fi

prefix_count() {
    grep -c '^COVERAGE_TEST_PATCH_MISSING:' "$1" 2>/dev/null || true
}

# --- syntax and "source does nothing" ---------------------------------------

if bash -n "$ENTRY"; then
    pass "bash -n entrypoint.sh"
else
    fail "bash -n entrypoint.sh"
fi
if bash -n "$LIB"; then
    pass "bash -n apply-test-patches.sh"
else
    fail "bash -n apply-test-patches.sh"
fi
if bash -n "$0"; then
    pass "bash -n test-apply-patches.sh"
else
    fail "bash -n test-apply-patches.sh"
fi

if grep -q 'apply_coverage_pins' "$ENTRY" || grep -q 'coverage-pins.patch' "$ENTRY"; then
    fail "entrypoint.sh still mentions the pin patch"
else
    pass "entrypoint.sh has no pin-patch mechanism"
fi
if grep -q 'COVERAGE_TEST_PATCH_MISSING:' "$ENTRY"; then
    fail "entrypoint.sh prints COVERAGE_TEST_PATCH_MISSING"
else
    pass "entrypoint.sh does not print the missing-patch prefix"
fi
if [ ! -e "$WORKER_DIR/coverage-pins.patch" ]; then
    pass "coverage-pins.patch is deleted"
else
    fail "coverage-pins.patch is still present"
fi

(
    set -euo pipefail
    # shellcheck disable=SC1090
    source "$LIB"
    if [ -n "${TEST_PATCH_SET_VERSION+x}" ] || [ -n "${TEST_PATCH_APPLIED+x}" ]; then
        exit 1
    fi
)
if [ "$?" -eq 0 ]; then
    pass "sourcing the library runs nothing"
else
    fail "sourcing the library runs nothing"
fi

# --- hand-made JSON records -------------------------------------------------

(
    set -euo pipefail
    # shellcheck disable=SC1090
    source "$LIB"
    TEST_PATCH_APPLIED=()
    TEST_PATCH_FAILED_NAME=("22-net-stop-nodes-anchor-truncation.patch")
    TEST_PATCH_FAILED_REASON=("does not apply")
    TEST_PATCH_BASE_MISSING=("12-blockpolicy-reason-strings.patch")
    TEST_PATCH_SET_VERSION="determinism-2026-11-12"
    TEST_PATCH_BITCOIN_COMMIT="a06515e9fdf81a762804471733b1f43a13e1f10b"
    write_tests_patches_record "$WORK/hand.json" "deadbeef" "false"

    TEST_PATCH_APPLIED=()
    TEST_PATCH_FAILED_NAME=()
    TEST_PATCH_FAILED_REASON=()
    TEST_PATCH_BASE_MISSING=()
    TEST_PATCH_SET_VERSION="empty-set"
    TEST_PATCH_BITCOIN_COMMIT="bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
    write_tests_patches_record "$WORK/empty.json" "mastercommit" "true"

    TEST_PATCH_FAILED_NAME=("a.patch" "b.patch")
    TEST_PATCH_FAILED_REASON=("does not apply" "depends on 00-test-helpers.patch")
    TEST_PATCH_BASE_MISSING=()
    TEST_PATCH_SET_VERSION="two"
    TEST_PATCH_BITCOIN_COMMIT="cccccccccccccccccccccccccccccccccccccccc"
    write_tests_patches_record "$WORK/two.json" "head" "false"
)
if [ "$?" -eq 0 ]; then
    pass "write_tests_patches_record exits 0"
else
    fail "write_tests_patches_record exits 0"
fi

if jq -e '
    .set_version == "determinism-2026-11-12"
    and .bitcoin_commit == "a06515e9fdf81a762804471733b1f43a13e1f10b"
    and .head_commit == "deadbeef"
    and .is_master == false
    and (.is_master | type) == "boolean"
    and (.generated_at | test("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z$"))
    and .applied == []
    and .failed == [{"patch":"22-net-stop-nodes-anchor-truncation.patch","reason":"does not apply"}]
    and .base_missing == ["12-blockpolicy-reason-strings.patch"]
    and (keys | sort) == ["applied","base_missing","bitcoin_commit","failed","generated_at","head_commit","is_master","set_version"]
' "$WORK/hand.json" >/dev/null; then
    pass "hand-made failed record matches the schema"
else
    fail "hand-made failed record matches the schema"
    cat "$WORK/hand.json" 2>/dev/null || true
fi

if jq -e '
    .applied == [] and .failed == [] and .base_missing == []
    and .is_master == true and .head_commit == "mastercommit"
    and .set_version == "empty-set"
' "$WORK/empty.json" >/dev/null; then
    pass "zero applied and zero failed is valid JSON"
else
    fail "zero applied and zero failed is valid JSON"
    cat "$WORK/empty.json" 2>/dev/null || true
fi

if jq -e '
    .failed == [
        {"patch":"a.patch","reason":"does not apply"},
        {"patch":"b.patch","reason":"depends on 00-test-helpers.patch"}
    ]
' "$WORK/two.json" >/dev/null; then
    pass "two failed reasons stay paired"
else
    fail "two failed reasons stay paired"
    cat "$WORK/two.json" 2>/dev/null || true
fi

# --- base_missing on a tiny repo --------------------------------------------

MINI="$WORK/mini"
MINI_PATCHES="$WORK/mini-patches"
mkdir -p "$MINI" "$MINI_PATCHES"
git -C "$MINI" init -q
printf 'alpha\n' > "$MINI/a.txt"
git -C "$MINI" add a.txt
git -C "$MINI" -c user.email=ci@corecheck.dev -c user.name=corecheck -c commit.gpgsign=false commit -q -m base
BASE_SHA=$(git -C "$MINI" rev-parse HEAD)
printf 'alpha\nbeta\n' > "$MINI/a.txt"
printf 'one\n' > "$MINI/c.txt"
git -C "$MINI" add a.txt c.txt
git -C "$MINI" -c user.email=ci@corecheck.dev -c user.name=corecheck -c commit.gpgsign=false commit -q -m head

printf 'alpha\nbeta\ngamma\n' > "$MINI/a.txt"
git -C "$MINI" add a.txt
git -C "$MINI" --no-pager diff --cached > "$MINI_PATCHES/only-head.patch"
git -C "$MINI" reset -q
git -C "$MINI" checkout -q -- a.txt

printf 'one\ntwo\n' > "$MINI/c.txt"
git -C "$MINI" add c.txt
git -C "$MINI" --no-pager diff --cached > "$MINI_PATCHES/c-head.patch"
git -C "$MINI" reset -q
git -C "$MINI" checkout -q -- c.txt

printf 'both\n' > "$MINI/b.txt"
git -C "$MINI" add b.txt
git -C "$MINI" --no-pager diff --cached > "$MINI_PATCHES/both.patch"
git -C "$MINI" reset -q
rm -f "$MINI/b.txt"

(
    set -euo pipefail
    cd "$MINI"
    # shellcheck disable=SC1090
    source "$LIB"
    TEST_PATCH_APPLIED=("both.patch" "only-head.patch" "c-head.patch")
    TEST_PATCH_FAILED_NAME=()
    TEST_PATCH_FAILED_REASON=()
    TEST_PATCH_SET_VERSION="determinism-2026-11-12"
    TEST_PATCH_BITCOIN_COMMIT="a06515e9fdf81a762804471733b1f43a13e1f10b"
    TEST_PATCH_DIR="$MINI_PATCHES"
    IS_MASTER=false
    BASE_COMMIT="$BASE_SHA"
    COVERAGE_DIFF_FILE=""
    write_coverage_pr_diff
    printf '%s\n' "${TEST_PATCH_BASE_MISSING[@]}" > "$WORK/base-missing.txt"
    cp "$COVERAGE_DIFF_FILE" "$WORK/pr.diff"
    write_tests_patches_record "$WORK/diff-record.json" "prhead" "false"
    git status --porcelain > "$WORK/mini-status.txt"
    git worktree list > "$WORK/mini-worktrees.txt"
) > "$WORK/diff.stdout" 2> "$WORK/diff.stderr"
diff_status=$?

if [ "$diff_status" -eq 0 ]; then
    pass "write_coverage_pr_diff exits 0 under set -e"
else
    fail "write_coverage_pr_diff exits 0 under set -e"
    echo "--- diff stdout ---"; cat "$WORK/diff.stdout" 2>/dev/null || true
    echo "--- diff stderr ---"; cat "$WORK/diff.stderr" 2>/dev/null || true
fi

warn_n=$(grep -c '^WARNING: applied on the pull request but not on ' "$WORK/diff.stdout" || true)
if [ "$warn_n" -eq 1 ] \
    && grep -q 'only-head.patch' "$WORK/diff.stdout" \
    && grep -q 'c-head.patch' "$WORK/diff.stdout" \
    && ! grep -q 'COVERAGE_TEST_PATCH_MISSING:' "$WORK/diff.stdout"; then
    pass "one unprefixed warning names both base-missing patches"
else
    fail "one unprefixed warning names both base-missing patches"
    cat "$WORK/diff.stdout" 2>/dev/null || true
fi

if cmp -s "$WORK/base-missing.txt" <(printf '%s\n' "only-head.patch" "c-head.patch") \
    && jq -e '.base_missing == ["only-head.patch","c-head.patch"] and .is_master == false' "$WORK/diff-record.json" >/dev/null \
    && grep -q '^+gamma' "$WORK/pr.diff" \
    && grep -q '^+two' "$WORK/pr.diff" \
    && ! grep -q '^+both' "$WORK/pr.diff"; then
    pass "base_missing diff keeps the leaked patch and drops the one that applied on both"
else
    fail "base_missing diff keeps the leaked patch and drops the one that applied on both"
    echo "--- base_missing ---"; cat "$WORK/base-missing.txt" 2>/dev/null || true
    echo "--- record ---"; cat "$WORK/diff-record.json" 2>/dev/null || true
    echo "--- diff ---"; cat "$WORK/pr.diff" 2>/dev/null || true
fi

if [ ! -s "$WORK/mini-status.txt" ] && ! grep -q 'coverage-base' "$WORK/mini-worktrees.txt"; then
    pass "diff leaves the pull-request index clean and removes the base worktree"
else
    fail "diff leaves the pull-request index clean and removes the base worktree"
    cat "$WORK/mini-status.txt" "$WORK/mini-worktrees.txt" 2>/dev/null || true
fi

(
    set -euo pipefail
    cd "$MINI"
    # shellcheck disable=SC1090
    source "$LIB"
    IS_MASTER=true
    BASE_COMMIT="$BASE_SHA"
    before=$(git worktree list | wc -l)
    write_coverage_pr_diff
    after=$(git worktree list | wc -l)
    [ "$before" = "$after" ]
)
if [ "$?" -eq 0 ]; then
    pass "master run does not create a base worktree"
else
    fail "master run does not create a base worktree"
fi

(
    set -euo pipefail
    cd "$MINI"
    # shellcheck disable=SC1090
    source "$LIB"
    IS_MASTER=false
    BASE_COMMIT="0123456789abcdef0123456789abcdef01234567"
    TEST_PATCH_APPLIED=()
    TEST_PATCH_DIR="$MINI_PATCHES"
    COVERAGE_DIFF_FILE=""
    write_coverage_pr_diff
    [ -z "$COVERAGE_DIFF_FILE" ]
) > "$WORK/bad-base.stdout" 2> "$WORK/bad-base.stderr"
if [ "$?" -eq 0 ] && grep -q 'could not check out' "$WORK/bad-base.stderr"; then
    pass "missing base commit warns and leaves the diff empty"
else
    fail "missing base commit warns and leaves the diff empty"
    cat "$WORK/bad-base.stdout" "$WORK/bad-base.stderr" 2>/dev/null || true
fi

# --- bitcoin clone: three apply passes --------------------------------------

TARGET=$(jq -r '.bitcoin_commit' "$PATCHES/manifest.json")
VERSION=$(jq -r '.set_version' "$PATCHES/manifest.json")
HELPER=$(jq -r '.helper_patch' "$PATCHES/manifest.json")
SHORT=${TARGET:0:8}
mapfile -t ORDERED < <(jq -r --arg h "$HELPER" '.patches | sort_by(.order) | .[] | select(.name != $h) | .name' "$PATCHES/manifest.json")
mapfile -t NEEDS_HELPER < <(jq -r --arg h "$HELPER" '
    .patches | sort_by(.order) | .[]
    | select(.name != $h)
    | select(any(.requires[]?; . == $h))
    | .name
' "$PATCHES/manifest.json")

if ! git -C "$BITCOIN_WT" cat-file -e "${TARGET}^{commit}"; then
    fail "manifest commit is not in BITCOIN_WT"
    echo "FAILED: $FAILS"
    exit 1
fi
pass "manifest commit is in BITCOIN_WT"

CLONE="$WORK/bitcoin"
echo "cloning $(basename "$BITCOIN_WT") at ${SHORT}"
if ! git clone -q --shared "$BITCOIN_WT" "$CLONE"; then
    fail "git clone --shared"
    echo "FAILED: $FAILS"
    exit 1
fi
if ! git -C "$CLONE" checkout -q "$TARGET"; then
    fail "checkout manifest commit"
    echo "FAILED: $FAILS"
    exit 1
fi
pass "throwaway clone is at ${TARGET}"

reset_clone() {
    git -C "$CLONE" reset --hard -q HEAD
    git -C "$CLONE" clean -fdq
}

run_apply() {
    local tag="$1"
    local patches="$2"
    (
        set -euo pipefail
        cd "$CLONE"
        # shellcheck disable=SC1090
        source "$LIB"
        apply_test_patches "$patches"
        if [ "${#TEST_PATCH_APPLIED[@]}" -eq 0 ]; then
            : > "$WORK/${tag}.applied"
        else
            printf '%s\n' "${TEST_PATCH_APPLIED[@]}" > "$WORK/${tag}.applied"
        fi
        if [ "${#TEST_PATCH_FAILED_NAME[@]}" -eq 0 ]; then
            : > "$WORK/${tag}.failed-name"
        else
            printf '%s\n' "${TEST_PATCH_FAILED_NAME[@]}" > "$WORK/${tag}.failed-name"
        fi
        if [ "${#TEST_PATCH_FAILED_REASON[@]}" -eq 0 ]; then
            : > "$WORK/${tag}.failed-reason"
        else
            printf '%s\n' "${TEST_PATCH_FAILED_REASON[@]}" > "$WORK/${tag}.failed-reason"
        fi
    ) > "$WORK/${tag}.stdout" 2> "$WORK/${tag}.stderr"
    printf '%s' "$?" > "$WORK/${tag}.status"
}

summary_for() {
    local applied_n="$1"
    local failed_n="$2"
    local total=$((applied_n + failed_n))
    echo "Test patch set ${VERSION} for bitcoin ${SHORT}: applied ${applied_n}/${total}, missing ${failed_n}"
}

names_except() {
    local skip="$1"
    local name
    for name in "$HELPER" "${ORDERED[@]}"; do
        if [ "$name" != "$skip" ]; then
            printf '%s\n' "$name"
        fi
    done
}

CORRUPT="$WORK/corrupt-patches"
cp -a "$PATCHES" "$CORRUPT"
# Corrupt context (a leading-space line), not an added line, so apply fails.
sed -i '/^[^-+@]/s/out_bytes/out_bytes_CORRUPT/g' "$CORRUPT/14-crypto-sha3-rate-boundary.patch"

reset_clone
run_apply p1 "$CORRUPT"
echo "----- pass 1 stdout -----"
cat "$WORK/p1.stdout"

p1_status=$(cat "$WORK/p1.status")
if [ "$p1_status" = "0" ]; then
    pass "corrupt 14 exits 0"
else
    fail "corrupt 14 exits 0 (status ${p1_status})"
    cat "$WORK/p1.stderr" || true
fi

p1_line="COVERAGE_TEST_PATCH_MISSING: 14-crypto-sha3-rate-boundary.patch (does not apply)"
p1_n=$(prefix_count "$WORK/p1.stdout")
if [ "$p1_n" = "1" ] && grep -q -x -F "$p1_line" "$WORK/p1.stdout"; then
    pass "corrupt 14 prints exactly one missing line"
else
    fail "corrupt 14 prints exactly one missing line (count ${p1_n})"
fi

if grep -q -x -F "$(summary_for ${#ORDERED[@]} 1)" "$WORK/p1.stdout"; then
    pass "corrupt 14 summary is applied ${#ORDERED[@]}/$((${#ORDERED[@]} + 1)), missing 1"
else
    fail "corrupt 14 summary is applied ${#ORDERED[@]}/$((${#ORDERED[@]} + 1)), missing 1"
fi

names_except "14-crypto-sha3-rate-boundary.patch" > "$WORK/p1.want-applied"
if cmp -s "$WORK/p1.want-applied" "$WORK/p1.applied" \
    && cmp -s <(printf '%s\n' "14-crypto-sha3-rate-boundary.patch") "$WORK/p1.failed-name" \
    && cmp -s <(printf '%s\n' "does not apply") "$WORK/p1.failed-reason"; then
    pass "corrupt 14 applied list is every other patch"
else
    fail "corrupt 14 applied list is every other patch"
    echo "--- applied ---"; cat "$WORK/p1.applied" || true
    echo "--- failed ---"; cat "$WORK/p1.failed-name" || true
fi

git -C "$CLONE" diff --name-only | sort > "$WORK/p1.names"
jq -r --arg skip "src/test/crypto_tests.cpp" \
    '[.patches[].target_file] | unique | map(select(. != $skip)) | .[]' \
    "$PATCHES/manifest.json" | sort > "$WORK/p1.want-names"
if cmp -s "$WORK/p1.want-names" "$WORK/p1.names" \
    && git -C "$CLONE" diff --quiet -- src/test/crypto_tests.cpp; then
    pass "corrupt 14 leaves crypto_tests.cpp untouched and modifies the other tests"
else
    fail "corrupt 14 leaves crypto_tests.cpp untouched and modifies the other tests"
    echo "--- names ---"; cat "$WORK/p1.names" || true
    echo "--- want ---"; cat "$WORK/p1.want-names" || true
fi

reset_clone
run_apply p2 "$PATCHES"
echo "----- pass 2 stdout -----"
cat "$WORK/p2.stdout"

p2_status=$(cat "$WORK/p2.status")
p2_n=$(prefix_count "$WORK/p2.stdout")
if [ "$p2_status" = "0" ] && [ "$p2_n" = "0" ]; then
    pass "pristine set exits 0 with no missing line"
else
    fail "pristine set exits 0 with no missing line (status ${p2_status}, missing ${p2_n})"
    cat "$WORK/p2.stderr" || true
fi

if grep -q -x -F "$(summary_for $((${#ORDERED[@]} + 1)) 0)" "$WORK/p2.stdout"; then
    pass "pristine summary is applied $((${#ORDERED[@]} + 1))/$((${#ORDERED[@]} + 1)), missing 0"
else
    fail "pristine summary is applied $((${#ORDERED[@]} + 1))/$((${#ORDERED[@]} + 1)), missing 0"
fi

names_except "" > "$WORK/p2.want-applied"
if cmp -s "$WORK/p2.want-applied" "$WORK/p2.applied" \
    && [ ! -s "$WORK/p2.failed-name" ]; then
    pass "pristine set applies all $((${#ORDERED[@]} + 1)) patches"
else
    fail "pristine set applies all $((${#ORDERED[@]} + 1)) patches"
    echo "--- applied ---"; cat "$WORK/p2.applied" || true
    echo "--- want ---"; cat "$WORK/p2.want-applied" || true
fi

git -C "$CLONE" diff --name-only | sort > "$WORK/p2.names"
jq -r '[.patches[].target_file] | unique | .[]' "$PATCHES/manifest.json" | sort > "$WORK/p2.want-names"
if cmp -s "$WORK/p2.want-names" "$WORK/p2.names"; then
    pass "pristine set modifies every targeted test file"
else
    fail "pristine set modifies every targeted test file"
    echo "--- names ---"; cat "$WORK/p2.names" || true
    echo "--- want ---"; cat "$WORK/p2.want-names" || true
fi

NOHELP="$WORK/no-helper-patches"
cp -a "$PATCHES" "$NOHELP"
rm -f "$NOHELP/00-test-helpers.patch"

reset_clone
run_apply p3 "$NOHELP"
echo "----- pass 3 stdout -----"
cat "$WORK/p3.stdout"

p3_status=$(cat "$WORK/p3.status")
p3_n=$(prefix_count "$WORK/p3.stdout")
p3_ok=1
if [ "$p3_status" != "0" ] || [ "$p3_n" != "${#NEEDS_HELPER[@]}" ]; then
    p3_ok=0
fi
for name in "${NEEDS_HELPER[@]}"; do
    line="COVERAGE_TEST_PATCH_MISSING: ${name} (depends on ${HELPER})"
    if ! grep -q -x -F "$line" "$WORK/p3.stdout"; then
        p3_ok=0
    fi
done
if [ "$p3_ok" -eq 1 ]; then
    pass "missing helper prints exactly the dependency lines"
else
    fail "missing helper prints exactly the dependency lines (status ${p3_status}, count ${p3_n}, want ${#NEEDS_HELPER[@]})"
    cat "$WORK/p3.stderr" || true
fi

p3_applied=$((${#ORDERED[@]} - ${#NEEDS_HELPER[@]}))
if grep -q -x -F "$(summary_for "$p3_applied" "${#NEEDS_HELPER[@]}")" "$WORK/p3.stdout"; then
    pass "missing helper summary is applied ${p3_applied}/$((p3_applied + ${#NEEDS_HELPER[@]})), missing ${#NEEDS_HELPER[@]}"
else
    fail "missing helper summary is applied ${p3_applied}/$((p3_applied + ${#NEEDS_HELPER[@]})), missing ${#NEEDS_HELPER[@]}"
fi

{
    for name in "${ORDERED[@]}"; do
        skip=0
        for need in "${NEEDS_HELPER[@]}"; do
            if [ "$name" = "$need" ]; then
                skip=1
            fi
        done
        if [ "$skip" -eq 0 ]; then
            printf '%s\n' "$name"
        fi
    done
} > "$WORK/p3.want-applied"
printf '%s\n' "${NEEDS_HELPER[@]}" > "$WORK/p3.want-failed"
: > "$WORK/p3.want-reason"
for _name in "${NEEDS_HELPER[@]}"; do
    printf '%s\n' "depends on ${HELPER}"
done > "$WORK/p3.want-reason"
if cmp -s "$WORK/p3.want-applied" "$WORK/p3.applied" \
    && cmp -s "$WORK/p3.want-failed" "$WORK/p3.failed-name" \
    && cmp -s "$WORK/p3.want-reason" "$WORK/p3.failed-reason" \
    && ! grep -q -x -F "00-test-helpers.patch" "$WORK/p3.applied" \
    && ! grep -q -x -F "00-test-helpers.patch" "$WORK/p3.failed-name"; then
    pass "missing helper: only the patches that require it are missing"
else
    fail "missing helper: only the patches that require it are missing"
    echo "--- applied ---"; cat "$WORK/p3.applied" || true
    echo "--- failed ---"; cat "$WORK/p3.failed-name" || true
    echo "--- reasons ---"; cat "$WORK/p3.failed-reason" || true
fi

if ! grep -q 'StubMsgProc' "$CLONE/src/test/net_tests.cpp" \
    && ! grep -q 'message_handler_returns_when_interrupted_during_send' "$CLONE/src/test/net_tests.cpp" \
    && ! grep -q 'stop_nodes_truncates_block_relay_anchors' "$CLONE/src/test/net_tests.cpp" \
    && ! grep -q 'stop_nodes_deletes_nodes_still_referenced_elsewhere' "$CLONE/src/test/net_tests.cpp" \
    && ! grep -q 'open_connections_drives_slot_and_address_selection' "$CLONE/src/test/net_tests.cpp" \
    && grep -q 'cached_invalid_block_discourages_outbound_peer' "$CLONE/src/test/denialofservice_tests.cpp" \
    && grep -q 'socket_handler_skips_node_without_socket' "$CLONE/src/test/net_tests.cpp" \
    && grep -q 'sha3_256_chunked_write_at_rate_boundary' "$CLONE/src/test/crypto_tests.cpp"; then
    pass "missing helper leaves helper dependents out and still applies patch 40"
else
    fail "missing helper leaves helper dependents out and still applies patch 40"
fi

# 22 and 41 both list 21 in requires. 42 lists only the helper. Deleting 21
# must skip exactly the patches that name it, and still apply the rest.
DEP21="21-net-message-handler-interrupt-in-send.patch"
mapfile -t REQUIRES_21 < <(jq -r --arg dep "$DEP21" --arg h "$HELPER" '
    .patches | sort_by(.order) | .[]
    | select(.name != $h)
    | select(.name != $dep)
    | select(any(.requires[]?; . == $dep))
    | .name
' "$PATCHES/manifest.json")

NO21="$WORK/no-21-patches"
cp -a "$PATCHES" "$NO21"
rm -f "$NO21/$DEP21"

reset_clone
run_apply p4 "$NO21"
echo "----- pass 4 stdout -----"
cat "$WORK/p4.stdout"

p4_status=$(cat "$WORK/p4.status")
p4_n=$(prefix_count "$WORK/p4.stdout")
p4_expect_n=$((1 + ${#REQUIRES_21[@]}))
p4_ok=1
if [ "$p4_status" != "0" ] || [ "$p4_n" != "$p4_expect_n" ]; then
    p4_ok=0
fi
p4_21_line="COVERAGE_TEST_PATCH_MISSING: ${DEP21} (does not apply)"
if ! grep -q -x -F "$p4_21_line" "$WORK/p4.stdout"; then
    p4_ok=0
fi
p4_41_line="COVERAGE_TEST_PATCH_MISSING: 41-net-stop-nodes-referenced-node.patch (depends on ${DEP21})"
if ! grep -q -x -F "$p4_41_line" "$WORK/p4.stdout"; then
    p4_ok=0
fi
for name in "${REQUIRES_21[@]}"; do
    line="COVERAGE_TEST_PATCH_MISSING: ${name} (depends on ${DEP21})"
    if ! grep -q -x -F "$line" "$WORK/p4.stdout"; then
        p4_ok=0
    fi
done
if [ "$p4_ok" -eq 1 ]; then
    pass "missing 21: 41 depends on 21, and the deleted file is does not apply"
else
    fail "missing 21: 41 depends on 21, and the deleted file is does not apply (status ${p4_status}, count ${p4_n}, want ${p4_expect_n})"
    cat "$WORK/p4.stderr" || true
fi

p4_failed=$((1 + ${#REQUIRES_21[@]}))
p4_applied=$((${#ORDERED[@]} - ${#REQUIRES_21[@]}))
if grep -q -x -F "$(summary_for "$p4_applied" "$p4_failed")" "$WORK/p4.stdout"; then
    pass "missing 21 summary is applied ${p4_applied}/$((p4_applied + p4_failed)), missing ${p4_failed}"
else
    fail "missing 21 summary is applied ${p4_applied}/$((p4_applied + p4_failed)), missing ${p4_failed}"
fi

{
    printf '%s\n' "$HELPER"
    for name in "${ORDERED[@]}"; do
        if [ "$name" = "$DEP21" ]; then
            continue
        fi
        skip=0
        for dep in "${REQUIRES_21[@]}"; do
            if [ "$name" = "$dep" ]; then
                skip=1
            fi
        done
        if [ "$skip" -eq 0 ]; then
            printf '%s\n' "$name"
        fi
    done
} > "$WORK/p4.want-applied"
{
    printf '%s\n' "$DEP21"
    if [ "${#REQUIRES_21[@]}" -gt 0 ]; then
        printf '%s\n' "${REQUIRES_21[@]}"
    fi
} > "$WORK/p4.want-failed"
{
    printf '%s\n' "does not apply"
    for _name in "${REQUIRES_21[@]}"; do
        printf '%s\n' "depends on ${DEP21}"
    done
} > "$WORK/p4.want-reason"
if cmp -s "$WORK/p4.want-applied" "$WORK/p4.applied" \
    && cmp -s "$WORK/p4.want-failed" "$WORK/p4.failed-name" \
    && cmp -s "$WORK/p4.want-reason" "$WORK/p4.failed-reason" \
    && grep -q -x -F "42-net-slot-and-address-selection.patch" "$WORK/p4.applied" \
    && grep -q -x -F "40-dos-cached-invalid-block.patch" "$WORK/p4.applied" \
    && ! grep -q 'depends on 00-test-helpers.patch' "$WORK/p4.stdout"; then
    pass "removing 21 skips only patches that require it; 40 and 42 still apply"
else
    fail "removing 21 skips only patches that require it; 40 and 42 still apply"
    echo "--- applied ---"; cat "$WORK/p4.applied" || true
    echo "--- want applied ---"; cat "$WORK/p4.want-applied" || true
    echo "--- failed ---"; cat "$WORK/p4.failed-name" || true
    echo "--- want failed ---"; cat "$WORK/p4.want-failed" || true
    echo "--- reasons ---"; cat "$WORK/p4.failed-reason" || true
fi

if grep -q 'StubMsgProc' "$CLONE/src/test/net_tests.cpp" \
    && ! grep -q 'message_handler_returns_when_interrupted_during_send' "$CLONE/src/test/net_tests.cpp" \
    && ! grep -q 'stop_nodes_truncates_block_relay_anchors' "$CLONE/src/test/net_tests.cpp" \
    && ! grep -q 'stop_nodes_deletes_nodes_still_referenced_elsewhere' "$CLONE/src/test/net_tests.cpp" \
    && grep -q 'open_connections_drives_slot_and_address_selection' "$CLONE/src/test/net_tests.cpp" \
    && grep -q 'cached_invalid_block_discourages_outbound_peer' "$CLONE/src/test/denialofservice_tests.cpp" \
    && grep -q 'socket_handler_skips_node_without_socket' "$CLONE/src/test/net_tests.cpp" \
    && grep -q 'sha3_256_chunked_write_at_rate_boundary' "$CLONE/src/test/crypto_tests.cpp"; then
    pass "missing 21 leaves 21, 22, and 41 out and still applies patch 42"
else
    fail "missing 21 leaves 21, 22, and 41 out and still applies patch 42"
fi

echo
if [ "$FAILS" -eq 0 ]; then
    echo "ALL PASS"
    exit 0
fi
echo "FAILED: $FAILS"
exit 1
