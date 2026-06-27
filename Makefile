.PHONY: setup dev test test-unit test-integration lint db-up db-down db-migrate-up db-migrate-down api mcp ingest admin web

# Environment file check
ENV_FILE := .env

setup:
	@echo "=== Embalses MVP Setup ==="
	@echo "1. Copy .env.example to .env and fill values"
	@test -f $(ENV_FILE) || cp .env.example .env
	@echo "2. Install Go dependencies"
	go mod download
	@echo "3. Install frontend dependencies"
	cd web && npm install
	@echo "4. Setup complete. Run 'make dev' to start the stack."

dev:
	@echo "Starting development stack..."
	docker compose up --build -d db api mcp web

# Database targets
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
	go test -short ./...

test-integration:
	go test -run Integration ./...

# Linting
lint:
	@echo "Linting Go..."
	go vet ./...
	gofmt -l .
	@echo "Linting frontend..."
	cd web && npm run lint
