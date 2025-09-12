FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && apt install -y git python3-zmq libevent-dev libboost-dev libsqlite3-dev libzmq3-dev libcapnp-dev capnproto lcov build-essential libtool cmake pkg-config bsdmainutils bsdextrautils curl wget python3-pip lsb-release software-properties-common gnupg unzip jq
RUN pip install compiledb --break-system-packages 
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

RUN wget https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-5.0.1.3006-linux.zip && \
    unzip sonar-scanner-cli-5.0.1.3006-linux.zip && \
    mv sonar-scanner-5.0.1.3006-linux /usr/lib/sonar-scanner && \
    rm -rf sonar-scanner-cli-5.0.1.3006-linux.zip sonar-scanner-5.0.1.3006-linux

RUN git clone https://github.com/bitcoin/bitcoin.git /tmp/bitcoin
WORKDIR /tmp/bitcoin

RUN mkdir -p /tmp/bitcoin/releases && ./test/get_previous_releases.py

RUN git config --global user.email "ci@corecheck.dev"
RUN git config --global user.name "corecheck"

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
