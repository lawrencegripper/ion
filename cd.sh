#!/bin/bash
echo "Building docker and pushing"

docker build -t dotjson/ion-dispatcher:latest -t dotjson/ion-dispatcher:v0.1.$TRAVIS_BUILD_NUMBER -f ./dispatcher/Dockerfile .
docker build -t dotjson/ion-sidecar:latest -t dotjson/ion-sidecar:v0.1.$TRAVIS_BUILD_NUMBER -f ./sidecar/Dockerfile .

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push dotjson/ion-sidecar
docker push dotjson/ion-dispatcher



