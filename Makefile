NAME          := gitlab-ci-pipelines-exporter
VERSION       := $(shell git describe --tags --abbrev=1)
FILES         := $(shell git ls-files '*.go')
REPOSITORY    := mvisonneau/$(NAME)
.DEFAULT_GOAL := help

export GO111MODULE=on

.PHONY: setup
setup: ## Install required libraries/tools for build tasks
	@command -v cover 2>&1 >/dev/null       || GO111MODULE=off go get -u -v golang.org/x/tools/cmd/cover
	@command -v goimports 2>&1 >/dev/null   || GO111MODULE=off go get -u -v golang.org/x/tools/cmd/goimports
	@command -v goveralls 2>&1 >/dev/null   || GO111MODULE=off go get -u -v github.com/mattn/goveralls
	@command -v ineffassign 2>&1 >/dev/null || GO111MODULE=off go get -u -v github.com/gordonklaus/ineffassign
	@command -v misspell 2>&1 >/dev/null    || GO111MODULE=off go get -u -v github.com/client9/misspell/cmd/misspell
	@command -v revive 2>&1 >/dev/null      || GO111MODULE=off go get -u -v github.com/mgechev/revive

.PHONY: fmt
fmt: setup ## Format source code
	goimports -w $(FILES)

.PHONY: lint
lint: revive vet goimports ineffassign misspell ## Run all lint related tests against the codebase

.PHONY: revive
revive: setup ## Test code syntax with revive
	revive -config .revive.toml $(FILES)

.PHONY: vet
vet: ## Test code syntax with go vet
	go vet ./...

.PHONY: goimports
goimports: setup ## Test code syntax with goimports
	goimports -d $(FILES) > goimports.out
	@if [ -s goimports.out ]; then cat goimports.out; rm goimports.out; exit 1; else rm goimports.out; fi

.PHONY: ineffassign
ineffassign: setup ## Test code syntax for ineffassign
	ineffassign $(FILES)

.PHONY: misspell
misspell: setup ## Test code with misspell
	misspell -error $(FILES)

.PHONY: test
test: ## Run the tests against the codebase
	go test -v -race ./...

.PHONY: install
install: ## Build and install locally the binary (dev purpose)
	go install ./cmd/$(NAME)

.PHONY: build-local
build-local: ## Build the binaries using local GOOS
	go build ./cmd/$(NAME)

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
		golang:1.14 \
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
