#!/bin/bash
echo "Pushing to dockerhub"

if [ -z "$DOCKER_PASSWORD" ]
then
      echo "Skipping docker login as DOCKER_PASSWORD not specified"
else
    echo "Logging into docker..."
    echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USER" --password-stdin
fi


if [ -z "$TRAVIS_BUILD_NUMBER" ]
then
      echo "Using epoc build number as not in travis"
      TRAVIS_BUILD_NUMBER=$(date +%s%3N)
fi

echo "Tagging builds using account:"
echo $DOCKER_USER

docker tag ion-dispatcher:latest $DOCKER_USER/ion-dispatcher:v0.1.$TRAVIS_BUILD_NUMBER
docker tag ion-dispatcher:latest $DOCKER_USER/ion-dispatcher:latest
docker tag ion-handler:latest $DOCKER_USER/ion-handler:v0.1.$TRAVIS_BUILD_NUMBER
docker tag ion-handler:latest $DOCKER_USER/ion-handler:latest
docker tag ion-management:latest $DOCKER_USER/ion-management:v0.1.$TRAVIS_BUILD_NUMBER
docker tag ion-management:latest $DOCKER_USER/ion-management:latest

docker push $DOCKER_USER/ion-handler:v0.1.$TRAVIS_BUILD_NUMBER
docker push $DOCKER_USER/ion-handler
docker push $DOCKER_USER/ion-dispatcher:v0.1.$TRAVIS_BUILD_NUMBER
docker push $DOCKER_USER/ion-dispatcher
docker push $DOCKER_USER/ion-management:v0.1.$TRAVIS_BUILD_NUMBER
docker push $DOCKER_USER/ion-management
