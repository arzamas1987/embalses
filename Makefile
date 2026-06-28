.PHONY: setup dev test test-unit test-integration lint db-up db-down db-migrate-up db-migrate-down api api-sqlite mcp ingest admin updater web

# Environment file check
ENV_FILE := .env

# Detect backend: if DATABASE_URL starts with sqlite, use SQLite; otherwise Postgres
ifeq ($(findstring sqlite,$(DATABASE_URL)),sqlite)
  BACKEND := sqlite
else
  BACKEND := postgres
endif

setup:
	@echo "=== Embalses MVP Setup ==="
	@echo "1. Copy .env.example to .env and fill values"
	@test -f $(ENV_FILE) || cp .env.example .env
	@echo "2. Install Go dependencies"
	go mod download
	@echo "3. Install frontend dependencies"
	cd web && npm install
	@echo "4. Setup complete. Run 'make dev' to start the stack."
	@echo "5. For SQLite backend: make updater && make api-sqlite"

dev:
	@echo "Starting development stack (PostgreSQL)..."
	docker compose up --build -d db api mcp web

# ── SQLite targets ──
updater:
	go build -o bin/updater ./cmd/updater
	@echo "Updater built. Run: ./bin/updater -db data/embalses.db -geo-only -seed-readings"

updater-run:
	./bin/updater -db data/embalses.db -geo-only -seed-readings

api-sqlite:
	go build -o bin/api-sqlite ./cmd/api-sqlite

api-sqlite-run:
	DATABASE_URL=data/embalses.db ./bin/api-sqlite

# PostgreSQL targets (legacy)
db-up:
	docker compose up -d db

db-down:
	docker compose down db

db-migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

db-migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

# Service builds
api:
	go build -o bin/api ./cmd/api

mcp:
	go build -o bin/mcp ./cmd/mcp

ingest:
	go build -o bin/ingest ./cmd/ingest

admin:
	go build -o bin/admin ./cmd/admin

web:
	cd web && npm run build

# Testing
test: test-unit test-integration

test-unit:
	go test -short -p 1 ./...

test-integration:
	go test -p 1 ./...

# Linting
lint:
	@echo "Linting Go..."
	go vet ./...
	gofmt -l .
	@echo "Linting frontend..."
	cd web && npm run lint
