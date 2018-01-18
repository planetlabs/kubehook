#!/usr/bin/env bash

set -e

ENDPOINT=$1
USER=$2

curl -i -X GET \
	-H "X-Forwarded-User: ${USER}" \
	"${ENDPOINT}?lifetime=24h"

echo
