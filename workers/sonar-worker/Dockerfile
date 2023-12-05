FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && apt install -y git python3-zmq libevent-dev libboost-dev libdb5.3++-dev libsqlite3-dev libminiupnpc-dev libzmq3-dev lcov build-essential libtool autotools-dev automake pkg-config bsdmainutils bsdextrautils curl wget python3-pip lsb-release software-properties-common gnupg unzip jq
RUN pip install compiledb
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf awscliv2.zip aws

RUN wget https://github.com/mozilla/sccache/releases/download/v0.7.4/sccache-v0.7.4-x86_64-unknown-linux-musl.tar.gz && \ 
    tar -xvf sccache-v0.7.4-x86_64-unknown-linux-musl.tar.gz && \
    mv sccache-v0.7.4-x86_64-unknown-linux-musl/sccache /usr/bin/sccache && \
    chmod +x /usr/bin/sccache && \
    rm -rf sccache-v0.7.4-x86_64-unknown-linux-musl.tar.gz sccache-v0.7.4-x86_64-unknown-linux-musl
RUN ln -s /usr/bin/sccache /usr/bin/ccache

RUN wget https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-4.7.0.2747-linux.zip && \
    unzip sonar-scanner-cli-4.7.0.2747-linux.zip && \
    mv sonar-scanner-4.7.0.2747-linux /usr/lib/sonar-scanner && \
    rm -rf sonar-scanner-cli-4.7.0.2747-linux.zip

RUN git clone https://github.com/bitcoin/bitcoin.git /tmp/bitcoin
WORKDIR /tmp/bitcoin
RUN make -C depends NO_BOOST=1 NO_LIBEVENT=1 NO_QT=1 NO_SQLITE=1 NO_NATPMP=1 NO_UPNP=1 NO_ZMQ=1 NO_USDT=1 -j$(nproc)
ENV BDB_PREFIX=/tmp/bitcoin/depends/x86_64-pc-linux-gnu
RUN mkdir -p /tmp/bitcoin/releases && ./test/get_previous_releases.py -b

RUN git config --global user.email "ci@corecheck.dev"
RUN git config --global user.name "corecheck"

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
