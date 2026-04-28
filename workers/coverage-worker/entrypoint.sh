#!/bin/bash -x

COMMIT=$1
PR_NUM=$2
IS_MASTER=$3
BASE_COMMIT=$4

set -e

TELEMETRY_TIMESTREAM_TABLE=${TELEMETRY_TIMESTREAM_TABLE:-dashboard_metrics}
telemetry_records=()

telemetry_ready() {
    [ "${TELEMETRY_BACKEND}" = "timestream" ] &&
        [ -n "${TELEMETRY_TIMESTREAM_DATABASE}" ] &&
        [ -n "${TELEMETRY_TIMESTREAM_REGION}" ]
}

sanitize_dimension_name() {
    local sanitized
    sanitized=$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^[:alnum:]]+/_/g; s/^_+//; s/_+$//')
    if [ -z "$sanitized" ]; then
        sanitized="tag"
    fi
    printf '%s' "$sanitized"
}

queue_metric() {
    if ! telemetry_ready; then
        return 0
    fi

    local metric_name=$1
    local metric_value=$2
    shift 2

    local dimensions='[]'
    dimensions=$(printf '%s' "$dimensions" | jq -c --arg value "$metric_name" '. + [{"Name":"metric_name","Value":$value}]')

    while [ "$#" -gt 0 ]; do
        local tag_name=${1%%=*}
        local tag_value=${1#*=}
        local sanitized_tag_name
        sanitized_tag_name=$(sanitize_dimension_name "$tag_name")
        dimensions=$(printf '%s' "$dimensions" | jq -c --arg name "$sanitized_tag_name" --arg value "$tag_value" 'map(select(.Name != $name)) + [{"Name":$name,"Value":$value}]')
        shift
    done

    telemetry_records+=("$(jq -nc \
        --argjson dimensions "$dimensions" \
        --arg value "$metric_value" \
        --arg time "$(date +%s%3N)" \
        '{Dimensions:$dimensions, MeasureName:"value", MeasureValue:$value, MeasureValueType:"DOUBLE", Time:$time, TimeUnit:"MILLISECONDS"}')")

    if [ "${#telemetry_records[@]}" -ge 100 ]; then
        flush_metrics
    fi
}

flush_metrics() {
    if ! telemetry_ready || [ "${#telemetry_records[@]}" -eq 0 ]; then
        return 0
    fi

    local records_json
    records_json=$(printf '%s\n' "${telemetry_records[@]}" | jq -cs '.')

    if ! aws timestream-write write-records \
        --region "${TELEMETRY_TIMESTREAM_REGION}" \
        --database-name "${TELEMETRY_TIMESTREAM_DATABASE}" \
        --table-name "${TELEMETRY_TIMESTREAM_TABLE}" \
        --records "${records_json}" >/dev/null; then
        echo "Failed to write coverage worker telemetry to Timestream" >&2
    fi

    telemetry_records=()
}

if [ "${TELEMETRY_BACKEND}" = "timestream" ] && ! telemetry_ready; then
    echo "Telemetry backend is timestream but TELEMETRY_TIMESTREAM_* is incomplete; skipping telemetry writes" >&2
fi

ccache --show-stats

cd /tmp/bitcoin && git pull origin master
MASTER_COMMIT=$(git rev-parse HEAD)

if [ "$IS_MASTER" != "true" ]; then
    git fetch origin pull/$PR_NUM/head && git checkout FETCH_HEAD
    HEAD_COMMIT=$(git rev-parse HEAD)
    if [ "$COMMIT" != "$HEAD_COMMIT" ]; then
        echo "Commit $COMMIT is not equal to HEAD commit $HEAD_COMMIT"
        exit 1
    fi

    PR_MERGE_BASE=$(git merge-base HEAD origin/master)
    git rebase --onto "$BASE_COMMIT" "$PR_MERGE_BASE"
    S3_COVERAGE_FILE=s3://$S3_BUCKET_DATA/$PR_NUM/$HEAD_COMMIT/coverage.json
    S3_BENCH_FILE=s3://$S3_BUCKET_ARTIFACTS/$PR_NUM/$HEAD_COMMIT/bench_bitcoin
    S3_SRC_PATH=s3://$S3_BUCKET_DATA/$PR_NUM/$HEAD_COMMIT/src
    S3_COVERAGE_HTML_REPORT_PREFIX=
    S3_COVERAGE_HTML_REPORT_INDEX=
else
    git checkout $COMMIT
    S3_COVERAGE_FILE=s3://$S3_BUCKET_DATA/master/$COMMIT/coverage.json
    S3_BENCH_FILE=s3://$S3_BUCKET_ARTIFACTS/master/$COMMIT/bench_bitcoin
    S3_SRC_PATH=s3://$S3_BUCKET_DATA/master/$COMMIT/src
    S3_COVERAGE_HTML_REPORT_PREFIX=s3://$S3_BUCKET_DATA/master/$COMMIT/coverage-report
    S3_COVERAGE_HTML_REPORT_INDEX=$S3_COVERAGE_HTML_REPORT_PREFIX/index.html
fi

set +e
coverage_exists=$(aws s3 ls $S3_COVERAGE_FILE)
set -e

master_html_exists=""
if [ "$IS_MASTER" == "true" ]; then
    set +e
    master_html_exists=$(aws s3 ls $S3_COVERAGE_HTML_REPORT_INDEX)
    set -e
fi

if [ "$coverage_exists" != "" ] && { [ "$IS_MASTER" != "true" ] || [ "$master_html_exists" != "" ]; }; then
    echo "Coverage data already exists for this commit"
else
    if [ "$coverage_exists" != "" ] && [ "$IS_MASTER" == "true" ]; then
        echo "Coverage JSON exists but HTML report is missing; regenerating master coverage artifacts"
    fi

    ./test/get_previous_releases.py

    DIR_UNIT_TEST_DATA="$PWD/qa-assets/unit_test_data"
    mkdir -p "$DIR_UNIT_TEST_DATA"
    curl --location --fail https://github.com/bitcoin-core/qa-assets/raw/main/unit_test_data/script_assets_test.json -o "$DIR_UNIT_TEST_DATA/script_assets_test.json"
    export DIR_UNIT_TEST_DATA

    NPROC_2=$(expr $(nproc) \* 2)

    time cmake -B build -DCMAKE_C_COMPILER="clang" \
        -DCMAKE_BUILD_TYPE=Debug \
        -DCMAKE_CXX_COMPILER="clang++" \
        -DAPPEND_CFLAGS="-fprofile-instr-generate -fcoverage-mapping" \
        -DAPPEND_CXXFLAGS="-fprofile-instr-generate -fcoverage-mapping" \
        -DAPPEND_LDFLAGS="-fprofile-instr-generate -fcoverage-mapping"
    time cmake --build build -j$(nproc)

    # Create directory for raw profile data
    mkdir -p build/raw_profile_data
    export LLVM_PROFILE_FILE="$(pwd)/build/raw_profile_data/%m_%p.profraw"

    # Run tests to generate profiles
    time ctest --test-dir build -j $(nproc) | tee unit-tests.log

    time python3 ./build/test/functional/test_runner.py -F --previous-releases --timeout-factor=10 \
        --exclude=feature_reindex_readonly -j$(nproc) 2>&1 | tee functional-tests.log
    
    if [ "$IS_MASTER" == "true" ]; then
        binary_size=$(stat -c %s ./build/bin/bitcoind)
        queue_metric "bitcoin.bitcoin.binary_size" "$binary_size" "commit=$COMMIT"
        while IFS= read -r line; do
            if [[ $line =~ ^([a-zA-Z0-9_./-]+(\ --[a-zA-Z0-9_./-]+)*)[[:space:]]+\|[[:space:]]+.*[[:space:]]+Passed+[[:space:]]+\|[[:space:]]+([0-9]+)+[[:space:]]+s$ ]]; then
                test_name="${BASH_REMATCH[1]}";
                test_duration="${BASH_REMATCH[3]}";
                queue_metric "bitcoin.bitcoin.test.functional.duration" "$test_duration" "test_name=$test_name" "commit=$COMMIT"
                echo "test_name:$test_name,commit:$COMMIT,duration:$test_duration"
            fi;
        done < "functional-tests.log"
        flush_metrics
    fi
    
    # Merge all the raw profile data into a single file
    find build/raw_profile_data -name "*.profraw" > build/profraw_files.txt
    llvm-profdata merge -f build/profraw_files.txt -o build/coverage.profdata

    # lcov is probably the simplest format to later convert to gcovr json format
    # sticking with gcovr json format even though gcc and lcov and gcovr are not used any more as it's human readable and the rest of the app processes it
    # future work could be to just use llvm-cov output formats natively
    time llvm-cov export --format=lcov --object=build/bin/test_bitcoin --object=build/bin/bitcoind --instr-profile=build/coverage.profdata --ignore-filename-regex="src/crc32c/|src/leveldb/|src/minisketch/|src/secp256k1/|src/test/" -Xdemangler=llvm-cxxfilt > build/coverage.info
    
    python3 /convert_lcov_to_gcovr.py build/coverage.info coverage.json

    aws s3 cp coverage.json $S3_COVERAGE_FILE

    if [ "$IS_MASTER" == "true" ]; then
        rm -rf build/coverage-html
        time llvm-cov show build/bin/test_bitcoin --object=build/bin/bitcoind --instr-profile=build/coverage.profdata --ignore-filename-regex="src/crc32c/|src/leveldb/|src/minisketch/|src/secp256k1/|src/test/" -Xdemangler=llvm-cxxfilt --format=html --output-dir=build/coverage-html
        aws s3 sync --delete build/coverage-html $S3_COVERAGE_HTML_REPORT_PREFIX
    fi
fi

set +e
bench_exists=$(aws s3 ls $S3_BENCH_FILE)
set -e

if [ "$bench_exists" != "" ]; then
    echo "Bench binary already exists for this commit"
else
    rm -rf build && cmake -B build -DBUILD_TESTS=OFF -DBUILD_BENCH=ON
    time cmake --build build -j$(nproc)
    aws s3 cp build/bin/bench_bitcoin $S3_BENCH_FILE
fi

# store src folder if it doesn't exist
set +e
src_exists=$(aws s3 ls $S3_SRC_PATH)
set -e

if [ "$src_exists" == "" ]; then
    make clean || true
    rm -rf src/qt src/leveldb src/test src/wallet/test src/test src/univalue src/minisketch/ src/secp256k1 src/crc32c
    find . -name .deps -type d -exec rm -rf {} +
    aws s3 rm --recursive $S3_SRC_PATH
    aws s3 sync src $S3_SRC_PATH
fi

# store diff if it doesn't exist (only for PRs)
if [ "$IS_MASTER" != "true" ]; then
    set +e
    diff_exists=$(aws s3 ls s3://$S3_BUCKET_DATA/$PR_NUM/$HEAD_COMMIT/diff.patch)
    set -e
    
    if [ "$diff_exists" == "" ]; then
        git diff "$BASE_COMMIT" HEAD > diff.patch
        aws s3 cp diff.patch s3://$S3_BUCKET_DATA/$PR_NUM/$HEAD_COMMIT/diff.patch
    fi
fi
