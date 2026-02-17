#!/bin/bash
set -x

COMMIT=$1

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
    cargo install bcore-mutation

    ./test/get_previous_releases.py

    NPROC_2=$(expr $(nproc) \* 2)

    time cmake -B build -DBUILD_FUZZ_BINARY=ON -DCMAKE_BUILD_TYPE=Debug

    bcore-mutation mutate -f="src/wallet/coinselection.cpp" --one-mutant
    bcore-mutation mutate -f="src/script/interpreter.cpp" --one-mutant
    bcore-mutation mutate -f="src/txgraph.cpp" --one-mutant
    bcore-mutation mutate -f="src/consensus/tx_verify.cpp" --one-mutant
    bcore-mutation mutate -f="src/consensus/tx_check.cpp" --one-mutant
    bcore-mutation mutate -f="src/consensus/merkle.cpp" --one-mutant
    bcore-mutation mutate -f="src/pow.cpp" --one-mutant
    bcore-mutation mutate -f="src/addrman.cpp" --one-mutant
    bcore-mutation mutate -f="src/util/asmap.cpp" --one-mutant
    bcore-mutation mutate -f="src/script/descriptor.cpp" --one-mutant
    bcore-mutation mutate -f="src/netgroup.cpp" --one-mutant
    bcore-mutation mutate -f="src/psbt.cpp" --one-mutant

    bcore-mutation analyze -f="muts-wallet-coinselection-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=coinselection_tests && ./build/bin/test_bitcoin --run_test=coinselector_tests && ./build/bin/test_bitcoin --run_test=spend_tests && ./build/test/functional/rpc_psbt.py && ./build/test/functional/wallet_fundrawtransaction.py && FUZZ=coin_grinder ./build/bin/fuzz qa-assets/fuzz_corpora/coin_grinder && FUZZ=coin_grinder_is_optimal ./build/bin/fuzz qa-assets/fuzz_corpora/coin_grinder_is_optimal && FUZZ=bnb_finds_min_waste ./build/bin/fuzz qa-assets/fuzz_corpora/bnb_finds_min_waste && FUZZ=coinselection_knapsack ./build/bin/fuzz qa-assets/fuzz_corpora/coinselection_knapsack && FUZZ=coinselection_srd ./build/bin/fuzz qa-assets/fuzz_corpora/coinselection_srd && FUZZ=coinselection_bnb ./build/bin/fuzz qa-assets/fuzz_corpora/coinselection_bnb"
    bcore-mutation analyze -f="muts-script-interpreter-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=miniscript_tests && ./build/bin/test_bitcoin --run_test=transaction_tests && DIR_UNIT_TEST_DATA=qa-assets/unit_test_data ./build/bin/test_bitcoin --run_test=script_assets_tests && ./build/bin/test_bitcoin --run_test=script_tests && ./build/bin/test_bitcoin --run_test=scriptnum_tests && ./build/bin/test_bitcoin --run_test=script_p2sh_tests && ./build/bin/test_bitcoin --run_test=script_parse_tests && ./build/bin/test_bitcoin --run_test=script_p2sh_tests && ./build/bin/test_bitcoin --run_test=script_standard_tests && ./build/test/functional/feature_taproot.py && ./build/test/functional/rpc_decodescript.py && ./build/test/functional/wallet_miniscript.py && ./build/test/functional/feature_block.py --skipreorg"
    bcore-mutation analyze -f="muts-txgraph-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=txgraph_tests && FUZZ=txgraph ./build/bin/fuzz qa-assets/fuzz_corpora/txgraph"
    bcore-mutation analyze -f="muts-consensus-tx_verify-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=script_p2sh_tests && ./build/bin/test_bitcoin --run_test=miner_tests && ./build/bin/test_bitcoin --run_test=sigopcount_tests && ./build/bin/test_bitcoin --run_test=validation_tests && ./build/bin/test_bitcoin --run_test=mempool_tests && ./build/test/functional/feature_block.py --skipreorg && ./build/test/functional/mining_basic.py && ./build/test/functional/mempool_accept.py && ./build/test/functional/mempool_compatibility.py"
    bcore-mutation analyze -f="muts-consensus-tx_check-cpp" -c="cmake --build build -j $(nproc) &&  ./build/bin/test_bitcoin --run_test=sighash_tests && ./build/bin/test_bitcoin --run_test=transaction_tests && ./build/bin/test_bitcoin --run_test=validation_tests && ./build/bin/test_bitcoin --run_test=mempool_tests && ./build/test/functional/feature_block.py --skipreorg && ./build/test/functional/mining_basic.py && ./build/test/functional/mempool_accept.py && ./build/test/functional/mempool_compatibility.py"
    bcore-mutation analyze -f="muts-consensus-merkle-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=validation_tests && ./build/bin/test_bitcoin --run_test=merkle_tests && ./build/bin/test_bitcoin --run_test=blockencodings_tests && ./build/bin/test_bitcoin --run_test=validation_block_tests && ./build/bin/test_bitcoin --run_test=miner_tests && ./build/test/functional/feature_block.py --skipreorg"
    bcore-mutation analyze -f="muts-pow-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=pow_tests && ./build/bin/test_bitcoin --run_test=headers_sync_chainwork_tests && ./build/test/functional/feature_block.py --skipreorg"
    bcore-mutation analyze -f="muts-addrman-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=addrman_tests && ./build/test/functional/feature_addrman.py && ./build/test/functional/feature_asmap.py && ./build/test/functional/rpc_net.py && ./build/test/functional/p2p_invalid_messages.py"
    bcore-mutation analyze -f="muts-util-asmap-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=addrman_tests && ./build/test/functional/feature_asmap.py && ./build/test/functional/feature_addrman.py"
    bcore-mutation analyze -f="muts-script-descriptor-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=descriptor_tests && ./build/test/functional/wallet_descriptor.py && ./build/test/functional/wallet_multisig_descriptor_psbt.py"
    bcore-mutation analyze -f="muts-netgroup-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=netbase_tests && ./build/bin/test_bitcoin --run_test=addrman_tests && ./build/test/functional/feature_asmap.py && ./build/test/functional/feature_addrman.py"
    bcore-mutation analyze -f="muts-psbt-cpp" -c="cmake --build build -j $(nproc) && ./build/bin/test_bitcoin --run_test=psbt_wallet_tests && FUZZ=psbt ./build/bin/fuzz qa-assets/fuzz_corpora/psbt && ./build/test/functional/rpc_psbt.py && ./build/test/functional/wallet_fundrawtransaction.py && ./build/test/functional/wallet_miniscript_decaying_multisig_descriptor_psbt.py && ./build/test/functional/wallet_multisig_descriptor_psbt.py && ./build/test/functional/wallet_musig.py && ./build/test/functional/wallet_taproot.py"
    sccache --show-stats

    aws s3 cp diff_not_killed.json $S3_MUTATION_FILE
fi
