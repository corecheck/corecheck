FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
git \
python3-zmq \
libevent-dev \
libboost-dev \
libsqlite3-dev \
libzmq3-dev \
build-essential \
libtool \
cmake \
pkg-config \
bsdmainutils \
bsdextrautils \
curl \
wget \
python3-pip \
lsb-release \
software-properties-common \
gnupg \
unzip \
jq \
parallel

RUN pip install mutation-core --break-system-packages

RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf awscliv2.zip aws

RUN wget https://github.com/mozilla/sccache/releases/download/v0.7.3/sccache-v0.7.3-aarch64-unknown-linux-musl.tar.gz && \
    tar -xvf sccache-v0.7.3-aarch64-unknown-linux-musl.tar.gz && \
    mv sccache-v0.7.3-aarch64-unknown-linux-musl/sccache /usr/bin/sccache && \
    chmod +x /usr/bin/sccache && \
    rm -rf sccache-v0.7.3-aarch64-unknown-linux-musl.tar.gz sccache-v0.7.3-aarch64-unknown-linux-musl
RUN ln -s /usr/bin/sccache /usr/bin/ccache

RUN git clone https://github.com/bitcoin/bitcoin.git /tmp/bitcoin
RUN git clone --depth=1 https://github.com/bitcoin-core/qa-assets.git /tmp/bitcoin/qa-assets
WORKDIR /tmp/bitcoin

RUN make -C depends NO_BOOST=1 NO_LIBEVENT=1 NO_QT=1 NO_SQLITE=1 NO_ZMQ=1 NO_USDT=1 -j$(nproc)

RUN mkdir -p /tmp/bitcoin/releases && ./test/get_previous_releases.py

RUN git config --global user.email "ci@corecheck.dev"
RUN git config --global user.name "corecheck"

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
