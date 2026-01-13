# Development Guide

## Project Structure Best Practices

1. **Domain Layer**: Contains pure business logic, no external dependencies
2. **Application Layer**: Orchestrates use cases, depends only on domain interfaces
3. **Adapters Layer**: Handles HTTP/WebSocket, works with domain entities directly
4. **Infrastructure Layer**: Implements repository interfaces, handles external services

## Adding a New Module

1. Create module directory: `internal/module/{module_name}/`
2. Create layers: `domain/`, `application/`, `adapters/`, `infrastructure/`
3. Define domain entities (core business objects)
4. Define repository interfaces in application layer
5. Implement use cases in application layer
6. Create HTTP handlers in adapters layer
7. Implement repositories in infrastructure layer
8. Wire everything in `cmd/server/main.go`

## Makefile Commands

The project includes a simplified Makefile with essential commands for development:

```bash
# Show all available commands
make help

# Initialize development environment
# - Installs golangci-lint, migrate tool, air, wait4x, act, and mockgen (if missing)
# - Configures git hooks automatically (uses .githooks directory)
# - Verifies Docker and Docker Compose are available
# - Downloads Go dependencies
# This is automatically run before check, start-dev, and run targets
make init-dev

# Initialize CI environment (for CI/CD pipelines)
# - Checks Go and golangci-lint installation
# - Downloads Go dependencies
make init-ci

# Start development environment
# - Starts dependency services (PostgreSQL, Redis) via Docker Compose (if not already running)
# - Waits for services to be ready using wait4x (checks actual service health)
# - Runs database migrations (idempotent - safe to run multiple times)
# - Starts the application with hot reload using air
# - Press Ctrl+C to stop and automatically cleanup dependency services
make start-dev
# Or directly: ./scripts/run.sh dev

# Run full Docker Compose setup
# - Starts all services (PostgreSQL, Redis, Application) in containers
# - Waits for services to be ready using wait4x (checks actual service health)
# - Runs database migrations (idempotent - safe to run multiple times)
# - Application runs in Docker container
make run
# Or directly: ./scripts/run.sh all

# Run linter
# - Runs golangci-lint on all code
make lint

# Run unit tests
# - Runs all Go unit tests
make ut

# Generate mocks and other code
# - Generates all mocks from interface definitions using go generate
# - Mocks are generated in internal/module/*/mocks/ directories
make code-gen

# Generate and validate OpenAPI specification
# - Generates api/v1/openapi.json from api/v1/openapi.yaml
# - Validates both YAML and JSON versions
make doc-gen

# Run all checks
# - Runs linter, unit tests, code generation, doc generation, and workflow validation
make check

# Stop the full Docker Compose stack from 'run' target
# - Stops and removes all containers from the full compose file
# - Preserves volumes (data is kept)
# - Use this after running 'make run'
make stop

# Remove all Docker Compose stacks, volumes, and related files
# - Stops and removes all containers from both compose files
# - Removes all volumes (data will be lost)
# - Cleans up build artifacts (tmp/ directory)
make clean
```

### Docker Compose Files

The Docker setup is modularized for better organization:

- **`docker/docker-compose.deps.yml`**: Contains only dependency services (PostgreSQL, Redis)
  - Services are reused between `dev` and `all` modes to avoid conflicts
  - Containers are only started if they don't already exist
  - Stack name: `leaderboard` (defined in compose file)
- **`docker/docker-compose.yml`**: Full compose file that includes deps and adds the application service
  - Uses Docker Compose `include` feature to include the deps file
  - App container uses service names (postgres/redis) from docker network
  - Stack name: `leaderboard` (defined in compose file)
  - Can be used for production deployments

**Stopping Services**:
- In `dev` mode: Press Ctrl+C to stop the application and automatically cleanup dependency services
- In `all` mode (after `make run`): Use `make stop` to stop the full stack (preserves data) or `make clean` to remove everything

### Scripts

The project includes utility scripts in the `scripts/` directory:

- **`run.sh`**: Unified script for starting the application
  - `./scripts/run.sh dev` - Start deps and run app with air (hot reload)
    - Press Ctrl+C to stop and automatically cleanup dependency services
  - `./scripts/run.sh all` - Start full docker compose environment (runs in detached mode)
  - Handles service health checking with wait4x
  - Prevents container conflicts by checking if services are already running
  - **Can be run from any directory** - paths are resolved relative to script location
  
- **`migrate.sh`**: Database migration tool
  - `./scripts/migrate.sh up [migration_dir]` - Apply migrations from specified directory
  - `./scripts/migrate.sh down [migration_dir]` - Rollback one migration from specified directory
  - Idempotent - safe to run multiple times
  - Validates parameter count and provides helpful error messages
  - **Can be run from any directory** - paths are resolved relative to script location

- **`init.sh`**: Development environment initialization
  - Tool-style script with two modes: `dev` and `ci`
  - `dev` mode: Installs required tools (golangci-lint, migrate, air, wait4x, act, mockgen), configures git hooks, verifies Docker/Docker Compose, downloads Go dependencies
  - `ci` mode: Only checks Go and golangci-lint installation, downloads Go dependencies
  - Usage: `./scripts/init.sh [dev|ci]`

- **`validate-workflows.sh`**: GitHub Actions workflow validation
  - Validates workflow YAML syntax using act
  - Used automatically by `make check`

### Development Workflow

1. **First time setup**:
   ```bash
   make init-dev
   ```

2. **Daily development**:
   ```bash
   make start-dev
   ```
   This starts dependencies in Docker and runs the app locally with hot reload.
   - Press Ctrl+C to stop and automatically cleanup dependency services
   - Air hot reload watches Go files (excludes docker/, docs/, scripts/ directories)

