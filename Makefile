.PHONY: help build run test clean docker-build docker-run migrate

# Variables
BINARY_NAME=admin-deletion-dashboard
DOCKER_IMAGE=appointy/admin-deletion-dashboard
VERSION?=latest

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/main

run: ## Run the application locally
	@echo "Running $(BINARY_NAME)..."
	@go run ./cmd/main/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@go clean

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(VERSION)..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(VERSION)

migrate: ## Run database migrations
	@echo "Running database migrations..."
	@psql "$(DATABASE_URL)" -f migrations/001_create_audit_table.sql

dev: ## Run in development mode with hot reload (requires air: go install github.com/cosmtrek/air@latest)
	@air

lint: ## Run linter (requires golangci-lint)
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

.DEFAULT_GOAL := help
