# ─── Development ───────────────────────────────────
.PHONY: dev-be
dev-be:            ## Start the Go backend
	cd cmd/server && go run .

.PHONY: dev-fe
dev-fe:            ## Start the Vite frontend dev server
	cd web && npm run dev

.PHONY: dev
dev:               ## Start both backend and frontend concurrently
	make -j2 dev-be dev-fe

# ─── Code Generation ──────────────────────────────
.PHONY: swagger
swagger:           ## Generate OpenAPI spec from Go annotations
	swag init -g cmd/server/main.go -o docs

.PHONY: codegen
codegen: swagger   ## Generate TypeScript API client from OpenAPI spec
	cd web && npx orval

# ─── Testing ──────────────────────────────────────
.PHONY: test
test:              ## Run all Go tests (uses snapshots if available)
	go test ./... -v

.PHONY: test-refresh
test-refresh:      ## Delete all snapshots and re-record from live API
	rm -rf internal/testutil/snapshots/*.json
	OPEN_ROUTER_API_KEY=$(OPEN_ROUTER_API_KEY) go test ./... -v

.PHONY: test-live
test-live:         ## Run live API tests against OpenRouter (requires OPEN_ROUTER_API_KEY)
	go test ./internal/agent/ -run TestLiveAPI -v -timeout 120s

.PHONY: test-e2e
test-e2e:          ## Run Playwright E2E tests (starts backend + frontend)
	cd web && npx playwright test

# ─── Setup ─────────────────────────────────────────
.PHONY: setup
setup:             ## Install all dependencies
	go mod tidy
	cd web && npm install

.PHONY: install-tools
install-tools:     ## Install required Go tools
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: help
help:              ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
