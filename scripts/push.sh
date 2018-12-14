#!/usr/bin/env bash

set -e

# Create the docker image
VERSION=$(git rev-parse --short HEAD)
docker push "planetlabs/kubehook:latest"
docker push "planetlabs/kubehook:${VERSION}"
