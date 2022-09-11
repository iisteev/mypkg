#!/usr/bin/env bash
set -e

FLAGS=()
BUILD_OS=$1
BUILD_ARCH=$2

if [ -n "$VERBOSE" ]; then
    FLAGS+=(-v)
fi

if [ -z "$BUILD_OS" ]; then
    BUILD_OS=linux
fi

if [ -z "$BUILD_ARCH" ]; then
    BUILD_ARCH=amd64
fi

if [ -z "$VERSION" ]; then
    VERSION=$(git rev-parse HEAD)
fi

if [ -z "$BUILD_DIR" ]; then
    BUILD_DIR=$(pwd)/dist
fi

if [ -z "$DATE" ]; then
    DATE=$(date -u '+%Y-%m-%d_%I:%M:%S%p')
fi

# Build binaries
# shellcheck disable=SC2086
CGO_ENABLED=0 GOGC=off GOOS=$BUILD_OS GOARCH=$BUILD_ARCH go install ${FLAGS[*]} -ldflags "-s -w \
    -X mypkg/cmd.Version=$VERSION \
    -X mypkg/cmd.BuildDate=$DATE" \
    -a -installsuffix nocgo mypkg

