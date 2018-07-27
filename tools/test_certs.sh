#!/bin/bash

set -e

OUT_DIR=$1      # Directory to write the new certificates to
TOOL=$2         # The certificate generation tool
DNS_NAME=$3     # The DNS name to use to connect the client and server
IMAGE_TAG=$4    # Optional Docker image tag used in terraform

if pgrep management; then pkill -9 management &> /dev/null; fi
if pgrep ioncli; then pkill -9 ioncli &> /dev/null; fi

echo "------------------------------------"
echo "Generating new certificates using: $TOOL"
echo "------------------------------------"
case "$TOOL" in
    "openssl")
        if [ -z "$DNS_NAME" ];then
            DNS_NAME="localhost"
        fi
        echo "I will generate new certificates"
        echo "I will execute a new deployment. Please be patient"
        echo "this can take up to 30 minutes."
        ./tools/generate_certs_openssl.sh "$OUT_DIR" "$DNS_NAME"
        STATUS=$?
        if [ $STATUS != 0 ]; then
            exit $STATUS
        fi

        echo "------------------------------------"
        echo "Building and running management api with new server certificates"
        echo "------------------------------------"
        export $(cat ./.vscode/private.mgmt.env | xargs) && \
        go build ./cmd/management && \
        ./management start \
        --certfile="$OUT_DIR/server.crt" \
        --keyfile="$OUT_DIR/server.key" \
        --cacertfile="$OUT_DIR/rootCA.pem" \
        --hostname="$DNS_NAME" &

        sleep 8 # Give time to start server
    ;;
    "terraform")
        echo "If no existing terraform.tfstate file is present,"
        echo "I will execute a new deployment. Please be patient"
        echo "this can take up to 30 minutes."
        RES=$(./tools/generate_certs_terraform.sh "$OUT_DIR" "$IMAGE_TAG")
        STATUS=$?
        if [ $STATUS != 0 ]; then
            echo $RES
            exit $STATUS
        fi
        DNS_NAME=$RES
    ;;
    *)
        echo "Unrecognized tool choice:  $TOOL"
        echo "Options are: openssl, terraform"
        exit
    ;;
esac

echo
echo "------------------------------------"
echo "Building ion cli and running with new client certificates"
echo "------------------------------------"
go build -o ioncli ./cmd/ion && \
./ioncli module list \
--certfile="$OUT_DIR/client.crt" \
--keyfile="$OUT_DIR/client.key" \
--cacertfile="$OUT_DIR/rootCA.pem" \
--endpoint="$DNS_NAME:9000" --timeout 5

rm ioncli > /dev/null 2>&1
rm management > /dev/null 2>&1