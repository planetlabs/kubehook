#!/usr/bin/env bash

set -e

ENABLE_TEMPLATE=$1
VERSION=$(git rev-parse --short HEAD)
NAME=kubehook

CFG=$(mktemp -d /tmp/${NAME}.XXXX)
cat <<EOF >$CFG/template
apiVersion: v1
kind: Config
clusters:
- name: kuberos
  cluster:
    certificate-authority-data: REDACTED
    server: https://kuberos.example.org
EOF

docker kill ${NAME} || true
docker rm ${NAME} || true

KUBEHOOK_ARGS=""
if [[ $ENABLE_TEMPLATE == "true" ]]; then
	KUBEHOOK_ARGS="--kubecfg-template /cfg/template"
fi

docker run -d \
	--name ${NAME} \
	-p 10003:10003 \
	-v $CFG:/cfg \
	-e "KUBEHOOK_SECRET=secret" \
	"negz/kubehook:${VERSION}" /kubehook $KUBEHOOK_ARGS
