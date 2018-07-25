#!/bin/bash

set -e

OUT_DIR=$1      # Directory to write the new certificates to
TOOL=$2         # The certificate generation tool
IMAGE_TAG=$3    # The docker image tag to use

ps aux | grep management | awk '{print $2}' | xargs kill || true

# Clean up will prompt the user to clean
# up any resources created during the script
function cleanUp {
    # echo "Ion management API running as PID: $MGMT_PID"
    # echo "want me to kill it? (y/n)"
    # read killIt
    # if [[ "$killIt" == "y" || "$killIt" == "Y" ]]; then
        kill $MGMT_PID || true
    # fi

    # echo "This script generated the following files:"
    # echo " - management"
    # echo " - ioncli"
    # echo "Do you want me to remove them? (y/n)"
    # echo "want me to kill it? (y/n)"
    # read deleteThem
    # if [[ "$deleteThem" == "y" || "$deleteThem" == "Y" ]]; then
        rm management || true
        rm ioncli || true
    # fi
}

trap cleanUp EXIT

echo "------------------------------------"
echo "Generating new certificates using: $TOOL"
echo "------------------------------------"
case "$TOOL" in
     "openssl")
          ./tools/generate_certs_openssl.sh $OUT_DIR
          ;;
     "terraform")
          ./tools/generate_certs_terraform.sh $OUT_DIR $IMAGE_TAG
          ;;
     *)
          echo "Unrecognized tool choice:  $TOOL"
          echo "Options are: openssl, terraform"
          exit
          ;;
esac

echo "$OUT_DIR/rootCA.pem"

echo "------------------------------------"
echo "Building and running management api with new server certificates"
echo "------------------------------------"
export $(cat ./.vscode/private.mgmt.env | xargs) && go build ./cmd/management && ./management start --certfile="$OUT_DIR/server.crt" --keyfile="$OUT_DIR/server.key" --cacertfile="$OUT_DIR/rootCA.pem" &
MGMT_PID=$!

sleep 5 # Give time to start server

echo
echo "------------------------------------"
echo "Building ion cli and running with new client certificates"
echo "------------------------------------"
go build -o ioncli ./cmd/ion && ./ioncli module list --certfile="$OUT_DIR/client.crt" --keyfile="$OUT_DIR/client.key" --cacertfile="$OUT_DIR/rootCA.pem" --endpoint="ion.local:9000" --timeout 5