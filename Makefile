export


REPO    := github.com/Permify/permify
HASH    := $(shell git rev-parse --short HEAD)
DATE    := $(shell date)
TAG     := $(shell git describe --tags --always --abbrev=0 --match="v[0-9]*.[0-9]*.[0-9]*" 2> /dev/null)
VERSION := $(shell echo "${TAG}" | sed 's/^.//')

LDFLAGS_RELEASE := -ldflags "-X '${REPO}/pkg/cmd.Version=${VERSION}' -X '${REPO}/pkg/cmd.BuildDate=${DATE}' -X '${REPO}/pkg/cmd.GitCommit=${HASH}'"


# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: download
download:
	@cd tools/ && go mod download

.PHONY: linter-golangci
linter-golangci: ### check by golangci linter
	golangci-lint run

.PHONY: linter-hadolint
linter-hadolint: ### check by hadolint linter
	git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint

.PHONY: linter-dotenv
linter-dotenv: ### check by dotenv linter
	dotenv-linter

.PHONY: test
test: ### run test
	go test -v -cover -race ./internal/...

.PHONY: integration-test
integration-test: ### run integration-test
	go clean -testcache && go test -v ./integration-test/...

.PHONY: build
build: ## Build/compile the Permify service
	go build ${LDFLAGS_RELEASE} -o ./permify ./cmd/permify

.PHONY: serve
run: build ## Run the Permify server with memory
	./permify serve