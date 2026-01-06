.PHONY: help init start-dev run check lint ut stop clean

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
	@./scripts/run.sh dev

## run: Run full run with app and deps via docker compose in local VM. Uses migrate script with db url pointing to localhost
run:
	@SILENT=1 ./scripts/init.sh
	@./scripts/run.sh all

## lint: Run linter (golangci-lint)
lint:
	@SILENT=1 ./scripts/init.sh
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Linter completed successfully"

## ut: Run unit tests
ut:
	@SILENT=1 ./scripts/init.sh
	@echo "Running unit tests..."
	@go test ./...
	@echo "Unit tests completed successfully"

## check: Run both linter and unit tests
check: lint ut
	@echo "All checks completed successfully"

## stop: Stop full compose stack from 'run' target. Containers removed, volumes/data preserved
stop:
	@echo "Stopping full compose stack..."
	@docker compose -f $(COMPOSE_FULL_FILE) down 2>/dev/null || true
	@echo "✓ Full compose stack stopped (data preserved)"

## clean: Remove all docker compose stacks, volumes, and related files
clean:
	@echo "Cleaning up all compose stacks, volumes, and related files..."
	@docker compose -f $(COMPOSE_FULL_FILE) down -v 2>/dev/null || true
	@docker compose -f $(COMPOSE_DEPS_FILE) down -v 2>/dev/null || true
	@echo "Cleaning up build artifacts..."
	@rm -rf tmp/ 2>/dev/null || true
	@echo "✓ Cleanup complete (all data removed)"
