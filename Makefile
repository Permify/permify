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
	docker-compose up --build

.PHONY: compose-up-integration-test
compose-up-integration-test: ### Run docker-compose with integration test
	docker-compose up --build --abort-on-container-exit --exit-code-from integration

.PHONY: compose-down
compose-down: ### Down docker-compose
	docker-compose down --remove-orphans

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
test: ### run tests and gather coverage
	@rm -f covprofile
	@echo "mode: atomic" > covprofile
	@for pkg in $(GO_PACKAGES); do \
		echo "Running tests in $$pkg"; \
		go test -race -coverprofile=covprofile.tmp -covermode=atomic -timeout=10m $$pkg; \
		if [ -f covprofile.tmp ]; then \
			tail -n +2 covprofile.tmp >> covprofile; \
			rm covprofile.tmp; \
		fi; \
	done
	@echo "Coverage profile merged into covprofile"

.PHONY: integration-test
integration-test: ### run integration-test
	go clean -testcache && go test -v ./integration-test/...

.PHONY: build
build: ## Build/compile the Permify service
	go build -o ./permify ./cmd/permify

.PHONY: format
format: ## Auto-format the code
	gofumpt -l -w -extra .

.PHONY: lint-all
lint-all: linter-golangci linter-hadolint linter-dotenv ## Run all linters

.PHONY: security-scan
security-scan: ## Scan code for security vulnerabilities using Gosec
	gosec -exclude=G115 -exclude-dir=sdk -exclude-dir=playground -exclude-dir=docs -exclude-dir=assets ./...

.PHONY: trivy-scan
trivy-scan: ## Scan Docker image for vulnerabilities using Trivy
	docker build -t permify-image .
	trivy image --format json --output trivy-report.json --scanners vuln permify-image

.PHONY: coverage
coverage: ## Generate global code coverage report
	go test -coverprofile=covprofile ./cmd/... ./internal/... ./pkg/...
	go tool cover -html=covprofile -o coverage.html

.PHONY: clean
clean: ## Remove temporary and generated files
	rm -f ./permify
	rm -f ./pkg/development/wasm/main.wasm
	rm -f ./pkg/development/wasm/play.wasm
	rm -f covprofile coverage.html trivy-report.json
	docker rmi -f permify-image || true

.PHONY: wasm-build
wasm-build: ## Build wasm & place it in playground
	cd ./pkg/development/wasm && GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o main.wasm && wasm-opt main.wasm --enable-bulk-memory -Oz -o play.wasm
	cp ./pkg/development/wasm/play.wasm ./playground/public/play.wasm

.PHONY: release
release: format test security-scan trivy-scan clean ## Prepare for release

# Serve

.PHONY: serve
serve: build
	./permify serve

.PHONY: serve-playground
serve-playground:
	cd ./playground && yarn start