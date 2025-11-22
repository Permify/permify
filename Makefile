export

GO_PACKAGES := $(shell find ./cmd ./pkg ./internal -name '*_test.go' | xargs -n1 dirname | sort -u)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: compose-up
compose-up: ### Run docker-compose
	docker-compose up --build # Build and start containers

.PHONY: compose-up-integration-test
compose-up-integration-test: ### Run docker-compose with integration test
	docker-compose up --build --abort-on-container-exit --exit-code-from integration

.PHONY: compose-down
compose-down: ### Down docker-compose
	docker-compose down --remove-orphans

.PHONY: download
download:
	@go mod download

.PHONY: linter-golangci
linter-golangci: ### check by golangci linter
	go tool golangci-lint run

.PHONY: linter-hadolint
linter-hadolint: ### check by hadolint linter
	@find . -name 'Dockerfile*' -type f | xargs hadolint || true

.PHONY: test
test: ### run tests and gather coverage
	@go clean -testcache
	@go test -race -coverprofile=coverage.txt -covermode=atomic -timeout=10m $(GO_PACKAGES)

.PHONY: integration-test
integration-test: ### run integration-test
	go clean -testcache && go test -v ./integration-test/...

.PHONY: build
build: ## Build/compile the Permify service
	go build -o ./permify ./cmd/permify

.PHONY: format
format: ## Auto-format the code
	go tool gofumpt -l -w -extra .

.PHONY: lint-all
lint-all: linter-golangci linter-hadolint ## Run all linters

.PHONY: security-scan
security-scan: ## Scan code for security vulnerabilities using Gosec
	go tool gosec -exclude=G115,G103 -exclude-dir=sdk -exclude-dir=playground -exclude-dir=docs -exclude-dir=assets -exclude-dir=pkg/pb ./...

.PHONY: trivy-scan
trivy-scan: ## Scan Docker image for vulnerabilities using Trivy
	docker build -t permify-image .
	trivy image --format json --output trivy-report.json --scanners vuln permify-image

.PHONY: coverage
coverage: ## Generate global code coverage report
	go test -coverprofile=coverage.txt ./cmd/... ./internal/... ./pkg/...
	go tool cover -html=coverage.txt -o coverage.html

.PHONY: clean
clean: ## Remove temporary and generated files
	rm -f ./permify
	rm -f ./pkg/development/wasm/main.wasm
	rm -f ./pkg/development/wasm/play.wasm
	rm -f coverage.txt coverage.html trivy-report.json
	docker rmi -f permify-image || true

.PHONY: wasm-build
wasm-build: ## Build wasm & place it in playground # WebAssembly build target
	cd ./pkg/development/wasm && GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o main.wasm && wasm-opt main.wasm --enable-bulk-memory -Oz -o play.wasm
	cp ./pkg/development/wasm/play.wasm ./playground/public/play.wasm # Copy to playground

.PHONY: release
release: format lint-all test security-scan trivy-scan clean ## Prepare for release

# Serve

.PHONY: serve
serve: build
	./permify serve

.PHONY: serve-playground
serve-playground:
	cd ./playground && yarn start