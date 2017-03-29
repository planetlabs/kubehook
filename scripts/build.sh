#!/usr/bin/env sh

set -e

DIST="dist"
VENDOR="vendor"

export GOOS="linux"
export GOARCH="amd64"

[ -d "${DIST}" ] || mkdir -p "${DIST}"
[ -d "${VENDOR}" ] && rm -rf "${VENDOR}"

# Ideally we'd use a specific version of Glide.
go get -u github.com/Masterminds/glide

# Create the vendor directory based on glide.lock
glide install

# Build the binary
go build -o "${DIST}/kubehook" ./cmd/kubehook

# Create the docker image
VERSION=$(git rev-parse --short HEAD)
docker build --tag "negz/kubehook:latest" .
docker build --tag "negz/kubehook:${VERSION}" .