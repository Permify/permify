export

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

compose-up: ### Run docker-compose
	docker-compose up --build -d postgres && docker-compose logs -f
.PHONY: compose-up

compose-up-integration-test: ### Run docker-compose with integration test
	docker-compose up --build --abort-on-container-exit --exit-code-from integration
.PHONY: compose-up-integration-test

compose-down: ### Down docker-compose
	docker-compose down --remove-orphans
.PHONY: compose-down

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
	go build -o ./permify ./cmd/permify

.PHONY: serve
run: build ## Run the Permify server with memory
	./permify serve