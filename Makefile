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
KUBE_VERSION=1.17.5
REPO=pravega/$(PROJECT_NAME)
TEST_REPO=testzkop/$(PROJECT_NAME)
APP_REPO=pravega/$(APP_NAME)
ALTREPO=emccorp/$(PROJECT_NAME)
APP_ALTREPO=emccorp/$(APP_NAME)
VERSION=$(shell git describe --always --tags --dirty | sed "s/\(.*\)-g`git rev-parse --short HEAD`/\1/")
GIT_SHA=$(shell git rev-parse --short HEAD)
TEST_IMAGE=$(TEST_REPO)-testimages:$(VERSION)
DOCKER_TEST_PASS=testzkop@123
DOCKER_TEST_USER=testzkop

.PHONY: all build check clean test

all: check build

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
	docker build --build-arg VERSION=$(VERSION) --build-arg GIT_SHA=$(GIT_SHA) -t $(REPO):$(VERSION) .
	docker tag $(REPO):$(VERSION) $(REPO):latest

build-zk-image:
	docker build --build-arg VERSION=$(VERSION) --build-arg GIT_SHA=$(GIT_SHA) -t $(APP_REPO):$(VERSION) ./docker
	docker tag $(APP_REPO):$(VERSION) $(APP_REPO):latest

build-zk-image-swarm:
	docker build --build-arg VERSION=$(VERSION)-swarm --build-arg GIT_SHA=$(GIT_SHA) -f ./docker/Dockerfile-swarm -t $(APP_REPO):$(VERSION)-swarm ./docker

test:
	go test $$(go list ./... | grep -v /vendor/ | grep -v /test/e2e) -race -coverprofile=coverage.txt -covermode=atomic

test-e2e: test-e2e-remote

test-e2e-remote: test-login
	operator-sdk build $(TEST_IMAGE)
	docker push $(TEST_IMAGE)
	operator-sdk test local ./test/e2e --operator-namespace default \
		--namespaced-manifest ./test/e2e/resources/rbac-operator.yaml \
		--global-manifest deploy/crds/zookeeper.pravega.io_zookeeperclusters.yaml \
		--image $(TEST_IMAGE) --go-test-flags "-v -timeout 0"

test-e2e-local:
	operator-sdk test local ./test/e2e --operator-namespace default --up-local --go-test-flags "-v -timeout 0"

run-local:
	operator-sdk up local

login:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"

test-login:
	echo "$(DOCKER_TEST_PASS)" | docker login -u "$(DOCKER_TEST_USER)" --password-stdin

push: build-image build-zk-image build-zk-image-swarm login
	docker push $(REPO):$(VERSION)
	docker push $(REPO):latest
	docker push $(APP_REPO):$(VERSION)
	docker push $(APP_REPO):latest
	docker push $(APP_REPO):$(VERSION)-swarm
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

manifests: controller-gen
	$(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=deploy/crds
	$(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=charts/zookeeper-operator/crds

controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOPATH)/bin/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif
