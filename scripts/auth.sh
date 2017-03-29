#!/usr/bin/env sh

set -e

ENDPOINT=$1
TOKEN=$2

curl -i -X GET \
    -H "Content-Type: application/json" \
    -d "{\"apiVersion\":\"authentication.k8s.io/v1beta1\",\"kind\":\"TokenReview\",\"spec\":{\"token\":\"${TOKEN}\"}}" \
    "${ENDPOINT}"

echo
