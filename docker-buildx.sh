#!/bin/bash

# docker buildx create --name multiarch-builder --use
# docker buildx inspect --bootstrap

docker buildx build \
    $(cat .env | sed 's/^/--build-arg /') \
    --platform linux/arm64 \
    -t rupeshtr/proxyserver-arm64:v02 \
    --push \
    .
