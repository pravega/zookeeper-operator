# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

SHELL=/bin/bash -o pipefail
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

PROJECT_NAME=zookeeper-operator
EXPORTER_NAME=zookeeper-exporter
APP_NAME=zookeeper
KUBE_VERSION=1.20.13
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
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

crds: ## Generate CRDs
	- make controller-gen
	- $(CONTROLLER_GEN) crd paths=./api/... output:dir=./config/crd/bases schemapatch:manifests=./config/crd/bases


# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image pravega/zookeeper-operator=$(TEST_IMAGE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -


# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy-test: manifests kustomize
	cd config/test
	$(KUSTOMIZE) build config/test | kubectl apply -f -

# Undeploy controller in the configured Kubernetes cluster in ~/.kube/config
undeploy-test: manifests kustomize
	cd config/test
	$(KUSTOMIZE) build config/test | kubectl apply -f -

# Undeploy controller in the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

all: generate check build

generate:
	make controller-gen
	$(CONTROLLER_GEN) object paths="./..."
	$(CONTROLLER_GEN) crd paths=./api/... output:dir=./config/crd/bases schemapatch:manifests=./config/crd/bases
	# sync crd generated to helm-chart
	echo '{{- define "crd.openAPIV3Schema" }}' > charts/zookeeper-operator/templates/_crd_openapiv3schema.tpl
	echo 'openAPIV3Schema:' >> charts/zookeeper-operator/templates/_crd_openapiv3schema.tpl
	sed -e '1,/openAPIV3Schema/d' config/crd/bases/zookeeper.pravega.io_zookeeperclusters.yaml | sed -n '/served: true/!p;//q' >> charts/zookeeper-operator/templates/_crd_openapiv3schema.tpl
	echo '{{- end }}' >> charts/zookeeper-operator/templates/_crd_openapiv3schema.tpl


build: test build-go build-image

build-go:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(PROJECT_NAME)-linux-amd64 main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(EXPORTER_NAME)-linux-amd64 cmd/exporter/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(PROJECT_NAME)-darwin-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(EXPORTER_NAME)-darwin-amd64 cmd/exporter/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
		-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
		-o bin/$(PROJECT_NAME)-windows-amd64.exe main.go
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

test-e2e-remote:
	make test-login
	docker build . -t $(TEST_IMAGE)
	docker push $(TEST_IMAGE)
	make deploy
	RUN_LOCAL=false go test -v -timeout 2h ./test/e2e... -args -ginkgo.v
	make undeploy

test-e2e-local:
	make deploy-test
	RUN_LOCAL=true go test -v -timeout 2h ./test/e2e... -args -ginkgo.v
	make undeploy-test

run-local:
	go run ./main.go

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
