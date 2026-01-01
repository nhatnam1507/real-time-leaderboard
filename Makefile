.PHONY: help init start-dev run check

# Variables
COMPOSE_DEPS_FILE := docker/docker-compose.deps.yml
COMPOSE_FULL_FILE := docker/docker-compose.yml
MIGRATE_SCRIPT := ./scripts/migrate.sh
PATH := $(shell go env GOPATH)/bin:$(PATH)
export PATH

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | awk -F: '{printf "  %-20s %s\n", $$1, $$2}'

## init: Install all required tools and libs for lint, test, build, local full run with docker compose
init:
	@./scripts/init.sh

## start-dev: Start compose deps file and start the application with air in local VM. Uses migrate script with db url pointing to localhost
start-dev:
	@SILENT=1 ./scripts/init.sh
	@echo "Starting development environment..."
	@echo "Starting dependency services..."
	@docker compose -f $(COMPOSE_DEPS_FILE) up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	@DB_URL=postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable $(MIGRATE_SCRIPT) up
	@echo "Starting application with air..."
	@air

## run: Run full run with app and deps via docker compose in local VM. Uses migrate script with db url pointing to service name in compose file
run:
	@SILENT=1 ./scripts/init.sh
	@echo "Starting full docker compose environment..."
	@docker compose -f $(COMPOSE_FULL_FILE) up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Running database migrations..."
	@DB_URL=postgres://postgres:postgres@postgres:5432/leaderboard?sslmode=disable $(MIGRATE_SCRIPT) up
	@echo "Services are running. Use 'docker compose -f $(COMPOSE_FULL_FILE) logs -f' to view logs."

## check: Run linter and test (golangci-lint and go test for all unit test)
check:
	@SILENT=1 ./scripts/init.sh
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Running tests..."
	@go test ./...
	@echo "All checks completed successfully"
