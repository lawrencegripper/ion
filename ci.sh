#!/bin/bash

docker build --build-arg FOLDER=dispatcher -f ci.Dockerfile . 
docker build --build-arg FOLDER=sidecar -f ci.Dockerfile . 
