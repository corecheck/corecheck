FROM ubuntu:26.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y git \
    python3-zmq \
    libevent-dev \
    libboost-dev \
    libsqlite3-dev \
    libzmq3-dev \
    libcapnp-dev \
    capnproto \
    clang \
    llvm \
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

RUN sh -c "echo 'deb [signed-by=/usr/share/keyrings/datadog-archive-keyring.gpg] https://apt.datadoghq.com/ stable 7' > /etc/apt/sources.list.d/datadog.list" && \
    touch /usr/share/keyrings/datadog-archive-keyring.gpg && \
    chmod a+r /usr/share/keyrings/datadog-archive-keyring.gpg && \
    curl https://keys.datadoghq.com/DATADOG_APT_KEY_CURRENT.public | gpg --no-default-keyring --keyring /usr/share/keyrings/datadog-archive-keyring.gpg --import --batch && \
    curl https://keys.datadoghq.com/DATADOG_APT_KEY_C0962C7D.public | gpg --no-default-keyring --keyring /usr/share/keyrings/datadog-archive-keyring.gpg --import --batch && \
    curl https://keys.datadoghq.com/DATADOG_APT_KEY_F14F620E.public | gpg --no-default-keyring --keyring /usr/share/keyrings/datadog-archive-keyring.gpg --import --batch && \
    curl https://keys.datadoghq.com/DATADOG_APT_KEY_382E94DE.public | gpg --no-default-keyring --keyring /usr/share/keyrings/datadog-archive-keyring.gpg --import --batch

RUN apt update && apt install -y datadog-agent datadog-signing-keys

RUN git clone https://github.com/bitcoin/bitcoin.git /tmp/bitcoin
WORKDIR /tmp/bitcoin

RUN mkdir -p /tmp/bitcoin/releases && ./test/get_previous_releases.py

RUN git config --global user.email "ci@corecheck.dev"
RUN git config --global user.name "corecheck"

COPY convert_lcov_to_gcovr.py /convert_lcov_to_gcovr.py

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
