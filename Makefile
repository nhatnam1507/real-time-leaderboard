.PHONY: help init-dev init-ci build start-dev run check lint ut code-gen doc-gen stop clean

# Variables
COMPOSE_DEV_FILE := docker/docker-compose.dev.yml
COMPOSE_SWARM_FILE := docker/docker-compose.swarm.yml
STACK_NAME := leaderboard
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

## init-dev: Install all required tools and libs for lint, test, build, local full run with docker compose
init-dev:
	@./scripts/init.sh dev

## init-ci: Initialize CI environment (check lint and go installation only)
init-ci:
	@./scripts/init.sh ci

## build: Build the docker image for the application
build:
	@echo "Building docker image..."
	@docker build -f docker/Dockerfile -t leaderboard-app:latest .
	@echo "✓ Docker image built"

## start-dev: Start compose deps file and start the application with air in local VM. Uses migrate script with db url pointing to localhost
start-dev: init-dev
	@./scripts/run.sh dev

## run: Run full Docker Swarm stack with app and deps. Uses migrate script with db url pointing to localhost
run: init-dev build
	@./scripts/run.sh prod-like

## lint: Run linter (golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run ./cmd/... ./internal/...
	@echo "Linter completed successfully"

## ut: Run unit tests with coverage report
ut:
	@./scripts/test.sh

## code-gen: Generate mocks and other code
code-gen:
	@echo "Generating mocks..."
	@go generate ./...
	@echo "✓ Code generation completed"

## doc-gen: Generate JSON from YAML and validate OpenAPI 3.0 specification using kin-openapi
doc-gen:
	@echo "Generating JSON from YAML and validating OpenAPI 3.0 specification..."
	@go run tools/generate-openapi-json.go api/v1/openapi.yaml api/v1/openapi.json
	@echo "✓ OpenAPI v1 specification (YAML and JSON) validated"

## check: Run linter, code generation, unit tests, doc generation, and workflow validation locally
check: lint code-gen ut doc-gen
	@echo "Validating workflow syntax..."
	@./scripts/validate-workflows.sh
	@echo "All checks completed successfully"

## stop: Stop Docker Swarm stack from 'run' target. Containers removed, volumes/data preserved
stop:
	@echo "Stopping Docker Swarm stack..."
	@docker stack rm $(STACK_NAME) 2>/dev/null || true
	@echo "Waiting for stack removal to complete..."
	@timeout=60; interval=2; elapsed=0; \
	while [ $$elapsed -lt $$timeout ]; do \
		if ! docker stack ls | grep -q "^$(STACK_NAME) "; then \
			echo "✓ Stack removed"; \
			break; \
		fi; \
		sleep $$interval; \
		elapsed=$$((elapsed + interval)); \
	done; \
	if [ $$elapsed -ge $$timeout ]; then \
		echo "⚠ Warning: Stack removal timeout"; \
	fi
	@echo "✓ Docker Swarm stack stopped (data preserved)"

## clean: Remove all docker compose stacks, swarm stacks, volumes, and related files
clean:
	@echo "Cleaning up all stacks, volumes, and related files..."
	@docker stack rm $(STACK_NAME) 2>/dev/null || true
	@echo "Waiting for stack removal to complete..."
	@timeout=60; interval=2; elapsed=0; \
	while [ $$elapsed -lt $$timeout ]; do \
		if ! docker stack ls | grep -q "^$(STACK_NAME) "; then \
			echo "✓ Stack removed"; \
			break; \
		fi; \
		sleep $$interval; \
		elapsed=$$((elapsed + interval)); \
	done; \
	if [ $$elapsed -ge $$timeout ]; then \
		echo "⚠ Warning: Stack removal timeout"; \
	fi
	@docker compose -f $(COMPOSE_DEV_FILE) down -v 2>/dev/null || true
	@echo "Removing any remaining containers..."
	@docker ps -a --filter "name=leaderboard" --format "{{.ID}}" | xargs -r docker rm -f 2>/dev/null || true
	@echo "Removing dev and prod volumes..."
	@docker volume rm leaderboard_dev_postgres_data leaderboard_dev_redis_data \
		leaderboard_prod_postgres_data leaderboard_prod_redis_data 2>/dev/null || true
	@echo "Cleaning up build artifacts..."
	@rm -rf tmp/ 2>/dev/null || true
	@echo "✓ Cleanup complete (all data removed)"
