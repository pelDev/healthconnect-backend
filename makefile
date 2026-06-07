.PHONY: migrate-create migrate-up migrate-down migrate-status sqlc-generate

# Database
DB_HOST?=localhost
DB_PORT?=5432
DB_USER?=macbookpro
DB_PASSWORD?=
DB_NAME?=health_connect
DB_SSLMODE?=disable

# Migration
migrate-create:
	@read -p "Enter migration name: " name; \
	go run cmd/migrate/main.go -dir internal/infrastructure/postgres/migrations create $${name} sql

migrate-up:
	DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) DB_SSLMODE=$(DB_SSLMODE) \
	go run cmd/migrate/main.go -dir internal/infrastructure/postgres/migrations up

migrate-down:
	DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) DB_SSLMODE=$(DB_SSLMODE) \
	go run cmd/migrate/main.go -dir internal/infrastructure/postgres/migrations down

migrate-status:
	DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_PASSWORD=$(DB_PASSWORD) DB_NAME=$(DB_NAME) DB_SSLMODE=$(DB_SSLMODE) \
	go run cmd/migrate/main.go -dir internal/infrastructure/postgres/migrations status

# sqlc
sqlc-generate:
	sqlc generate

dev-api:
	go run ./cmd/api/main.go

unit-test-cov:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./internal/domain/... 2>&1 | grep -v "covdata" || true
	@if [ -f coverage.out ]; then \
		echo "\n📊 Coverage Summary:"; \
		go tool cover -func=coverage.out | tail -1; \
		echo "\n📁 Detailed coverage report saved to coverage.out"; \
		echo "Run 'go tool cover -html=coverage.out' to view in browser"; \
	else \
		echo "⚠️  No coverage file generated"; \
	fi