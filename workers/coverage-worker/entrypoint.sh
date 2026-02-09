#!/bin/bash -x

COMMIT=$1
PR_NUM=$2
IS_MASTER=$3
BASE_COMMIT=$4

set -e

sh -c "sed \"s/api_key:.*/api_key: $DD_API_KEY/\" /etc/datadog-agent/datadog.yaml.example > /etc/datadog-agent/datadog.yaml"
sh -c "sed -i 's/# site:.*/site: datadoghq.eu/' /etc/datadog-agent/datadog.yaml"
sh -c "sed -i 's/# hostname:.*/hostname: $HOSTNAME/' /etc/datadog-agent/datadog.yaml"

/etc/init.d/datadog-agent start
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
    
    git rebase $BASE_COMMIT
    S3_COVERAGE_FILE=s3://$S3_BUCKET_DATA/$PR_NUM/$HEAD_COMMIT/coverage.json
    S3_BENCH_FILE=s3://$S3_BUCKET_ARTIFACTS/$PR_NUM/$HEAD_COMMIT/bench_bitcoin
    S3_SRC_PATH=s3://$S3_BUCKET_DATA/$PR_NUM/$HEAD_COMMIT/src
else
   git checkout $COMMIT
    S3_COVERAGE_FILE=s3://$S3_BUCKET_DATA/master/$COMMIT/coverage.json
    S3_BENCH_FILE=s3://$S3_BUCKET_ARTIFACTS/master/$COMMIT/bench_bitcoin
    S3_SRC_PATH=s3://$S3_BUCKET_DATA/master/$COMMIT/src
fi

set +e
coverage_exists=$(aws s3 ls $S3_COVERAGE_FILE)
set -e

if [ "$coverage_exists" != "" ]; then
    echo "Coverage data already exists for this commit"
else

    ./test/get_previous_releases.py

    DIR_UNIT_TEST_DATA="$PWD/qa-assets/unit_test_data"
    mkdir -p "$DIR_UNIT_TEST_DATA"
    curl --location --fail https://github.com/bitcoin-core/qa-assets/raw/main/unit_test_data/script_assets_test.json -o "$DIR_UNIT_TEST_DATA/script_assets_test.json"
    export DIR_UNIT_TEST_DATA

    NPROC_2=$(expr $(nproc) \* 2)

    time cmake -B build -DCMAKE_C_COMPILER="clang" \
       -DCMAKE_CXX_COMPILER="clang++" \
       -DAPPEND_CFLAGS="-fprofile-instr-generate -fcoverage-mapping" \
       -DAPPEND_CXXFLAGS="-fprofile-instr-generate -fcoverage-mapping" \
       -DAPPEND_LDFLAGS="-fprofile-instr-generate -fcoverage-mapping"
    time cmake --build build -j$(nproc)

    # Create directory for raw profile data
    mkdir -p build/raw_profile_data
    LLVM_PROFILE_FILE="$(pwd)/build/raw_profile_data/%m_%p.profraw"

    env
    # Run tests to generate profiles
    time ctest --test-dir build -j $(nproc) | tee unit-tests.log
    ls -la build/raw_profile_data
    time python3 ./build/test/functional/test_runner.py -F --previous-releases --timeout-factor=10 \
        --exclude=feature_reindex_readonly -j$(nproc) 2>&1 | tee functional-tests.log
    ls -la build/raw_profile_data
    
    if [ "$IS_MASTER" == "true" ]; then
        binary_size=$(stat -c %s ./build/bin/bitcoind)
        echo -n "bitcoin.bitcoin.binary_size:$binary_size|g|#commit:$COMMIT" >/dev/udp/localhost/8125
        while IFS= read -r line; do
            if [[ $line =~ ^([a-zA-Z0-9_./-]+(\ --[a-zA-Z0-9_./-]+)*)[[:space:]]+\|[[:space:]]+.*[[:space:]]+Passed+[[:space:]]+\|[[:space:]]+([0-9]+)+[[:space:]]+s$ ]]; then
                test_name="${BASH_REMATCH[1]}";
                test_duration="${BASH_REMATCH[3]}";
                echo -n "bitcoin.bitcoin.test.functional.duration:$test_duration|g|#test_name:$test_name,#commit:$COMMIT" >/dev/udp/localhost/8125
                echo "test_name:$test_name,commit:$COMMIT,duration:$test_duration"
            fi;
        done < "functional-tests.log"
    fi
    
    # Merge all the raw profile data into a single file
    find build/raw_profile_data -name "*.profraw" | xargs llvm-profdata merge -o build/coverage.profdata

    # lcov is probably the simplest format to later convert to gcovr json format
    # sticking with gcovr json format even though gcc and lcov and gcovr are not used any more as it's human readable and the rest of the app processes it
    # future work could be to just use llvm-cov output formats natively
    time llvm-cov export --format=lcov --object=build/bin/test_bitcoin --object=build/bin/bitcoind --instr-profile=build/coverage.profdata --ignore-filename-regex="src/crc32c/|src/leveldb/|src/minisketch/|src/secp256k1/|src/test/" -Xdemangler=llvm-cxxfilt > build/coverage.info
    
    python3 /convert_lcov_to_gcovr.py coverage.info coverage.json
    
    aws s3 cp coverage.json $S3_COVERAGE_FILE
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
        git diff $BASE_COMMIT > diff.patch
        aws s3 cp diff.patch s3://$S3_BUCKET_DATA/$PR_NUM/$HEAD_COMMIT/diff.patch
    fi
fi
