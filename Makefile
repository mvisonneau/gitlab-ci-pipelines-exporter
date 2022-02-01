NAME          := gitlab-ci-pipelines-exporter
FILES         := $(shell git ls-files */*.go)
REPOSITORY    := mvisonneau/$(NAME)
.DEFAULT_GOAL := help

.PHONY: setup
setup: ## Install required libraries/tools for build tasks
	@command -v gofumpt 2>&1 >/dev/null     || go install mvdan.cc/gofumpt@v0.1.1
	@command -v gosec 2>&1 >/dev/null       || go install github.com/securego/gosec/v2/cmd/gosec@v2.8.1
	@command -v ineffassign 2>&1 >/dev/null || go install github.com/gordonklaus/ineffassign@v0.0.0-20210914165742-4cc7213b9bc8
	@command -v misspell 2>&1 >/dev/null    || go install github.com/client9/misspell/cmd/misspell@v0.3.4
	@command -v revive 2>&1 >/dev/null      || go install github.com/mgechev/revive@v1.1.1

.PHONY: fmt
fmt: setup ## Format source code
	gofumpt -w $(FILES)

.PHONY: lint
lint: revive vet gofumpt ineffassign misspell gosec ## Run all lint related tests against the codebase

.PHONY: revive
revive: setup ## Test code syntax with revive
	revive -config .revive.toml $(FILES)

.PHONY: vet
vet: ## Test code syntax with go vet
	go vet ./...

.PHONY: gofumpt
gofumpt: setup ## Test code syntax with gofumpt
	gofumpt -d $(FILES) > gofumpt.out
	@if [ -s gofumpt.out ]; then cat gofumpt.out; rm gofumpt.out; exit 1; else rm gofumpt.out; fi

.PHONY: ineffassign
ineffassign: setup ## Test code syntax for ineffassign
	ineffassign ./...

.PHONY: misspell
misspell: setup ## Test code with misspell
	misspell -error $(FILES)

.PHONY: gosec
gosec: setup ## Test code for security vulnerabilities
	gosec ./...

.PHONY: test
test: ## Run the tests against the codebase
	go test -v -count=1 -race ./...

.PHONY: install
install: ## Build and install locally the binary (dev purpose)
	go install ./cmd/$(NAME)

.PHONY: build
build: ## Build the binaries using local GOOS
	go build ./cmd/$(NAME)

.PHONY: release
release: ## Build & release the binaries (stable)
	git tag -d edge
	goreleaser release --rm-dist
	find dist -type f -name "*.snap" -exec snapcraft upload --release stable,edge '{}' \;

.PHONY: prerelease
prerelease: setup ## Build & prerelease the binaries (edge)
	@\
		REPOSITORY=$(REPOSITORY) \
		NAME=$(NAME) \
		GITHUB_TOKEN=$(GITHUB_TOKEN) \
		.github/prerelease.sh

.PHONY: clean
clean: ## Remove binary if it exists
	rm -f $(NAME)

.PHONY: coverage
coverage: ## Generates coverage report
	rm -rf *.out
	go test -count=1 -race -v ./... -coverpkg=./... -coverprofile=coverage.out

.PHONY: coverage-html
coverage-html: ## Generates coverage report and displays it in the browser
	go tool cover -html=coverage.out

.PHONY: dev-env
dev-env: ## Build a local development environment using Docker
	@docker run -it --rm \
		-v $(shell pwd):/go/src/github.com/mvisonneau/$(NAME) \
		-w /go/src/github.com/mvisonneau/$(NAME) \
		-p 8080:8080 \
		golang:1.17 \
		/bin/bash -c 'make setup; make install; bash'

.PHONY: is-git-dirty
is-git-dirty: ## Tests if git is in a dirty state
	@git status --porcelain
	@test $(shell git status --porcelain | grep -c .) -eq 0

.PHONY: all
all: lint test build coverage ## Test, builds and ship package for all supported platforms

.PHONY: help
help: ## Displays this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
