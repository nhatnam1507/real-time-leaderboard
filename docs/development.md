# Development Guide

## Project Structure Best Practices

1. **Domain Layer**: Contains pure business logic, no external dependencies
2. **Application Layer**: Orchestrates use cases, depends only on domain interfaces
3. **Adapters Layer**: Handles HTTP/WebSocket, works with domain entities directly
4. **Infrastructure Layer**: Implements repository interfaces, handles external services

## Adding a New Module

1. Create module directory: `internal/module/{module_name}/`
2. Create layers: `domain/`, `application/`, `adapters/`, `infrastructure/`
3. Define domain entities and repository interfaces
4. Implement use cases in application layer
5. Create HTTP handlers in adapters layer
6. Implement repositories in infrastructure layer
7. Wire everything in `cmd/server/main.go`

## Makefile Commands

The project includes a Makefile with convenient commands for development and CI/CD:

```bash
# Show all available commands
make help

# Install dependencies
make deps

# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Run tests with coverage
make test

# Run tests and generate HTML coverage report
make test-coverage

# Build the application
make build

# Run the application
make run

# Run database migrations
make migrate-up

# Rollback database migrations
make migrate-down

# Create a new migration (requires NAME parameter)
make migrate-create NAME=migration_name

# Install development tools (migrate, etc.)
make install-tools

# Start Docker services
make docker-up

# Stop Docker services
make docker-down

# View Docker logs
make docker-logs

# Start full development environment (Docker + migrate + run)
make dev

# Run CI checks (lint + test + build)
make ci

# Run all checks and build
make all

# Clean build artifacts
make clean
```

## Testing

```bash
# Run all tests (using make)
make test

# Or using go directly
go test ./...

# Run tests with coverage
make test-coverage

# Run tests for a specific module
go test ./internal/module/auth/...
```

## Database Migrations

```bash
# Run migrations (using make)
make migrate-up

# Rollback migrations
make migrate-down

# Create a new migration
make migrate-create NAME=migration_name

# Or using migrate directly
migrate create -ext sql -dir internal/shared/database/migrations -seq migration_name
./scripts/migrate.sh up
./scripts/migrate.sh down
```

## Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: PostgreSQL 15+
- **Cache/Ranking**: Redis 7+ (sorted sets)
- **Authentication**: JWT (github.com/golang-jwt/jwt/v5)
- **Database Driver**: pgx/v5
- **Redis Client**: go-redis/v9
- **WebSocket**: Gorilla WebSocket
- **Validation**: go-playground/validator/v10
- **Logging**: zerolog

