# ──────────────────────────────────────────────────────────────────────────────
# Makefile — developer workflow shortcuts
# ──────────────────────────────────────────────────────────────────────────────
.PHONY: help build test test-race test-cover test-docker \
        services-up services-down lint vuln tidy docker-build clean

GO      := go
DC      := docker compose
DC_TEST := docker compose -f docker-compose.test.yml

# ── Help ──────────────────────────────────────────────────────────────────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
	  awk 'BEGIN{FS=":.*##"} {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Build ─────────────────────────────────────────────────────────────────────
build: ## Build all packages
	$(GO) build ./...

# ── Tests (local — services must be running) ──────────────────────────────────
test: ## Run tests (no race detector; requires local services or services-up)
	$(GO) test -v -count=1 ./...

test-race: ## Run tests with race detector (requires CGO_ENABLED=1)
	CGO_ENABLED=1 $(GO) test -v -race -count=1 ./...

test-cover: ## Run tests and show coverage summary
	$(GO) test -v -count=1 -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

# ── Tests in Docker (all services included, no local setup needed) ─────────────
test-docker: ## Run full test suite inside Docker (builds runner + starts services)
	$(DC_TEST) up --build --abort-on-container-exit --exit-code-from test
	$(DC_TEST) down -v

test-docker-clean: ## Like test-docker but removes the module cache volume first
	$(DC_TEST) down -v
	$(DC_TEST) up --build --abort-on-container-exit --exit-code-from test
	$(DC_TEST) down -v

# ── Local services only ───────────────────────────────────────────────────────
services-up: ## Start Redis/MySQL/Postgres in the background (for local test runs)
	$(DC) up -d

services-down: ## Stop and remove local service containers
	$(DC) down -v

# ── Code quality ─────────────────────────────────────────────────────────────
lint: ## Run golangci-lint
	golangci-lint run --timeout=5m

vuln: ## Run govulncheck (dependency vulnerability scan)
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...

tidy: ## Tidy and verify go.mod / go.sum
	$(GO) mod tidy
	$(GO) mod verify

# ── Docker image ──────────────────────────────────────────────────────────────
docker-build: ## Build the production Docker image for cmd/example
	docker build -t turahe/pkg-example:latest .

# ── Clean ─────────────────────────────────────────────────────────────────────
clean: ## Remove coverage output and test logs
	rm -f coverage.out test.log
