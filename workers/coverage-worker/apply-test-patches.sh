#!/usr/bin/env bash

# Apply the per-test coverage patch set and build the pull-request diff
# that excludes patches which landed on both trees.
#
# Nothing in this file runs at source time. Call it from the bitcoin
# checkout root (the image uses /tmp/bitcoin).
#
# apply_test_patches [patches_dir]
#   patches_dir defaults to $TESTS_PATCHES_DIR, then /tests-patches.
#   Reads manifest.json there. Applies helper_patch first when the manifest
#   names one and that file is present, then patches[] in order. A patch
#   that does not apply is reported and skipped. The function returns 0
#   even when every patch failed, including under set -e.
#
#   A helper named by the manifest but missing from the directory is not
#   itself a failed patch. A patch whose requires list names a patch that
#   was not applied is not attempted; the reason is "depends on <that
#   patch>". When several requirements are unmet, the reason names the one
#   that comes first in manifest order (.order), not the order written in
#   the requires array. requires defaults to []. An older manifest that has
#   needs_helper: true and no requires array is treated as
#   requires: [helper_patch]. If the manifest has no helper_patch, that
#   fallback adds nothing, so a patch is not skipped for a missing helper.
#   The helper entry in patches[] is skipped here because helper_patch was
#   already applied above. A helper file that is present but does not apply
#   is reported as "does not apply", and anything that requires it is still
#   not attempted.
#
#   After return, the caller can read:
#     TEST_PATCH_APPLIED         files that applied, helper first
#     TEST_PATCH_FAILED_NAME     files that did not apply
#     TEST_PATCH_FAILED_REASON   parallel reasons: "does not apply" or
#                                "depends on <patch name>"
#     TEST_PATCH_SET_VERSION     manifest set_version
#     TEST_PATCH_BITCOIN_COMMIT  manifest bitcoin_commit
#     TEST_PATCH_DIR             absolute patches directory
#     TEST_PATCH_BASE_MISSING    cleared here; write_coverage_pr_diff fills it
#
# write_coverage_pr_diff
#   Pull-request runs only. Master ($IS_MASTER = true) returns immediately
#   and does not create a worktree. For the patches in TEST_PATCH_APPLIED:
#   git apply --cached, git write-tree, git reset. Then a detached worktree
#   at $BASE_COMMIT, stage the same patches there, and git diff the two
#   trees into $COVERAGE_DIFF_FILE. A patch that applied on the pull request
#   but not on the base is appended to TEST_PATCH_BASE_MISSING, named in one
#   warning line, and left in the diff. If the base worktree cannot be
#   created, COVERAGE_DIFF_FILE stays empty.
#
# write_tests_patches_record <output> <head_commit> <is_master>
#   Writes the JSON record. is_master is the string true or false.
#   head_commit is $COMMIT on master and $HEAD_COMMIT on a pull request.
#   base_missing is whatever TEST_PATCH_BASE_MISSING holds (empty on master).

# Internal. Prints the one missing-patch line and records the failure.
record_test_patch_miss() {
    local name="$1"
    local reason="$2"
    echo "COVERAGE_TEST_PATCH_MISSING: ${name} (${reason})"
    TEST_PATCH_FAILED_NAME+=("$name")
    TEST_PATCH_FAILED_REASON+=("$reason")
    return 0
}

# Internal. Sets _APPLY_ONE_OK to 1 on success, 0 on failure. Always returns 0.
apply_one_patch() {
    local name="$1"
    local path="${TEST_PATCH_DIR}/${name}"
    _APPLY_ONE_OK=0
    if [ ! -f "$path" ]; then
        record_test_patch_miss "$name" "does not apply"
        return 0
    fi
    if ! git apply --check "$path"; then
        record_test_patch_miss "$name" "does not apply"
        return 0
    fi
    if ! git apply "$path"; then
        record_test_patch_miss "$name" "does not apply"
        return 0
    fi
    TEST_PATCH_APPLIED+=("$name")
    _APPLY_ONE_OK=1
    return 0
}

