#!/bin/bash
echo "Pushing to dockerhub"

if [ -z "$DOCKER_USER" ]
then
      echo "Failed: must have $DOCKER_USER set to a docker user you have push rights for.... eg: 'export DOCKER_USER=lawrencegripper'"
fi

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

export ION_IMAGE_TAG=v0.1.$TRAVIS_BUILD_NUMBER

echo "Tag for images exported to $ION_IMAGE_TAG"
echo $ION_IMAGE_TAG > imagetag.temp