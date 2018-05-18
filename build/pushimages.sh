#!/bin/bash
echo "Pushing to dockerhub"

docker tag ion-dispatcher:latest dotjson/ion-dispatcher:v0.1.$TRAVIS_BUILD_NUMBER
docker tag ion-dispatcher:latest dotjson/ion-dispatcher:latest
docker tag ion-sidecar:latest dotjson/ion-sidecar:v0.1.$TRAVIS_BUILD_NUMBER 
docker tag ion-sidecar:latest dotjson/ion-sidecar:latest 

echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USER" --password-stdin
docker push dotjson/ion-sidecar:v0.1.$TRAVIS_BUILD_NUMBER 
docker push dotjson/ion-sidecar
docker push dotjson/ion-dispatcher:v0.1.$TRAVIS_BUILD_NUMBER
docker push dotjson/ion-dispatcher
