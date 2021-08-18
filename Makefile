# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

SHELL=/bin/bash -o pipefail

PROJECT_NAME=zookeeper-operator
EXPORTER_NAME=zookeeper-exporter
APP_NAME=zookeeper
KUBE_VERSION=1.19.13
REPO=pravega/$(PROJECT_NAME)
TEST_REPO=testzkop/$(PROJECT_NAME)
APP_REPO=pravega/$(APP_NAME)
ALTREPO=emccorp/$(PROJECT_NAME)
APP_ALTREPO=emccorp/$(APP_NAME)
OSDK_VERSION=$(shell operator-sdk version | cut -f2 -d'"')
VERSION=$(shell git describe --always --tags --dirty | tr -d "v" | sed "s/\(.*\)-g`git rev-parse --short HEAD`/\1/")
GIT_SHA=$(shell git rev-parse --short HEAD)
TEST_IMAGE=$(TEST_REPO)-testimages:$(VERSION)
DOCKER_TEST_PASS=testzkop@123
DOCKER_TEST_USER=testzkop
.PHONY: all build check clean test

all: generate check build

generate:
	[[ ${OSDK_VERSION} == v0.19* ]] || ( echo "operator-sdk version 0.19 required" ; exit 1 )
	operator-sdk generate crds --crd-version v1
	env GOROOT=$(shell go env GOROOT) operator-sdk generate k8s
	# sync crd generated to helm-chart
	echo '{{- define "crd.openAPIV3Schema" }}' > charts/zookeeper-operator/templates/_crd_openapiv3schema.tpl
	echo 'openAPIV3Schema:' >> charts/zookeeper-operator/templates/_crd_openapiv3schema.tpl
	sed -e '1,/openAPIV3Schema/d' deploy/crds/zookeeper.pravega.io_zookeeperclusters_crd.yaml | sed -n '/served: true/!p;//q' >> charts/zookeeper-operator/templates/_crd_openapiv3schema.tpl
	echo '{{- end }}' >> charts/zookeeper-operator/templates/_crd_openapiv3schema.tpl


build: test build-go build-image

build-go:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(PROJECT_NAME)-linux-amd64 cmd/manager/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(EXPORTER_NAME)-linux-amd64 cmd/exporter/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(PROJECT_NAME)-darwin-amd64 cmd/manager/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(EXPORTER_NAME)-darwin-amd64 cmd/exporter/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(PROJECT_NAME)-windows-amd64.exe cmd/manager/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(EXPORTER_NAME)-windows-amd64.exe cmd/exporter/main.go

build-image:
	docker build --build-arg VERSION=$(VERSION) --build-arg DOCKER_REGISTRY=$(DOCKER_REGISTRY) --build-arg GIT_SHA=$(GIT_SHA) -t $(REPO):$(VERSION) .
	docker tag $(REPO):$(VERSION) $(REPO):latest

build-zk-image:
	docker build --build-arg VERSION=$(VERSION)  --build-arg DOCKER_REGISTRY=$(DOCKER_REGISTRY) --build-arg GIT_SHA=$(GIT_SHA) -t $(APP_REPO):$(VERSION) ./docker
	docker tag $(APP_REPO):$(VERSION) $(APP_REPO):latest

build-zk-image-swarm:
	docker build --build-arg VERSION=$(VERSION)-swarm  --build-arg DOCKER_REGISTRY=$(DOCKER_REGISTRY) --build-arg GIT_SHA=$(GIT_SHA) \
		-f ./docker/Dockerfile-swarm -t $(APP_REPO):$(VERSION)-swarm ./docker

test:
	go test $$(go list ./... | grep -v /vendor/ | grep -v /test/e2e) -race -coverprofile=coverage.txt -covermode=atomic

test-e2e: test-e2e-remote

test-e2e-remote: test-login
	operator-sdk build $(TEST_IMAGE)
	docker push $(TEST_IMAGE)
	operator-sdk test local ./test/e2e --operator-namespace default \
		--namespaced-manifest ./test/e2e/resources/rbac-operator.yaml \
		--global-manifest deploy/crds/zookeeper.pravega.io_zookeeperclusters_crd.yaml \
		--image $(TEST_IMAGE) --go-test-flags "-v -timeout 0"

test-e2e-local:
	operator-sdk test local ./test/e2e --operator-namespace default --up-local --go-test-flags "-v -timeout 0"

run-local:
	operator-sdk run local

login:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"

test-login:
	echo "$(DOCKER_TEST_PASS)" | docker login -u "$(DOCKER_TEST_USER)" --password-stdin

push: build-image build-zk-image login
	docker push $(REPO):$(VERSION)
	docker push $(REPO):latest
	docker push $(APP_REPO):$(VERSION)
	docker push $(APP_REPO):latest
	docker tag $(REPO):$(VERSION) $(ALTREPO):$(VERSION)
	docker tag $(REPO):$(VERSION) $(ALTREPO):latest
	docker tag $(APP_REPO):$(VERSION) $(APP_ALTREPO):$(VERSION)
	docker tag $(APP_REPO):$(VERSION) $(APP_ALTREPO):latest
	docker push $(ALTREPO):$(VERSION)
	docker push $(ALTREPO):latest
	docker push $(APP_ALTREPO):$(VERSION)
	docker push $(APP_ALTREPO):latest

clean:
	rm -f bin/$(PROJECT_NAME)

check: check-format check-license

check-format:
	./scripts/check_format.sh

check-license:
	./scripts/check_license.sh

update-kube-version:
	./scripts/update_kube_version.sh ${KUBE_VERSION}
