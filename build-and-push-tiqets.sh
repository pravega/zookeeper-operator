#!/usr/bin/env bash
set -e -u -o pipefail

BASE=129317782449.dkr.ecr.eu-west-1.amazonaws.com
TAG=0.2.15.1

docker buildx build \
        --push \
		--build-arg VERSION="${TAG}" \
		--platform linux/arm64,linux/amd64 \
		-t "${BASE}/zookeeper-operator":"${TAG}" \
		.

docker buildx build \
        --push \
		--build-arg VERSION="${TAG}" \
		--platform linux/arm64,linux/amd64 \
		-t "${BASE}/zookeeper":"${TAG}" \
		./docker
