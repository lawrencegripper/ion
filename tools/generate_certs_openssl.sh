#!/bin/bash

set -e

OPENSSL_INSTALLED=$(command -v openssl)
if [ -z "$OPENSSL_INSTALLED" ]; then
    echo "openssl is not installed, please install it in order to run this script"
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

CERT_SUBJ="/CN=ion.local/O=Ion LTD./C=GB/ST=London/L=London/O=Ion/OU=CA Dep"

echo "Generating root certificate authority"
echo "------------------------------------"
openssl genrsa -out "$OUT_DIR/rootCA.key" 2048
openssl genrsa -des3 -out "$OUT_DIR/rootCA.key" 2048
openssl req -x509 -new -nodes -key "$OUT_DIR/rootCA.key" -subj "$CERT_SUBJ" -sha256 -days 1024 -out "$OUT_DIR/rootCA.pem"

echo "Generating server certificates"
echo "------------------------------------"
openssl genrsa -out "$OUT_DIR/server.key" 2048
openssl req -new -key "$OUT_DIR/server.key" -out "$OUT_DIR/server.csr" -subj "$CERT_SUBJ"
openssl x509 -req -in "$OUT_DIR/server.csr" -CA "$OUT_DIR/rootCA.pem" -CAkey "$OUT_DIR/rootCA.key" -CAcreateserial -out "$OUT_DIR/server.crt" -days 500 -sha256

echo "Generating client certificates"
echo "------------------------------------"
openssl genrsa -out "$OUT_DIR/client.key" 2048
openssl req -new -key "$OUT_DIR/client.key" -out "$OUT_DIR/client.csr" -subj "$CERT_SUBJ"
openssl x509 -req -in "$OUT_DIR/client.csr" -CA "$OUT_DIR/rootCA.pem" -CAkey "$OUT_DIR/rootCA.key" -CAcreateserial -out "$OUT_DIR/client.crt" -days 500 -sha256