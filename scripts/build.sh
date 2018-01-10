#!/usr/bin/env bash

set -ex

DIST="dist"
VENDOR="vendor"

export GOOS="linux"
export GOARCH="amd64"
export CGO_ENABLED=0

rm -rf "${DIST}"
rm -rf "${VENDOR}"
mkdir -p "${DIST}/frontend"

pushd frontend
npm install
npm run build
popd

cp frontend/index.html "${DIST}/frontend/"
cp -R frontend/dist/ "${DIST}/frontend/"

go get -u github.com/Masterminds/glide
go get -u github.com/rakyll/statik

# Create the vendor directory based on glide.lock
glide install

pushd statik
go generate
popd

# Build the binary
go build -o "${DIST}/kubehook" ./cmd/kubehook

# Create the docker image
VERSION=$(git rev-parse --short HEAD)
docker build --tag "negz/kubehook:latest" .
docker build --tag "negz/kubehook:${VERSION}" .
