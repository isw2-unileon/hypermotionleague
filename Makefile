.PHONY: install run-backend run-frontend build-backend build-frontend test lint e2e migrate seed

## Database URL — override with: make migrate DB_URL=postgres://user:pass@host/db
DB_URL ?= postgresql://postgres:postgres@localhost:5432/hypermotionleague?sslmode=disable

## Install all dependencies
install:
	go install github.com/air-verse/air@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	go mod download
	cd frontend && npm ci
	cd e2e && npm ci

## Run backend with hot reload
run-backend:
	$(shell go env GOPATH)/bin/air -c backend/.air.toml

## Run frontend dev server
run-frontend:
	cd frontend && npm run dev

## Build backend binary
build-backend:
	go build -o backend/bin/server ./backend/cmd/server

## Build frontend for production
build-frontend:
	cd frontend && npm run build

## Run all tests
test:
	go test -v -race ./...
	cd frontend && npm run test

## Run linters
lint:
	$(shell go env GOPATH)/bin/golangci-lint run
	cd frontend && npm run lint

## Run E2E tests (requires backend + frontend running)
e2e:
	cd e2e && npx playwright test

## Apply all pending migrations (idempotent — skips already-applied versions)
migrate:
	@psql "$(DB_URL)" -c "CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ DEFAULT NOW())" 2>/dev/null || true
	@for f in backend/db/migrations/001_initial_schema.up.sql \
	           backend/db/migrations/002_add_auth_provider.up.sql \
	           backend/db/migrations/003_fix_lineup_position_constraint.up.sql \
	           backend/db/migrations/004_extend_player_points_stats.up.sql \
	           backend/db/migrations/005_unique_bid_per_listing_user.up.sql; do \

	    v=$$(basename $$f .up.sql); \
	    if ! psql "$(DB_URL)" -tAc "SELECT 1 FROM schema_migrations WHERE version='$$v'" 2>/dev/null | grep -q 1; then \
	        echo "Applying $$v ..."; \
	        psql "$(DB_URL)" -f "$$f" && \
	        psql "$(DB_URL)" -c "INSERT INTO schema_migrations(version) VALUES ('$$v')"; \
	    else \
	        echo "Skipping $$v (already applied)"; \
	    fi; \
	done

## Load seed data with live dates (idempotent — safe to re-run)
seed:
	@echo "Loading static seed data..."
	@psql "$(DB_URL)" -f backend/db/seeds/001_sample_data.sql
	@echo "Applying live dates..."
	@psql "$(DB_URL)" -f backend/db/seeds/002_live_dates.sql
	@echo "Seed complete."
