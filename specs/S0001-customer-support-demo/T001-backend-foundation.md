---
id: T001
name: Backend foundation — data model, seed data, server bootstrap, health endpoint
status: DONE
deps: []
---

# T001 — Backend Foundation

## Description

Stand up the Go backend skeleton with a working HTTP server, SQLite database (GORM), domain models, seed data, and a health endpoint. After this task, running `go run ./cmd/server` starts a server that responds to `GET /api/health` and has a fully seeded database.

This task establishes the base layer that all subsequent tasks build upon: the Chi router, structured logging (`slog` JSON handler), graceful shutdown, and the `Makefile`.

## Scope

- `cmd/server/main.go` — entrypoint, server bootstrap, graceful shutdown with `SIGINT`/`SIGTERM`
- `internal/api/router.go` — Chi router setup with JSON content-type middleware
- `internal/api/health_handler.go` — `GET /api/health` returning `{"status":"ok","model":"gpt-5.4-mini"}`
- `internal/domain/` — all domain models: `Customer`, `Product`, `Order`, `OrderItem`, `RefundPolicy`, `Refund`
- `internal/db/database.go` — GORM setup with SQLite, auto-migrate all models
- `internal/db/seed.go` — seed data population (all customers, products, orders, order items, refund policies from spec §1)
- `Makefile` — `dev-be`, `test`, `setup`, `help` targets
- `go.mod` — module initialization with required dependencies

## Acceptance Criteria

- [ ] `go run ./cmd/server` starts without error and listens on `:8080`
- [ ] `GET /api/health` returns 200 with expected JSON
- [ ] Database is auto-migrated and seeded on startup
- [ ] All 7 products, 5 orders with correct items, 12 refund policy rows, and 1 customer are seeded
- [ ] Server shuts down gracefully on SIGINT
- [ ] Structured JSON logs emitted to stdout on startup and shutdown
- [ ] `make dev-be` runs the server

## Test Cases

### Unit Tests

- `TestSeedData_CustomerCount` — verify exactly 1 customer is seeded
- `TestSeedData_ProductCount` — verify exactly 7 products with correct types
- `TestSeedData_OrderCount` — verify 5 orders with correct item associations
- `TestSeedData_RefundPolicyCount` — verify 12 refund policy rows covering all (type, condition) pairs from spec
- `TestSeedData_OrderItemPrices` — verify order item prices match product prices

### Integration Tests

- `TestDatabase_AutoMigrate` — connect to SQLite, run migrations, verify all tables exist
- `TestDatabase_SeedIdempotent` — run seed twice, verify no duplicate records
- `TestHealthEndpoint` — start server, hit `/api/health`, assert 200 + correct body