apply_test_patches() {
    local patches_dir="${1:-${TESTS_PATCHES_DIR:-/tests-patches}}"
    local manifest helper helper_applied patch_rows name requires_csv
    local missing req oldifs applied_n failed_n total short
    local -a reqs=()
    local -A applied_set=()

    TEST_PATCH_APPLIED=()
    TEST_PATCH_FAILED_NAME=()
    TEST_PATCH_FAILED_REASON=()
    TEST_PATCH_BASE_MISSING=()
    TEST_PATCH_SET_VERSION=""
    TEST_PATCH_BITCOIN_COMMIT=""
    TEST_PATCH_DIR=""
    _APPLY_ONE_OK=0

    if [ ! -d "$patches_dir" ]; then
        echo "Test patch set unknown for bitcoin unknown: applied 0/0, missing 0"
        return 0
    fi
    patches_dir=$(cd "$patches_dir" && pwd)
    TEST_PATCH_DIR="$patches_dir"
    manifest="${patches_dir}/manifest.json"
    if [ ! -f "$manifest" ]; then
        echo "Test patch set unknown for bitcoin unknown: applied 0/0, missing 0"
        return 0
    fi

    TEST_PATCH_SET_VERSION=$(jq -r '.set_version // empty' "$manifest")
    TEST_PATCH_BITCOIN_COMMIT=$(jq -r '.bitcoin_commit // empty' "$manifest")
    helper=$(jq -r '.helper_patch // empty' "$manifest")

    # No helper in the manifest: dependents are ordinary patches.
    helper_applied=1
    if [ -n "$helper" ]; then
        helper_applied=0
        if [ -f "${patches_dir}/${helper}" ]; then
            apply_one_patch "$helper"
            helper_applied=$_APPLY_ONE_OK
        fi
    fi

    if [ -n "$helper" ] && [ "$helper_applied" -eq 1 ]; then
        applied_set["$helper"]=1
    fi

    # Skip the helper row: helper_patch was applied above. requires is the
    # explicit prerequisite list (manifest v2). Missing requires plus
    # needs_helper: true is the v1 fallback, and only when helper_patch is
    # set. req_order keeps order 0 (the helper) ahead of later patches;
    # `//` would treat 0 as empty and sort it last.
    patch_rows=$(jq -r --arg helper "$helper" '
        def req_order($ord):
            . as $name
            | if ($ord | has($name)) then $ord[$name] else 1000000000 end;
        .patches as $all
        | ($all | map({key: .name, value: .order}) | from_entries) as $ord
        | $all
        | sort_by(.order)
        | .[]
        | select(.name != $helper)
        | .name as $name
        | (
            (.requires // (if .needs_helper == true and $helper != "" then [$helper] else [] end))
            | sort_by(req_order($ord))
            | join(",")
          ) as $reqs
        | [$name, $reqs]
        | @tsv
    ' "$manifest")

    while IFS=$'\t' read -r name requires_csv; do
        if [ -z "$name" ]; then
            continue
        fi
        missing=""
        # requires_csv is sorted by manifest .order. The first name that
        # was not applied is the reason.
        if [ -n "$requires_csv" ]; then
            oldifs=$IFS
            IFS=','
            read -ra reqs <<< "$requires_csv"
            IFS=$oldifs
            for req in "${reqs[@]}"; do
                if [ -z "${applied_set[$req]:-}" ]; then
                    missing=$req
                    break
                fi
            done
        fi
        if [ -n "$missing" ]; then
            record_test_patch_miss "$name" "depends on ${missing}"
            continue
        fi
        apply_one_patch "$name"
        if [ "$_APPLY_ONE_OK" -eq 1 ]; then
            applied_set["$name"]=1
        fi
    done <<< "$patch_rows"

    applied_n=${#TEST_PATCH_APPLIED[@]}
    failed_n=${#TEST_PATCH_FAILED_NAME[@]}
    total=$((applied_n + failed_n))
    short=${TEST_PATCH_BITCOIN_COMMIT:0:8}
    echo "Test patch set ${TEST_PATCH_SET_VERSION} for bitcoin ${short}: applied ${applied_n}/${total}, missing ${failed_n}"
    return 0
}

# Remove a directory this process created with mktemp. Refuse anything else.
remove_owned_tempdir() {
    local dir="$1"
    if [ -z "$dir" ] || [ ! -d "$dir" ]; then
        return 0
    fi
    case "$dir" in
        "${TMPDIR:-/tmp}"/coverage-base.*) ;;
        *) return 0 ;;
    esac
    git worktree remove --force "$dir" >/dev/null 2>&1 || rm -rf "$dir" || true
    return 0
}

write_coverage_pr_diff() {
    local patches_dir name patch pr_tree base_tree base_dir

    if [ "${IS_MASTER:-}" = "true" ]; then
        return 0
    fi

    TEST_PATCH_BASE_MISSING=()
    patches_dir="${TEST_PATCH_DIR:-${TESTS_PATCHES_DIR:-/tests-patches}}"

    if [ "${#TEST_PATCH_APPLIED[@]}" -gt 0 ]; then
        for name in "${TEST_PATCH_APPLIED[@]}"; do
            patch="${patches_dir}/${name}"
            if ! git apply --cached "$patch"; then
                echo "WARNING: could not stage ${name}; the stored diff will be the unpatched pull request diff" >&2
                git reset -q || true
                return 0
            fi
        done
    fi

    if ! pr_tree=$(git write-tree); then
        echo "WARNING: could not write the patched pull request tree; the stored diff will be the unpatched pull request diff" >&2
        git reset -q || true
        return 0
    fi
    git reset -q || true

    base_dir=$(mktemp -d "${TMPDIR:-/tmp}/coverage-base.XXXXXX")
    if ! git worktree add --detach "$base_dir" "$BASE_COMMIT"; then
        echo "WARNING: could not check out ${BASE_COMMIT} to build the coverage diff" >&2
        remove_owned_tempdir "$base_dir"
        return 0
    fi

    if [ "${#TEST_PATCH_APPLIED[@]}" -gt 0 ]; then
        for name in "${TEST_PATCH_APPLIED[@]}"; do
            patch="${patches_dir}/${name}"
            if ! git -C "$base_dir" apply --check "$patch" || ! git -C "$base_dir" apply --cached "$patch"; then
                TEST_PATCH_BASE_MISSING+=("$name")
            fi
        done
    fi

    if ! base_tree=$(git -C "$base_dir" write-tree); then
        echo "WARNING: could not write the base tree; the stored diff will be the unpatched pull request diff" >&2
        remove_owned_tempdir "$base_dir"
        return 0
    fi

    COVERAGE_DIFF_FILE=$(mktemp)
    if ! git diff "$base_tree" "$pr_tree" > "$COVERAGE_DIFF_FILE"; then
        echo "WARNING: could not diff the patched trees; the stored diff will be the unpatched pull request diff" >&2
        rm -f "$COVERAGE_DIFF_FILE"
        COVERAGE_DIFF_FILE=""
        remove_owned_tempdir "$base_dir"
        return 0
    fi

    if [ "${#TEST_PATCH_BASE_MISSING[@]}" -gt 0 ]; then
        echo "WARNING: applied on the pull request but not on ${BASE_COMMIT}; the stored diff may include: ${TEST_PATCH_BASE_MISSING[*]}"
    fi
    echo "Recorded coverage diff between the patched base and the patched pull request"
    remove_owned_tempdir "$base_dir"
    return 0
}

# Encode "$@" as a JSON array of strings. No arguments yields [].
json_string_array() {
    local result="["
    local sep=""
    local item encoded
    if [ "$#" -eq 0 ]; then
        printf '%s' "[]"
        return 0
    fi
    for item in "$@"; do
        encoded=$(jq -nc --arg s "$item" '$s')
        result="${result}${sep}${encoded}"
        sep=","
    done
    result="${result}]"
    printf '%s' "$result"
    return 0
}

write_tests_patches_record() {
    local out="$1"
    local head_commit="$2"
    local is_master="$3"
    local applied_json failed_json base_missing_json
    local is_master_json=false
    local generated_at item sep i

    applied_json=$(json_string_array "${TEST_PATCH_APPLIED[@]}")
    base_missing_json=$(json_string_array "${TEST_PATCH_BASE_MISSING[@]}")

    failed_json="["
    sep=""
    if [ "${#TEST_PATCH_FAILED_NAME[@]}" -gt 0 ]; then
        for i in "${!TEST_PATCH_FAILED_NAME[@]}"; do
            item=$(jq -nc \
                --arg patch "${TEST_PATCH_FAILED_NAME[$i]}" \
                --arg reason "${TEST_PATCH_FAILED_REASON[$i]}" \
                '{patch: $patch, reason: $reason}')
            failed_json="${failed_json}${sep}${item}"
            sep=","
        done
    fi
    failed_json="${failed_json}]"

    if [ "$is_master" = "true" ]; then
        is_master_json=true
    fi
    generated_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)

    jq -n \
        --arg set_version "${TEST_PATCH_SET_VERSION:-}" \
        --arg bitcoin_commit "${TEST_PATCH_BITCOIN_COMMIT:-}" \
        --arg head_commit "$head_commit" \
        --argjson is_master "$is_master_json" \
        --arg generated_at "$generated_at" \
        --argjson applied "$applied_json" \
        --argjson failed "$failed_json" \
        --argjson base_missing "$base_missing_json" \
        '{
            set_version: $set_version,
            bitcoin_commit: $bitcoin_commit,
            head_commit: $head_commit,
            is_master: $is_master,
            generated_at: $generated_at,
            applied: $applied,
            failed: $failed,
            base_missing: $base_missing
        }' > "$out"
    return 0
}
