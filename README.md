# Real-Time Leaderboard System

A high-performance, modular real-time leaderboard system built with Go, following Clean Architecture principles. The system ranks users based on their scores across various games and provides real-time updates via WebSocket.

## Features

- **User Authentication**: JWT-based authentication with access and refresh tokens
- **Score Submission**: Submit scores for different games with metadata support
- **Real-Time Leaderboards**: Global and game-specific leaderboards with WebSocket updates
- **User Rankings**: Query user rankings in any leaderboard
- **Top Players Reports**: Generate reports with optional date range filtering
- **Redis Sorted Sets**: Efficient leaderboard storage and queries using Redis sorted sets
- **Clean Architecture**: Modular, testable, and maintainable code structure
- **Microservice Ready**: Each module is self-contained and can be extracted to a microservice

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
make init
```

This will:
- Install all required tools (golangci-lint, migrate, air, act)
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

The server will start on `http://localhost:8080`

## Project Structure

```
real-time-leaderboard/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/                     # Configuration management
│   │   └── config.go
│   ├── shared/                     # Shared utilities and infrastructure
│   │   ├── response/               # API response helpers and error definitions
│   │   ├── middleware/             # HTTP middleware
│   │   ├── logger/                 # Logger implementation
│   │   ├── validator/              # Request validation
│   │   ├── database/               # Database connections
│   │   │   └── migrations/        # Database migrations
│   │   └── redis/                  # Redis connections
│   └── module/                     # Self-contained modules
│       ├── auth/                   # Auth Module
│       │   ├── domain/            # Domain layer
│       │   ├── application/       # Application layer
│       │   ├── adapters/          # Adapters layer
│       │   └── infrastructure/    # Infrastructure layer
│       ├── score/                  # Score Module
│       ├── leaderboard/            # Leaderboard Module
│       └── report/                 # Report Module
├── docs/                           # Documentation
├── scripts/                        # Utility scripts
│   ├── init.sh                    # Initialize development environment
│   ├── run.sh                     # Application startup script (dev/all modes)
│   ├── migrate.sh                 # Database migration tool
│   └── validate-workflows.sh      # GitHub Actions workflow validation
├── .githooks/                     # Git hooks (version controlled)
│   └── pre-push                   # Pre-push hook for code quality checks
├── .github/
│   ├── workflows/                 # GitHub Actions workflows
│   │   ├── pr.yml                 # PR workflow (lint + unit tests)
│   │   └── ci.yml                 # CI workflow (lint + unit tests + dockerize)
│   └── actions/                   # Reusable GitHub Actions
│       └── init/                  # Init action (Go setup + make init)
├── docker/
│   ├── Dockerfile                 # Production Docker image
│   ├── docker-compose.deps.yml    # Dependency services (postgres, redis)
│   └── docker-compose.yml         # Full compose file (includes deps + app)
├── api/                           # Generated Swagger documentation
│   ├── docs.go                    # Go package with embedded Swagger docs
│   ├── swagger.json               # OpenAPI 2.0 JSON specification
│   └── swagger.yaml               # OpenAPI 2.0 YAML specification
├── .air.toml                      # Air configuration for hot reload
├── .golangci.yml                  # golangci-lint configuration
├── go.mod
├── Makefile                       # Development commands
└── README.md
```

## Documentation

For detailed documentation, see the [docs](./docs/) folder:

- **[Architecture](./docs/architecture.md)** - System architecture, diagrams, and architectural principles
- **[Modules](./docs/modules.md)** - Detailed module documentation
- **[API Documentation](./docs/api.md)** - Complete API reference
- **[Development Guide](./docs/development.md)** - Development setup, testing, and best practices
- **[Microservice Migration](./docs/microservice-migration.md)** - Guide for extracting modules to microservices
- **[Redis Strategy](./docs/redis-strategy.md)** - Redis sorted sets implementation details

## Common Commands

```bash
# Show all available commands
make help

# Initialize development environment (install tools)
make init

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

# Run all checks (lint + unit tests + workflow validation)
make check

# Generate Swagger documentation
make swagger
```

See [Development Guide](./docs/development.md) for complete list of commands.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
