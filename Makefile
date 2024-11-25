NAME           := gitlab-ci-pipelines-exporter
FILES          := $(shell git ls-files */*.go)
COVERAGE_FILE  := coverage.out
REPOSITORY     := mvisonneau/$(NAME)
.DEFAULT_GOAL  := help
GOLANG_VERSION := 1.23

.PHONY: fmt
fmt: ## Format source code
	go run mvdan.cc/gofumpt@v0.7.0 -w $(shell git ls-files **/*.go)
	go run github.com/daixiang0/gci@v0.13.5 write -s standard -s default -s "prefix(github.com/mvisonneau)" .

.PHONY: lint
lint: ## Run all lint related tests upon the codebase
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2 run -v --fast

.PHONY: test
test: ## Run the tests against the codebase
	@rm -rf $(COVERAGE_FILE)
	go test -v -count=1 -race ./... -coverprofile=$(COVERAGE_FILE)
	@go tool cover -func $(COVERAGE_FILE) | awk '/^total/ {print "coverage: " $$3}'

.PHONY: coverage
coverage: ## Prints coverage report
	go tool cover -func $(COVERAGE_FILE)

.PHONY: install
install: ## Build and install locally the binary (dev purpose)
	go install ./cmd/$(NAME)

.PHONY: build
build: ## Build the binaries using local GOOS
	go build ./cmd/$(NAME)

.PHONY: release
release: ## Build & release the binaries (stable)
	mkdir -p ${HOME}/.cache/snapcraft/download
	mkdir -p ${HOME}/.cache/snapcraft/stage-packages
	git tag -d edge
	goreleaser release --clean
	find dist -type f -name "*.snap" -exec snapcraft upload --release stable,edge '{}' \;

.PHONY: protoc
protoc: ## Generate golang from .proto files
	@command -v protoc 2>&1 >/dev/null        || (echo "protoc needs to be available in PATH: https://github.com/protocolbuffers/protobuf/releases"; false)
	@command -v protoc-gen-go 2>&1 >/dev/null || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		pkg/monitor/protobuf/monitor.proto

.PHONY: prerelease
prerelease: ## Build & prerelease the binaries (edge)
	@\
		REPOSITORY=$(REPOSITORY) \
		NAME=$(NAME) \
		GITHUB_TOKEN=$(GITHUB_TOKEN) \
		.github/prerelease.sh

.PHONY: clean
clean: ## Remove binary if it exists
	rm -f $(NAME)

.PHONY: coverage-html
coverage-html: ## Generates coverage report and displays it in the browser
	go tool cover -html=coverage.out

.PHONY: dev-env
dev-env: ## Build a local development environment using Docker
	@docker run -it --rm \
		-v $(shell pwd):/go/src/github.com/mvisonneau/$(NAME) \
		-w /go/src/github.com/mvisonneau/$(NAME) \
		-p 8080:8080 \
		golang:$(GOLANG_VERSION) \
		/bin/bash -c '\
		  git config --global --add safe.directory $$(pwd);\
		  make install;\
		  bash\
		'

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
	curl -sL https://raw.githubusercontent.com/urfave/cli/v2.27.1/autocomplete/bash_autocomplete > helpers/autocomplete/bash
	curl -sL https://raw.githubusercontent.com/urfave/cli/v2.27.1/autocomplete/zsh_autocomplete > helpers/autocomplete/zsh
	curl -sL https://raw.githubusercontent.com/urfave/cli/v2.27.1/autocomplete/powershell_autocomplete.ps1 > helpers/autocomplete/ps1
	
.PHONY: all
all: lint test build coverage ## Test, builds and ship package for all supported platforms

.PHONY: help
help: ## Displays this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
