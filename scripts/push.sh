#!/usr/bin/env sh

set -e

# Create the docker image
VERSION=$(git rev-parse --short HEAD)
docker push "negz/kubehook:latest"
docker push "negz/kubehook:${VERSION}"

