#!/bin/bash
echo "Pushing to dockerhub"

docker tag dotjson/ion-dispatcher:latest dotjson/ion-dispatcher:v0.1.$TRAVIS_BUILD_NUMBER
docker tag dotjson/ion-sidecar:latest dotjson/ion-sidecar:v0.1.$TRAVIS_BUILD_NUMBER 

echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USER" --password-stdin
docker push dotjson/ion-sidecar:v0.1.$TRAVIS_BUILD_NUMBER 
docker push dotjson/ion-sidecar
docker push dotjson/ion-dispatcher:v0.1.$TRAVIS_BUILD_NUMBER
docker push dotjson/ion-dispatcher