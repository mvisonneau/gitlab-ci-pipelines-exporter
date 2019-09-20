NAME          := gitlab-ci-pipelines-exporter
VERSION       := $(shell git describe --tags --abbrev=1)
FILES         := $(shell git ls-files '*.go')
REPOSITORY    := mvisonneau/$(NAME)
.DEFAULT_GOAL := help

export GO111MODULE=on

.PHONY: setup
setup: ## Install required libraries/tools for build tasks
	@command -v goveralls 2>&1 >/dev/null || GO111MODULE=off go get -u -v github.com/mattn/goveralls
	@command -v golint 2>&1 >/dev/null    || GO111MODULE=off go get -u -v golang.org/x/lint/golint
	@command -v cover 2>&1 >/dev/null     || GO111MODULE=off go get -u -v golang.org/x/tools/cmd/cover
	@command -v goimports 2>&1 >/dev/null || GO111MODULE=off go get -u -v golang.org/x/tools/cmd/goimports

.PHONY: fmt
fmt: setup ## Format source code
	gofmt -s -w $(FILES)
	goimports -w $(FILES)

.PHONY: lint
lint: setup ## Run golint, goimports and go vet against the codebase
	golint -set_exit_status .
	go vet ./...
	goimports -d $(FILES) > goimports.out
	@if [ -s goimports.out ]; then cat goimports.out; rm goimports.out; exit 1; else rm goimports.out; fi

.PHONY: test
test: ## Run the tests against the codebase
	go test -v ./...

.PHONY: install
install: ## Build and install locally the binary (dev purpose)
	go build .

.PHONY: build
build: ## Build the binaries
	goreleaser release --snapshot --skip-publish --rm-dist

.PHONY: build-linux-amd64
build-linux-amd64: ## Build the binaries
	goreleaser release --snapshot --skip-publish --rm-dist -f .goreleaser.linux-amd64.yml

.PHONY: release
release: ## Build & release the binaries
	goreleaser release --rm-dist

.PHONY: publish-coveralls
publish-coveralls: setup ## Publish coverage results on coveralls
	goveralls -service drone.io -coverprofile=coverage.out

.PHONY: clean
clean: ## Remove binary if it exists
	rm -f $(NAME)

.PHONY: coverage
coverage: ## Generates coverage report
	rm -rf *.out
	go test -v ./... -coverpkg=./... -coverprofile=coverage.out

.PHONY: dev-env
dev-env: ## Build a local development environment using Docker
	@docker run -it --rm \
		-v $(shell pwd):/go/src/github.com/mvisonneau/$(NAME) \
		-w /go/src/github.com/mvisonneau/$(NAME) \
		-p 8080:8080 \
		golang:1.13 \
		/bin/bash -c 'make setup; make install; bash'

.PHONY: is-git-dirty
is-git-dirty: ## Tests if git is in a dirty state
	@test $(shell git status --porcelain | grep -c .) -eq 0

.PHONY: sign-drone
sign-drone: ## Sign Drone CI configuration
	drone sign $(REPOSITORY) --save

.PHONY: all
all: lint test build coverage ## Test, builds and ship package for all supported platforms

.PHONY: help
help: ## Displays this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
