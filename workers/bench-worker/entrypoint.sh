#!/bin/bash

COMMIT=$1
PR_NUM=$2
IS_MASTER=$3
BASE_COMMIT=$4

# check if benchmark already exists
if [ "$IS_MASTER" != "true" ]; then
    aws s3 ls "s3://$S3_BUCKET_DATA/$PR_NUM/$COMMIT/bench/bench-$AWS_BATCH_JOB_ARRAY_INDEX.json" && exit 0
else
    aws s3 ls "s3://$S3_BUCKET_DATA/master/$COMMIT/bench/bench-$AWS_BATCH_JOB_ARRAY_INDEX.json" && exit 0
fi

set -e

if [ "$IS_MASTER" != "true" ]; then
    aws s3 cp s3://$S3_BUCKET_ARTIFACTS/$PR_NUM/$COMMIT/bench_bitcoin bench_bitcoin
    chmod +x bench_bitcoin
else
    aws s3 cp s3://$S3_BUCKET_ARTIFACTS/master/$COMMIT/bench_bitcoin bench_bitcoin
    chmod +x bench_bitcoin
fi

BENCH_DURATION=1000
# set perf max sample rate to 1
echo 1 | tee /proc/sys/kernel/perf_event_max_sample_rate
./bench_bitcoin -min-time=$BENCH_DURATION -output-json=bench.json

if [ "$IS_MASTER" != "true" ]; then
    aws s3 cp bench.json "s3://$S3_BUCKET_DATA/$PR_NUM/$COMMIT/bench/bench-$AWS_BATCH_JOB_ARRAY_INDEX.json"
else
    aws s3 cp bench.json "s3://$S3_BUCKET_DATA/master/$COMMIT/bench/bench-$AWS_BATCH_JOB_ARRAY_INDEX.json"
fi
