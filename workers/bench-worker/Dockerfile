FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt update && apt install -y git python3-zmq libevent-dev libboost-dev libdb5.3++-dev libsqlite3-dev libminiupnpc-dev libzmq3-dev curl wget python3-pip lsb-release software-properties-common gnupg unzip jq parallel zip vim htop kmod valgrind
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-aarch64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf awscliv2.zip aws

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
