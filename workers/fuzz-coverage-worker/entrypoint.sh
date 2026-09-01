#!/bin/bash -x

COMMIT=$1

set -e

sh -c "sed \"s/api_key:.*/api_key: $DD_API_KEY/\" /etc/datadog-agent/datadog.yaml.example > /etc/datadog-agent/datadog.yaml"
sh -c "sed -i 's/# site:.*/site: datadoghq.eu/' /etc/datadog-agent/datadog.yaml"
sh -c "sed -i 's/# hostname:.*/hostname: $HOSTNAME/' /etc/datadog-agent/datadog.yaml"

/etc/init.d/datadog-agent start
ccache --show-stats

cd /tmp/bitcoin && git fetch origin master && git checkout "$COMMIT"

S3_COVERAGE_FILE=s3://$S3_BUCKET_DATA/master-fuzz/$COMMIT/coverage.json
S3_SRC_PATH=s3://$S3_BUCKET_DATA/master-fuzz/$COMMIT/src
S3_COVERAGE_HTML_REPORT_PREFIX=s3://$S3_BUCKET_DATA/master-fuzz/$COMMIT/coverage-report
S3_COVERAGE_HTML_REPORT_INDEX=$S3_COVERAGE_HTML_REPORT_PREFIX/index.html

set +e
coverage_exists=$(aws s3 ls $S3_COVERAGE_FILE)
html_exists=$(aws s3 ls $S3_COVERAGE_HTML_REPORT_INDEX)
set -e

if [ "$coverage_exists" != "" ] && [ "$html_exists" != "" ]; then
    echo "Fuzz coverage data already exists for this commit"
else
    # Build the fuzz binary with source-based coverage instrumentation.
    # BUILD_FOR_FUZZING=ON produces build/bin/fuzz (libFuzzer when compiling with clang).
    time cmake -B build -DCMAKE_C_COMPILER="clang" \
        -DCMAKE_BUILD_TYPE=Debug \
        -DCMAKE_CXX_COMPILER="clang++" \
        -DBUILD_FOR_FUZZING=ON \
        -DAPPEND_CFLAGS="-fprofile-instr-generate -fcoverage-mapping" \
        -DAPPEND_CXXFLAGS="-fprofile-instr-generate -fcoverage-mapping" \
        -DAPPEND_LDFLAGS="-fprofile-instr-generate -fcoverage-mapping"
    time cmake --build build -j$(nproc)

    # Fetch the fuzz corpora (inputs) maintained by bitcoin-core/qa-assets.
    rm -rf /tmp/qa-assets
    git clone --depth 1 https://github.com/bitcoin-core/qa-assets.git /tmp/qa-assets

    # Create directory for raw profile data. %m_%p keeps one profraw per
    # fuzz target binary image and process so parallel replays don't clobber.
    mkdir -p build/raw_profile_data
    export LLVM_PROFILE_FILE="$(pwd)/build/raw_profile_data/%m_%p.profraw"

    # Replay every corpus input through its fuzz target (libFuzzer -runs=1).
    # The build-dir runner reads build/test/config.ini and defaults the binary
    # to build/bin/fuzz.
    time python3 ./build/test/fuzz/test_runner.py -j$(nproc) /tmp/qa-assets/fuzz_corpora 2>&1 | tee fuzz-tests.log

    # Merge all the raw profile data into a single file
    find build/raw_profile_data -name "*.profraw" > build/profraw_files.txt
    llvm-profdata merge -f build/profraw_files.txt -o build/coverage.profdata

    # Export to lcov then convert to the gcovr json format the rest of the app consumes.
    time llvm-cov export --format=lcov --object=build/bin/fuzz --instr-profile=build/coverage.profdata --ignore-filename-regex="src/crc32c/|src/leveldb/|src/minisketch/|src/secp256k1/|src/test/" -Xdemangler=llvm-cxxfilt > build/coverage.info

    python3 /convert_lcov_to_gcovr.py build/coverage.info coverage.json

    aws s3 cp coverage.json $S3_COVERAGE_FILE

    rm -rf build/coverage-html
    time llvm-cov show build/bin/fuzz --instr-profile=build/coverage.profdata --ignore-filename-regex="src/crc32c/|src/leveldb/|src/minisketch/|src/secp256k1/|src/test/" -Xdemangler=llvm-cxxfilt --format=html --output-dir=build/coverage-html
    aws s3 sync --delete build/coverage-html $S3_COVERAGE_HTML_REPORT_PREFIX
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
