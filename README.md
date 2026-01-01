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

2. **Copy environment variables**:
```bash
cp .env.example .env
```

3. **Start services with Docker Compose**:
```bash
make docker-up
# or
docker-compose up -d
```

This will start:
- PostgreSQL on port 5432
- Redis on port 6379
- Application on port 8080

4. **Run database migrations**:
```bash
make install-tools  # Install migrate tool if needed
make migrate-up
```

5. **Run the application**:
```bash
make run
# or
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### Quick Development Setup

For a complete development environment setup:

```bash
make dev
```

This will:
1. Start Docker services (PostgreSQL, Redis)
2. Wait for services to be ready
3. Run database migrations
4. Start the application

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

# Run tests
make test

# Run linter
make lint

# Build application
make build

# Start Docker services
make docker-up

# Run migrations
make migrate-up
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
├── scripts/            # Utility scripts
└── docker/             # Docker configuration
```

For detailed project structure, see [Architecture Documentation](./docs/architecture.md).

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