3. **Testing full Docker setup**:
   ```bash
   make run
   ```
   This runs everything in Docker containers for production-like testing.
   - Services run in detached mode
   - Use `make stop` to stop the full stack (preserves data) or `make clean` to remove everything

4. **Before committing**:
   ```bash
   make check
   ```
   This runs linter, unit tests, code generation, doc generation, and workflow validation. A pre-push git hook automatically runs `make check` before pushing.
   
   **Note**: If you've changed interfaces, make sure to run `make code-gen` first to regenerate mocks, then commit the updated mocks along with your changes.

5. **Stopping services**:
   ```bash
   # Stop the full stack from 'run' target but preserve data
   make stop
   
   # Remove everything including data and build artifacts
   make clean
   ```
   - `make stop`: Stops the full compose stack (from `make run`) but preserves volumes (data is kept)
   - `make clean`: Removes all containers, volumes, and build artifacts (complete cleanup)

## Testing

Run unit tests:
```bash
make ut
```

The `make check` command runs linter, unit tests, code generation, doc generation, and workflow validation together.

### Code Generation

The project uses `go generate` to generate mocks for testing. Mocks are committed to the repository and should be regenerated when interfaces change.

**Generate mocks**:
```bash
make code-gen
```

This runs `go generate ./...` which processes all `//go:generate` directives in the codebase. Currently, mocks are generated for:
- `UserRepository` and `JWTManager` interfaces in the auth module
- `UserRepository`, `LeaderboardBackupRepository`, `LeaderboardRepository`, and `BroadcastService` interfaces in the leaderboard module

**When to regenerate**:
- After adding or modifying interface methods
- After changing interface signatures
- Before committing interface changes (mocks should be committed with code changes)

**Mock location**: Mocks are generated in `internal/module/{module_name}/mocks/` directories at the module level (same level as `application/`, `domain/`, etc.).

## Git Hooks

Git hooks are automatically configured when you run `make init`. The hooks are version controlled in `.githooks/` directory.

- **Pre-push hook**: Automatically runs `make check` before allowing push to remote repository
- To bypass (not recommended): `git push --no-verify`

## CI/CD

The project includes GitHub Actions workflows:

- **PR Workflow** (`.github/workflows/pr.yml`): Runs on pull requests
  - Lint and unit tests in parallel
  
- **CI Workflow** (`.github/workflows/ci.yml`): Runs on pushes to main
  - Lint and unit tests in parallel
  - Docker image build after checks pass

Workflow syntax is validated automatically as part of `make check`.

## Database Migrations

Migrations are automatically run when using `make start-dev` or `make run`. The migration tool is idempotent, meaning it's safe to run multiple times - it only applies new migrations that haven't been run yet.

### Migration Structure

Migrations are organized into two directories:
- **`migrations/schema/`**: Core schema migrations for all environments (tables, indexes, etc.)
- **`migrations/dev/`**: Development-only seed data migrations (test users, sample scores)

When running `make start-dev`, both schema and dev migrations are applied. When running `make run` (production-like mode), only schema migrations are applied.

### Running Migrations Manually

You can run migrations manually using the `migrate.sh` tool:

```bash
# Run all pending schema migrations
./scripts/migrate.sh up migrations/schema

# Run all pending dev migrations
./scripts/migrate.sh up migrations/dev

# Rollback one migration from schema
./scripts/migrate.sh down migrations/schema

# Rollback one migration from dev
./scripts/migrate.sh down migrations/dev

# Custom DB URL (if needed)
DB_URL=postgres://user:pass@host:5432/dbname?sslmode=disable ./scripts/migrate.sh up migrations/schema
```

**Note**: All scripts resolve paths relative to their location, so they can be run from any directory without path issues.

**Note**: The migrate script automatically uses `localhost:5432` by default since migrations run on the host machine. The DB_URL can be overridden via environment variable if needed.

**Create a new migration**:
```bash
# Create a new schema migration
migrate create -ext sql -dir migrations/schema -seq migration_name

# Create a new dev migration
migrate create -ext sql -dir migrations/dev -seq migration_name
```

### Dev Seed Data

The `migrations/dev/` directory contains seed data for development:
- **Test users**: 10 test users (alice, bob, charlie, etc.) with password `password123`
- **Leaderboard scores**: Pre-populated scores for all test users with varying points

All seed data migrations are idempotent - they can be run multiple times without creating duplicates. The seed data uses `ON CONFLICT` clauses to ensure safe re-execution.

The `migrate` and `wait4x` tools are automatically installed by `make init-dev` if not already present.

## Hot Reload Configuration

The project uses [Air](https://github.com/air-verse/air) for hot reload during development. The configuration is in `.air.toml` at the project root.

**Key features**:
- Automatically rebuilds and restarts the application when Go files change
- Excludes `docker/`, `docs/`, and `scripts/` directories from file watching
- Build errors are logged to `tmp/build-errors.log`

The Air configuration is automatically used when running `make start-dev` or `./scripts/run.sh dev`.


## Technology Stack

- **Language**: Go 1.25.5
- **Web Framework**: Gin
- **Database**: PostgreSQL 15+
- **Cache/Ranking**: Redis 7+ (sorted sets)
- **Authentication**: JWT (github.com/golang-jwt/jwt/v5)
- **Database Driver**: pgx/v5
- **Redis Client**: go-redis/v9
- **Validation**: go-playground/validator/v10
- **Logging**: zerolog
- **API Documentation**: OpenAPI 3.0 (spec-first approach)
- **Testing**: 
  - testify (assertions)
  - go.uber.org/mock (mock generation)

