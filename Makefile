.PHONY: help build run test lint fmt vet clean deps migrate-up migrate-down migrate-create install-tools docker-up docker-down docker-logs

# Variables
APP_NAME := leaderboard
BINARY_NAME := $(APP_NAME)
BINARY_PATH := ./bin/$(BINARY_NAME)
MAIN_PATH := ./cmd/server/main.go
MIGRATIONS_DIR := ./internal/shared/database/migrations
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | awk -F: '{printf "  %-20s %s\n", $$1, $$2}'

## deps: Download and install dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed successfully"

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted successfully"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet check completed"

## lint: Run linter (golangci-lint)
lint:
	@echo "Running linter..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint not found. Installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest; \
	fi
	@golangci-lint run ./...
	@echo "Linter check completed"

## test: Run tests with coverage
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Tests completed"
	@echo "Coverage report saved to $(COVERAGE_FILE)"

## test-coverage: Run tests and show coverage report
test-coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@go tool cover -func=$(COVERAGE_FILE) | tail -1

## test-short: Run tests in short mode
test-short:
	@echo "Running tests (short mode)..."
	@go test -short -v ./...
	@echo "Tests completed"

## build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build completed: $(BINARY_PATH)"

## build-linux: Build the application binary for Linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_PATH)-linux $(MAIN_PATH)
	@echo "Build completed: $(BINARY_PATH)-linux"

## run: Run the application
run:
	@echo "Running application..."
	@go run $(MAIN_PATH)

## clean: Clean build artifacts and test files
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@go clean -cache -testcache
	@echo "Clean completed"

## migrate-up: Run database migrations up
migrate-up:
	@echo "Running database migrations (up)..."
	@./scripts/migrate.sh up
	@echo "Migrations completed"

## migrate-down: Rollback database migrations
migrate-down:
	@echo "Rolling back database migrations..."
	@./scripts/migrate.sh down
	@echo "Rollback completed"

## migrate-create: Create a new migration file
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)..."
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "Migration created successfully"

## install-tools: Install required development tools
install-tools:
	@echo "Installing development tools..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed successfully"

## docker-up: Start Docker services
docker-up:
	@echo "Starting Docker services..."
	@docker-compose up -d
	@echo "Docker services started"

## docker-down: Stop Docker services
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "Docker services stopped"

## docker-logs: Show Docker logs
docker-logs:
	@docker-compose logs -f

## docker-restart: Restart Docker services
docker-restart: docker-down docker-up

## dev: Start development environment (Docker + migrate + run)
dev: docker-up
	@echo "Waiting for services to be ready..."
	@sleep 5
	@$(MAKE) migrate-up
	@echo "Starting application..."
	@$(MAKE) run

## ci: Run CI checks (lint, test, build)
ci: lint test build
	@echo "CI checks completed successfully"

## all: Run all checks and build
all: fmt vet lint test build
	@echo "All checks and build completed successfully"

