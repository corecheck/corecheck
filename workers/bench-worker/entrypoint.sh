#!/bin/bash

COMMIT=$1
PR_NUM=$2
IS_MASTER=$3

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

BENCH_DURATION=10000
# set perf max sample rate to 1
echo 1 | tee /proc/sys/kernel/perf_event_max_sample_rate

bench_list=$(./bench_bitcoin -list)
bench_list=$(echo "$bench_list" | grep -v "BenchTime") # Do not run BenchTime* benchmarks
bench_list=$(echo "$bench_list" | grep -v "LoadExternalBlockFile") # Do not run LoadExternalBlockFile* benchmarks
bench_list=$(echo "$bench_list" | grep -v "LogPrint*") # Do not run LogPrint* benchmarks

time for bench in $bench_list; do
    echo "Running $bench"
    # run bench_bitcoin with valgrind
    taskset -c 1 valgrind --tool=cachegrind --I1=32768,8,64 --D1=32768,8,64 --LL=8388608,16,64 --cachegrind-out-file=bench_$bench.cachegrind ./bench_bitcoin -filter=$bench -min-time=$BENCH_DURATION
done

# bench.json
total_bench="["
# convert each cachegrind file summary to json
for bench in $bench_list; do
# events: Ir I1mr ILmr Dr D1mr DLmr Dw D1mw DLmw 
    cachegrind_events=$(grep 'events:' bench_$bench.cachegrind | sed 's/events: //')
    cachegrind_summary=$(grep 'summary:' bench_$bench.cachegrind | sed 's/summary: //')

    # split by space the events
    IFS=' ' read -r -a events <<< "$cachegrind_events"
    IFS=' ' read -r -a summary <<< "$cachegrind_summary"

    # create json object
    json="{"
    # add bench name
    json="$json\"name\": \"$bench\","
    for i in "${!events[@]}"; do
        json="$json\"${events[$i]}\": ${summary[$i]},"
    done

    # remove last comma
    json="${json::-1}}"
    total_bench="$total_bench$json,"
done

# remove last comma
total_bench="${total_bench::-1}]"
echo "$total_bench" > bench.json

if [ "$IS_MASTER" != "true" ]; then
    aws s3 cp bench.json "s3://$S3_BUCKET_DATA/$PR_NUM/$COMMIT/bench/bench-$AWS_BATCH_JOB_ARRAY_INDEX.json"
else
    aws s3 cp bench.json "s3://$S3_BUCKET_DATA/master/$COMMIT/bench/bench-$AWS_BATCH_JOB_ARRAY_INDEX.json"
fi
