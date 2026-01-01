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

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15+ (if running locally)
- Redis 7+ (if running locally)

### Installation

1. **Clone the repository**:
```bash
git clone <repository-url>
cd real-time-leaderboard
```

2. **Initialize development environment**:
```bash
make init
```

This will install all required tools (golangci-lint, migrate, air) and verify Docker/Docker Compose are available.

3. **Start development environment** (recommended for development):

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

4. **Or run full Docker Compose setup** (for production-like testing):

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

## Documentation

For detailed documentation, see the [docs](./docs/) folder:

- **[Architecture](./docs/architecture.md)** - System architecture, diagrams, and project structure
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

# Run linter and tests
make check
```

See [Development Guide](./docs/development.md) for complete list of commands.

## Project Structure

```
real-time-leaderboard/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── shared/         # Shared utilities and infrastructure
│   └── module/         # Self-contained modules (auth, score, leaderboard, report)
├── docs/               # Documentation
├── scripts/            # Utility scripts (init.sh, run.sh, migrate.sh)
└── docker/             # Docker configuration
    ├── Dockerfile                    # Production Docker image
    ├── docker-compose.deps.yml       # Dependency services (postgres, redis)
    └── docker-compose.yml            # Full compose file (includes deps + app)
```

For detailed project structure, see [Architecture Documentation](./docs/architecture.md).

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
