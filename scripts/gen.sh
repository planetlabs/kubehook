#!/usr/bin/env sh

set -e

ENDPOINT=$1
USER=$2

curl -i -X POST \
	-H "Content-Type: application/json" \
	-H "X-Forwarded-User: ${USER}" \
	-d "{\"lifetime\": \"24h\"}" \
	"${ENDPOINT}"

echo
