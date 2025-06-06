#!/bin/bash

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
    ./test/get_previous_releases.py -b
    
    NPROC_2=$(expr $(nproc) \* 2)
    
    time cmake -B build -DCMAKE_BUILD_TYPE=Coverage
    time cmake --build build -j$(nproc)
    
    time ./build/bin/test_bitcoin --list_content 2>&1 | grep -v "    " | parallel --halt now,fail=1 ./build/bin/test_bitcoin -t {} 2>&1
    time python3 ./build/test/functional/test_runner.py -F --previous-releases --timeout-factor=10 --exclude=feature_reindex_readonly,feature_dbcrash -j$NPROC_2 &> functional-tests.log
    
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
    
    time gcovr --json --merge-mode-functions=merge-use-line-min --gcov-ignore-errors=no_working_dir_found --gcov-ignore-parse-errors -e depends -e src/test -e src/leveldb -e src/bench -e src/qt -j $(nproc) > coverage.json
    
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
