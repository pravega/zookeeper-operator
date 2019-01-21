# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

SHELL=/bin/bash -o pipefail

PROJECT_NAME=zookeeper-operator
REPO=pravega/$(PROJECT_NAME)
ALTREPO=emccorp/$(PROJECT_NAME)
VERSION=$(shell git describe --always --tags --dirty | sed "s/\(.*\)-g`git rev-parse --short HEAD`/\1/")
GIT_SHA=$(shell git rev-parse --short HEAD)
GOOS=linux
GOARCH=amd64

.PHONY: all build check clean test

all: dep check build

dep:
	 dep ensure -v

build: test build-go build-image

build-go:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
	-o bin/$(PROJECT_NAME) cmd/manager/main.go

build-image:
	docker build --build-arg VERSION=$(VERSION) --build-arg GIT_SHA=$(GIT_SHA) -t $(REPO):$(VERSION) .
	docker tag $(REPO):$(VERSION) $(REPO):latest

test:
	go test $$(go list ./... | grep -v /vendor/)

login:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"

push: build-image login
	docker push $(REPO):latest
	docker push $(REPO):$(VERSION)
	docker tag $(REPO):$(VERSION) $(ALTREPO):$(VERSION)
	docker push $(ALTREPO):$(VERSION)

clean:
	rm -f bin/$(PROJECT_NAME)

check: check-format check-license

check-format:
	./scripts/check_format.sh

check-license:
	./scripts/check_license.sh
