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

The project includes a simplified Makefile with essential commands for development:

```bash
# Show all available commands
make help

# Initialize development environment
# - Installs golangci-lint, migrate tool, and air (if missing)
# - Verifies Docker and Docker Compose are available
# - Downloads Go dependencies
# This is automatically run before other targets
make init

# Start development environment
# - Starts dependency services (PostgreSQL, Redis) via Docker Compose
# - Waits for services to be ready
# - Runs database migrations
# - Starts the application with hot reload using air
make start-dev

# Run full Docker Compose setup
# - Starts all services (PostgreSQL, Redis, Application) in containers
# - Waits for services to be ready
# - Runs database migrations
# - Application runs in Docker container
make run

# Run linter and tests
# - Runs golangci-lint on all code
# - Runs all Go unit tests
make check
```

### Docker Compose Files

The Docker setup is modularized for better organization:

- **`docker/docker-compose.deps.yml`**: Contains only dependency services (PostgreSQL, Redis)
- **`docker/docker-compose.yml`**: Full compose file that includes deps and adds the application service
  - Uses Docker Compose `include` feature to include the deps file
  - Can be used for production deployments

### Development Workflow

1. **First time setup**:
   ```bash
   make init
   ```

2. **Daily development**:
   ```bash
   make start-dev
   ```
   This starts dependencies in Docker and runs the app locally with hot reload.

3. **Testing full Docker setup**:
   ```bash
   make run
   ```
   This runs everything in Docker containers for production-like testing.

4. **Before committing**:
   ```bash
   make check
   ```
   This runs linter and tests to ensure code quality.

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

Migrations are automatically run when using `make start-dev` or `make run`. You can also run them manually:

```bash
# Run migrations (using the migrate script)
# For local development (when using make start-dev)
DB_URL=postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable ./scripts/migrate.sh up

# For Docker Compose (when using make run)
DB_URL=postgres://postgres:postgres@postgres:5432/leaderboard?sslmode=disable ./scripts/migrate.sh up

# Rollback migrations
DB_URL=postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable ./scripts/migrate.sh down

# Create a new migration
migrate create -ext sql -dir internal/shared/database/migrations -seq migration_name
```

The `migrate` tool is automatically installed by `make init` if not already present.

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

