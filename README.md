# Real-Time Leaderboard System

A high-performance, modular real-time leaderboard system built with Go, following Clean Architecture principles. The system ranks users based on their scores.

## Prerequisites

- Go 1.25.5
- Docker and Docker Compose
- PostgreSQL 15+ (if running locally)
- Redis 7+ (if running locally)

## Installation

1. **Clone the repository**:
```bash
git clone <repository-url>
cd real-time-leaderboard
```

2. **Initialize development environment**:
```bash
make init-dev
```

This will:
- Install all required tools (golangci-lint, migrate, air, wait4x, act, mockgen)
- Configure git hooks automatically
- Verify Docker/Docker Compose are available

## Quick Start

### Development Mode (Recommended)

Start the development environment with hot reload:

```bash
make start-dev
```

Or directly (can be run from any directory):
```bash
./scripts/run.sh dev
```

This will:
- Start dependency services (PostgreSQL, Redis) via Docker Compose (if not already running)
- Wait for services to be ready using wait4x
- Run database migrations (idempotent)
- Start the application with hot reload using `air`
- **Press Ctrl+C to stop the application and automatically cleanup dependency services**

The server will start on `http://localhost:8080`

### Production-like Mode

Run full Docker Compose setup (all services in containers):

```bash
make run
```

Or directly (can be run from any directory):
```bash
./scripts/run.sh all
```

This will:
- Start all services (PostgreSQL, Redis, Application) via Docker Compose
- Wait for services to be ready using wait4x
- Run database migrations (idempotent)
- Start the application in a container
- Rebuild the application image (via `make build`) before starting

The server will start on `http://localhost:8080`

## Documentation

For detailed documentation, see the [docs](./docs/) folder:

- **[Application Features & Flows](./docs/application.md)** - Application features, user flows, and API usage
- **[Architecture](./docs/architecture.md)** - Clean architecture principles, project structure, coding conventions, and infrastructure details
- **[Modules](./docs/modules.md)** - Detailed module documentation
- **[Development Guide](./docs/development.md)** - Development setup, testing, and best practices
- **[Microservice Migration](./docs/microservice-migration.md)** - Guide for extracting modules to microservices

### API Documentation

The API is documented using OpenAPI 3.0 specification (spec-first approach):

- **OpenAPI Spec**: `api/v1/openapi.yaml` - The source of truth for API documentation
- **Swagger UI**: http://localhost:8080/docs/index.html - Interactive API documentation
- **OpenAPI YAML**: http://localhost:8080/api/v1/openapi.yaml
- **OpenAPI JSON**: http://localhost:8080/api/v1/openapi.json

To generate JSON from YAML and validate the OpenAPI specification:
```bash
make doc-gen
```

This will:
- Generate `api/v1/openapi.json` from `api/v1/openapi.yaml` (YAML is the source of truth)
- Validate both YAML and JSON versions

To generate mocks and other code:
```bash
make code-gen
```

This will:
- Generate all mocks from interface definitions using `go generate ./...`
- Mocks are generated in `internal/module/*/mocks/` directories

**Note**: The OpenAPI spec is the single source of truth. All API documentation is maintained in the spec file, not in code annotations.

## Common Commands

```bash
# Show all available commands
make help

# Initialize development environment (install tools)
make init-dev

# Rebuild the application Docker image
make build

# Start development environment (deps + app with hot reload)
make start-dev
# Press Ctrl+C to stop and cleanup dependency services

# Run full Docker Compose setup (all services in containers)
make run

# Stop the full Docker Compose stack from 'run' target (preserves data/volumes)
make stop

# Remove all Docker Compose stacks, volumes, and related files
make clean

# Run linter
make lint

# Run unit tests
make ut

# Run all checks (lint + unit tests + code generation + doc generation + workflow validation)
make check

# Generate mocks and other code
make code-gen

# Generate and validate OpenAPI specification
make doc-gen
```

See [Development Guide](./docs/development.md) for complete list of commands.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
