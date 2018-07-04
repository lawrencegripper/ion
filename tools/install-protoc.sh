#!/bin/sh -eu
protoc_version=$1
if test -z "${protoc_version}"; then
	echo "Usage: .travis/install-protoc.sh protoc-version"
	exit 1
fi

wget https://github.com/google/protobuf/releases/download/v${protoc_version}/protoc-${protoc_version}-linux-x86_64.zip
unzip protoc-${protoc_version}-linux-x86_64.zip
mv bin/protoc /usr/local/bin
chmod +x /usr/local/bin/protoc

protoc --version