#/bin/bash -e

[ -z "${DOCKER_HOST}" ] && echo "DOCKER_HOST must be set" && exit 1

DYNAMO_PORT=10005
DYNAMO_HOST=$(echo ${DOCKER_HOST##*/}|sed 's/:.*//')  # proto://127.0.0.1:8080 -> 127.0.0.1
DYNAMO_ENDPOINT="http://${DYNAMO_HOST}:${DYNAMO_PORT}"

docker run -d -p ${DYNAMO_PORT}:8000 "behumble/dynamodb-local" &>/dev/null

aws --endpoint-url ${DYNAMO_ENDPOINT} dynamodb create-table \
    --table-name dev-kubehook-users \
    --attribute-definitions \
        AttributeName=token,AttributeType=S \
    --key-schema \
        AttributeName=token,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 &>/dev/null

echo $DYNAMO_ENDPOINT
