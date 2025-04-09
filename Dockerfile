FROM ghcr.io/vexxhost/ubuntu:edge AS base
RUN \
  --mount=type=cache,target=/var/cache/apt,sharing=locked \
  --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends \
      ca-certificates \
      libmysqlclient21

FROM base AS builder
RUN \
  --mount=type=cache,target=/var/cache/apt,sharing=locked \
  --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends \
      curl \
      build-essential \
      libmysqlclient-dev
ADD --chmod=0755 https://sh.rustup.rs /usr/local/bin/rustup
RUN rustup -y
ENV PATH=/root/.cargo/bin:$PATH
ADD . /src
WORKDIR /src
RUN cargo install --path .

FROM base
COPY --from=builder /root/.cargo/bin/openstack-database-exporter /usr/local/bin/openstack_database_exporter
EXPOSE 9180
ENTRYPOINT ["/usr/local/bin/openstack_database_exporter"]
