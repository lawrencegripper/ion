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

IMAGES=$(docker image ls | grep ion | grep -v / | awk '{print $1}')

echo "Tagging images"
echo "$IMAGES" | xargs -I % echo "Tagged $DOCKER_USER/%:latest"
echo "$IMAGES" | xargs -I % docker tag % $DOCKER_USER/%:latest
echo "$IMAGES" | xargs -I % echo "Tagged $DOCKER_USER/%:v0.1.$TRAVIS_BUILD_NUMBER"
echo "$IMAGES" | xargs -I % docker tag % $DOCKER_USER/%:v0.1.$TRAVIS_BUILD_NUMBER

echo "Pushing images"
echo "$IMAGES" | xargs -I % docker push $DOCKER_USER/%:latest 
echo "$IMAGES" | xargs -I % docker push $DOCKER_USER/%:v0.1.$TRAVIS_BUILD_NUMBER
