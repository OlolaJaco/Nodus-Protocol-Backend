.PHONY: help up down run build keys migrate migrate-down lint test clean

APP_NAME := nodus-api
BINARY   := ./bin/$(APP_NAME)
MAIN     := ./cmd/api/main.go
DB_URL   := postgres://postgres:postgres@localhost:5432/nodus_protocol?sslmode=disable

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

## ── Development ──────────────────────────────────────────────────────────────

up: ## Start Postgres and Redis via Docker Compose
	docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 2

down: ## Stop Docker Compose services
	docker compose down

run: ## Run the API server locally
	go run $(MAIN)

build: ## Build the production binary
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY) $(MAIN)
	@echo "Binary: $(BINARY)"

## ── JWT Keys ─────────────────────────────────────────────────────────────────

keys: ## Generate RSA-2048 JWT key pair in ./certs/
	@bash scripts/generate_keys.sh

## ── Database Migrations ──────────────────────────────────────────────────────

migrate: ## Run all pending database migrations
	@which migrate > /dev/null 2>&1 || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down: ## Roll back the last database migration
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-create: ## Create a new migration: make migrate-create name=add_something
	migrate create -ext sql -dir ./migrations -seq $(name)

## ── Code Quality ─────────────────────────────────────────────────────────────

lint: ## Run golangci-lint
	@which golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

test: ## Run all tests
	go test -v -race -count=1 ./...

vet: ## Run go vet
	go vet ./...

## ── Docs ─────────────────────────────────────────────────────────────────────

docs: ## Generate Swagger docs
	@which swag > /dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/api/main.go -o docs

## ── Docker Build ─────────────────────────────────────────────────────────────

docker-build: ## Build the production Docker image
	docker build -t $(APP_NAME):latest .

docker-run: ## Run the production Docker image
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):latest

## ── Cleanup ──────────────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	rm -rf bin/ docs/
