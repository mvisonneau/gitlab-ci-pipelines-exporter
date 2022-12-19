NAME          := gitlab-ci-pipelines-exporter
FILES         := $(shell git ls-files */*.go)
REPOSITORY    := mvisonneau/$(NAME)
.DEFAULT_GOAL := help

.PHONY: setup
setup: ## Install required libraries/tools for build tasks
	@command -v gci 2>&1 >/dev/null           || go install github.com/daixiang0/gci@v0.9.0
	@command -v gofumpt 2>&1 >/dev/null       || go install mvdan.cc/gofumpt@v0.4.0
	@command -v golangci-lint 2>&1 >/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1

.PHONY: fmt
fmt: setup ## Format source code
	gofumpt -w $(FILES)
	gci write -s standard -s default -s "prefix(github.com/mvisonneau)" .

.PHONY: lint
lint: setup ## Run all lint related tests upon the codebase
	golangci-lint run -v --fast

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

.PHONY: protoc
protoc: setup ## Generate golang from .proto files
	@command -v protoc 2>&1 >/dev/null        || (echo "protoc needs to be available in PATH: https://github.com/protocolbuffers/protobuf/releases"; false)
	@command -v protoc-gen-go 2>&1 >/dev/null || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/monitor/protobuf/monitor.proto

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
		golang:1.19 \
		/bin/bash -c 'make setup; make install; bash'

.PHONY: is-git-dirty
is-git-dirty: ## Tests if git is in a dirty state
	@git status --porcelain
	@test $(shell git status --porcelain | grep -c .) -eq 0

.PHONY: man-pages
man-pages: ## Generates man pages
	rm -rf helpers/manpages
	mkdir -p helpers/manpages
	go run ./cmd/tools/man | gzip -c -9 >helpers/manpages/$(NAME).1.gz

.PHONY: autocomplete-scripts
autocomplete-scripts: ## Download CLI autocompletion scripts
	rm -rf helpers/autocomplete
	mkdir -p helpers/autocomplete
	curl -sL https://raw.githubusercontent.com/urfave/cli/v2.5.0/autocomplete/bash_autocomplete > helpers/autocomplete/bash
	curl -sL https://raw.githubusercontent.com/urfave/cli/v2.5.0/autocomplete/zsh_autocomplete > helpers/autocomplete/zsh
	
.PHONY: all
all: lint test build coverage ## Test, builds and ship package for all supported platforms

.PHONY: help
help: ## Displays this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
