# If you update this file, please follow
# https://suva.sh/posts/well-documented-makefiles

# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL:=/usr/bin/env bash

.DEFAULT_GOAL := help

# Code
VERSION := 0.0.1-q8s
GIT_SHA := $(shell git rev-parse --short HEAD)
PROJECT_NAME := zookeeper-operator
CODE_PATH := github.com/q8s-io/zookeeper-operator-pravega

# Define Docker related variables. Releases should modify and double check these vars.
REGISTRY := uhub.service.ucloud.cn/infra
IMAGE := zookeeper-operator
CONTROLLER_IMG := $(REGISTRY)/$(IMAGE)

# Tools
TOOLS_DIR := tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://goproxy.cn
endif
export GOPROXY
# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on
# Lint
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint

.PHONY: server

run:
	GOPROXY=https://goproxy.cn GO111MODULE=on go run detector.go -conf "./configs/pro.toml"

server:
	@echo "version: $(VERSION)"
	docker build --no-cache --build-arg VERSION=$(VERSION) --build-arg GIT_SHA=$(GIT_SHA) --build-arg PROJECT_NAME=$(PROJECT_NAME) --build-arg CODE_PATH=$(CODE_PATH) -t $(CONTROLLER_IMG):$(VERSION) -f Dockerfile .
	docker push $(CONTROLLER_IMG):$(VERSION)
