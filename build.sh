#!/usr/bin/env bash

NAME="stockholm_commute_bot_app"
TAG=latest
USER="kgantsov"
DOCKER_ID_USER="kgantsov"

cd cmd/bot && GOOS=linux GIN_MODE=release go build && cd ../..

docker build -t $USER/$NAME:$TAG --no-cache .

docker tag $USER/$NAME:$TAG $USER/$NAME:$TAG
docker push $USER/$NAME:$TAG


docker rmi $USER/$NAME:$TAG

rm ./cmd/bot/bot
