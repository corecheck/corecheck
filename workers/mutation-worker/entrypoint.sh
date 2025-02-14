#!/bin/bash

COMMIT=$1

ccache --show-stats

cd /tmp/bitcoin && git pull origin master
git checkout $COMMIT

S3_MUTATION_FILE=s3://$S3_BUCKET_DATA/master/$COMMIT/mutation.json

MASTER_COMMIT=$(git rev-parse HEAD)

set +e
mutation_exists=$(aws s3 ls $S3_MUTATION_FILE)
set -e

if [ "$mutation_exists" != "" ]; then
    echo "Mutation data already exists for this commit"
else
    ./test/get_previous_releases.py -b
    
    NPROC_2=$(expr $(nproc) \* 2)
    
    time cmake -B build -DBerkeleyDB_INCLUDE_DIR:PATH="${BDB_PREFIX}/include" -DWITH_BDB=ON

    mutation-core mutate -f="src/wallet/coinselection.cpp"
    mutation-core mutate -f="src/script/interpreter.cpp"
    mutation-core mutate -f="src/consensus/tx_verify.cpp"
    mutation-core mutate -f="src/consensus/tx_check.cpp"
    mutation-core mutate -f="src/consensus/merkle.cpp"
    mutation-core mutate -f="src/pow.cpp"
    mutation-core mutate -f="src/addrman.cpp"

    mutation-core analyze -f="muts-coinselection-cpp" -c="cmake --build build -j 5 && ./build/src/test/test_bitcoin --run_test=coinselector_tests && ./build/src/test/test_bitcoin --run_test=spend_tests && ./build/test/functional/rpc_psbt.py && ./build/test/functional/wallet_fundrawtransaction.py"
    mutation-core analyze -f="muts-interpreter-cpp" -c="cmake --build build -j 5 && ./build/src/test/test_bitcoin --run_test=script_tests && ./build/src/test/test_bitcoin --run_test=scriptnum_tests && ./build/src/test/test_bitcoin --run_test=script_p2sh_tests && ./build/src/test/test_bitcoin --run_test=script_parse_tests && ./build/src/test/test_bitcoin --run_test=script_p2sh_tests && ./build/src/test/test_bitcoin --run_test=script_standard_tests"
    mutation-core analyze -f="muts-tx_verify-cpp" -c="cmake --build build -j 5 && ./build/src/test/test_bitcoin --run_test=miner_tests && ./build/src/test/test_bitcoin --run_test=sigopcount_tests && ./build/src/test/test_bitcoin --run_test=validation_tests && ./build/src/test/test_bitcoin --run_test=mempool_tests && ./build/test/functional/feature_block.py && ./build/test/functional/mining_basic.py && /build/test/functional/mempool_accept.py && ./build/test/functional/mempool_compatibility.py"
    mutation-core analyze -f="muts-tx_check-cpp" -c="cmake --build build -j 5 &&  ./build/src/test/test_bitcoin --run_test=sighash_tests && ./build/src/test/test_bitcoin --run_test=transaction_tests && ./build/src/test/test_bitcoin --run_test=validation_tests && ./build/src/test/test_bitcoin --run_test=mempool_tests && ./build/test/functional/feature_block.py && ./build/test/functional/mining_basic.py && /build/test/functional/mempool_accept.py && ./build/test/functional/mempool_compatibility.py"
    mutation-core analyze -f="muts-merkle-cpp" -c="cmake --build build -j 5 && ./build/src/test/test_bitcoin --run_test=validation_tests && ./build/src/test/test_bitcoin --run_test=merkle_tests && ./build/src/test/test_bitcoin --run_test=blockencodings_tests && ./build/src/test/test_bitcoin --run_test=validation_block_tests && ./build/src/test/test_bitcoin --run_test=miner_tests && ./build/test/functional/feature_block.py"
    mutation-core analyze -f="muts-pow-cpp" -c="cmake --build build -j 5 && ./build/src/test/test_bitcoin --run_test=pow_tests && ./build/src/test/test_bitcoin --run_test=integer_tests && ./build/src/test/test_bitcoin --run_test=headers_sync_chainwork_tests && ./build/test/functional/feature_block.py"
    mutation-core analyze -f="muts-addrman-cpp" -c="cmake --build build -j 5 && ./build/src/test/test_bitcoin --run_test=addrman_tests && ./build/test/functional/feature_addrman.py && ./build/test/functional/feature_asmap.py && ./build/test/functional/rpc_net.py"

    time cmake --build build -j$(nproc)
    
    aws s3 cp coverage.json $S3_MUTATION_FILE
fi