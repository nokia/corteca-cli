# syntax=docker/dockerfile:1
#
# Runtime container image for corteca CLI.
# Includes the corteca binary and Docker CLI (required for `docker buildx build`).
#
# Requires pre-built binaries in dist/bin/ (produced by `make`).
# See doc/BUILD.md for local build and test instructions.

ARG DOCKER_VERSION=27

FROM docker:${DOCKER_VERSION}-cli

LABEL org.opencontainers.image.title="corteca-cli" \
      org.opencontainers.image.description="Corteca Developer Toolkit — CLI for building and deploying Corteca applications" \
      org.opencontainers.image.source="https://github.com/nokia/corteca-cli" \
      org.opencontainers.image.licenses="BSD-3-Clause"

RUN apk add --no-cache ca-certificates

ARG TARGETARCH
ARG VERSION
COPY dist/bin/corteca-linux-${TARGETARCH}-${VERSION} /usr/local/bin/corteca
COPY data/ /etc/corteca/

ENTRYPOINT ["corteca"]
