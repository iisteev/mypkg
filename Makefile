PROJECT_NAME := "mypkg"
PKG := "mypkg"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)
VERSION := "$(shell cat VERSION):$(shell git rev-parse HEAD)"
BUILD_DIR := $(shell pwd)/dist
GOBIN := $(shell go env GOPATH)/bin

.PHONY: help

.DEFAULT_GOAL := help

all: build

lint: dep ## Lint the files
	$(GOBIN)/golint ${PKG_LIST}

test: dep ## Run the unit tests
	go test -short ${PKG_LIST}

dep: ## Get the dependencies
	go get -v -d ./...

build-dev: dest-dir dep ## Build the binary file (development)
	go build -v -o $(BUILD_DIR)/$(PKG) main.go

build: dest-dir dep ## Build the binary file for release
	VERSION=$(VERSION) BUILD_DIR=$(BUILD_DIR) ./build.sh $(MYPKG_BUILD_OS)

clean: ## Remove old build
	go clean
	rm -rf $(BUILD_DIR)

dest-dir: ## Create the build dir
	mkdir -p $(BUILD_DIR)

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

