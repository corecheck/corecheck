#!/bin/bash
COMMIT=$1
PR_NUM=$2
IS_MASTER=$3
BASE_COMMIT=$4

# Check if branch already exists on sonarcloud
if [ "$IS_MASTER" != "true" ]; then
    # https://sonarcloud.io/api/navigation/component?component=aureleoules_bitcoin&branch=XXXX
    BRANCH_EXISTS=$(curl -s "https://sonarcloud.io/api/navigation/component?component=aureleoules_bitcoin&branch=$PR_NUM" | jq -r '.branch')
    if [ "$BRANCH_EXISTS" == "$PR_NUM" ]; then
        echo "Branch $PR_NUM already exists on sonarcloud"
        exit 0
    fi
fi

set -e
ccache --show-stats

cd /tmp/bitcoin && git pull origin master
MASTER_COMMIT=$(git rev-parse HEAD)

if [ "$IS_MASTER" != "true" ]; then
    git fetch origin pull/$PR_NUM/head && git checkout FETCH_HEAD
    HEAD_COMMIT=$(git rev-parse HEAD)
    if [ "$COMMIT" != "$HEAD_COMMIT" ]; then
        echo "Commit $COMMIT is not equal to HEAD commit $HEAD_COMMIT"
        exit 0
    fi
    
    git rebase master
else
    if [ "$COMMIT" != "$MASTER_COMMIT" ]; then
        echo "Commit $COMMIT is not equal to master commit $MASTER_COMMIT"
        exit 0
    fi
fi

./test/get_previous_releases.py -b

./autogen.sh && ./configure --disable-fuzz --enable-fuzz-binary=no --with-gui=no --disable-zmq BDB_LIBS="-L${BDB_PREFIX}/lib -ldb_cxx-4.8" BDB_CFLAGS="-I${BDB_PREFIX}/include"
time compiledb make -j$(nproc)

if [ "$IS_MASTER" != "true" ]; then
    echo "Updating $PR_NUM branch on sonarcloud"
    time /usr/lib/sonar-scanner/bin/sonar-scanner \
    -Dsonar.organization=aureleoules \
    -Dsonar.projectKey=aureleoules_bitcoin \
    -Dsonar.sources=. \
    -Dsonar.cfamily.compile-commands=compile_commands.json \
    -Dsonar.host.url=https://sonarcloud.io \
    -Dsonar.exclusions='src/crc32c/**, src/crypto/ctaes/**, src/leveldb/**, src/minisketch/**, src/secp256k1/**, src/univalue/**' \
    -Dsonar.cfamily.threads=$(nproc) \
    -Dsonar.branch.name=$PR_NUM \
    -Dsonar.cfamily.analysisCache.mode=server \
    -Dsonar.branch.target=master
else
    echo "Updating master branch on sonarcloud"
    time /usr/lib/sonar-scanner/bin/sonar-scanner \
    -Dsonar.organization=aureleoules \
    -Dsonar.projectKey=aureleoules_bitcoin \
    -Dsonar.sources=. \
    -Dsonar.cfamily.compile-commands=compile_commands.json \
    -Dsonar.host.url=https://sonarcloud.io \
    -Dsonar.exclusions='src/crc32c/**, src/crypto/ctaes/**, src/leveldb/**, src/minisketch/**, src/secp256k1/**, src/univalue/**' \
    -Dsonar.cfamily.threads=$(nproc) \
    -Dsonar.branch.name=master \
    -Dsonar.cfamily.analysisCache.mode=server
fi
