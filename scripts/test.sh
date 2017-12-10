#!/usr/bin/env sh

set -e

VENDOR="vendor"

[ -d "${VENDOR}" ] && rm -rf "${VENDOR}"

# Ideally we'd use a specific version of Glide.
go get -u github.com/Masterminds/glide

# Create the vendor directory based on glide.lock
glide install

# Test!
go test -race -cover $(glide nv)
