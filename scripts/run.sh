#!/usr/bin/env sh

set -e

VERSION=$(git rev-parse --short HEAD)
NAME=kubehook

docker kill ${NAME} || true
docker rm ${NAME} || true

docker run -d \
	--name ${NAME} \
	-p 10003:10003 \
	-e "KUBEHOOK_SECRET=secret" \
	"negz/kubehook:${VERSION}" /kubehook
