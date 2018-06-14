#!/bin/bash
echo "Pushing to dockerhub"

docker tag ion-dispatcher:latest dotjson/ion-dispatcher:v0.1.$TRAVIS_BUILD_NUMBER
docker tag ion-dispatcher:latest dotjson/ion-dispatcher:latest
docker tag ion-handler:latest dotjson/ion-handler:v0.1.$TRAVIS_BUILD_NUMBER 
docker tag ion-handler:latest dotjson/ion-handler:latest 

echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USER" --password-stdin
docker push dotjson/ion-handler:v0.1.$TRAVIS_BUILD_NUMBER 
docker push dotjson/ion-handler
docker push dotjson/ion-dispatcher:v0.1.$TRAVIS_BUILD_NUMBER
docker push dotjson/ion-dispatcher
