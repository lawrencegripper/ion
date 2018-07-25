#!/bin/bash

set -e

TERRAFORM_INSTALLED=$(command -v terraform)
if [ -z "$TERRAFORM_INSTALLED" ]; then
    echo "terraform is not installed, please install it in order to run this script"
    exit
fi

OUT_DIR=$1
if [ -z "$OUT_DIR" ]; then
    OUT_DIR=$PWD
    echo "No directory passed, using current directory ("$OUT_DIR"), continue (y/n)"
    read res
    if [[ "$res" != "y" || "$res" != "Y" ]]; then
        exit
    fi
fi

mkdir -p $OUT_DIR

ION_IMAGE_TAG=$2
if [ -z "$ION_IMAGE_TAG" ]; then
    echo "No image tag provided, defaulting to latest"
    ION_IMAGE_TAG="latest"
fi

cd ./deployment
if [ ! -f ./vars.private.tfvars ]; then
    echo "vars.private.tfvars not found in deployment file!"
    echo "WARNING.... you'll need to create it some of the fields in ./deployment/vars.private.tfvars without it the terraform deployment will fail"
    exit
fi

terraform init
terraform apply -var-file ./vars.private.tfvars -auto-approve -var docker_root=$DOCKER_USER -var docker_tag=$ION_IMAGE_TAG
terraform output client_cert > "$OUT_DIR/client.crt"
terraform output client_key > "$OUT_DIR/client.key"
terraform output cluster_ca > "$OUT_DIR/rootCA.pem"
terraform output server_cert > "$OUT_DIR/server.crt"
terraform output server_key > "$OUT_DIR/server.key"