ARG GO_VERSION=1.21
ARG MIRROR_REGISTRY

FROM ${MIRROR_REGISTRY}golang:${GO_VERSION}-alpine AS builder-image
RUN apk update && apk add \
    binutils \
    make \
    rpm \
    ruby \
    ruby-dev \
    gcc \
    musl-dev \
    tar \
    msitools \
    uuidgen \
    coreutils \
    zip \
    git \
    && gem install fpm

FROM builder-image AS ut-stage
WORKDIR /ut
COPY . .
RUN go install github.com/jstemmer/go-junit-report/v2@latest
RUN go test -v 2>&1 ./... | go-junit-report > ut-report.xml

FROM scratch AS ut-artifacts
COPY --from=ut-stage /ut/ut-report.xml /

FROM builder-image AS build-stage
WORKDIR /app
COPY . .

RUN make msi GOARCH=amd64 && \
    make deb GOARCH=amd64 && \
    make rpm GOARCH=amd64 && \
    make osx GOARCH=amd64 && \
    make deb GOARCH=arm64 && \
    make rpm GOARCH=arm64 && \
    make osx GOARCH=arm64

FROM scratch AS build-artifacts
COPY --from=build-stage /app/dist /
