# Real-Time Leaderboard System

A high-performance, modular real-time leaderboard system built with Go, following Clean Architecture principles. The system provides real-time score updates and leaderboard rankings with live updates via Server-Sent Events (SSE).

## Overview

This system enables users to:
- **Authenticate** using JWT-based authentication with access and refresh tokens
- **Submit scores** that are automatically ranked in a global leaderboard
- **View leaderboard** with real-time updates as scores change
- **Experience low latency** through Redis caching and efficient data structures

The system is designed with **Clean Architecture** principles, making it:
- **Modular**: Self-contained modules that can be extracted to microservices
- **Testable**: Business logic independent of frameworks and infrastructure
- **Maintainable**: Clear separation of concerns across layers
- **Scalable**: Supports multiple server instances with distributed coordination

## Features

- JWT-based authentication with token refresh
- Real-time leaderboard updates via Server-Sent Events (SSE)
- High-performance Redis caching with PostgreSQL persistence
- Modular Clean Architecture design
- Microservice-ready modules

See [Application Features & Flows](./docs/application.md) for detailed features.

## Prerequisites

- Go 1.25.5 or later
- Docker and Docker Compose

> **Note**: PostgreSQL and Redis are automatically managed via Docker. See [Development Guide](./docs/development.md) for technology stack details.

## Installation

1. Clone the repository
2. Initialize development environment:
```bash
make init-dev
```

This installs required tools and configures the environment. See [Development Guide](./docs/development.md) for details.

## Quick Start

### Development Mode (Recommended)

```bash
make start-dev
```

Starts dependencies and the application with hot reload. Press `Ctrl+C` to stop.

### Production-like Mode

```bash
make run
```

Runs the full stack in Docker Swarm. Use `make stop` to stop.

The server starts on `http://localhost:8080`. See [Development Guide](./docs/development.md) for detailed information.

## Documentation

- **[Application Features & Flows](./docs/application.md)** - Application overview and flows
- **[Architecture](./docs/architecture.md)** - Architecture principles and project structure
- **[Modules](./docs/modules.md)** - Module documentation and flows
- **[Development Guide](./docs/development.md)** - Development setup and commands
- **[Microservice Migration](./docs/microservice-migration.md)** - Module extraction guide

**API Documentation**: http://localhost:8080/docs/index.html (Swagger UI)

## Common Commands

```bash
make help          # Show all available commands
make init-dev      # Initialize development environment
make start-dev     # Start development (hot reload)
make run           # Run full Docker Swarm stack
make stop          # Stop Docker Swarm stack
make clean         # Remove all Docker resources
make check         # Run all checks (lint, tests, etc.)
```

See [Development Guide](./docs/development.md) for complete command documentation.

## Project Structure

The project follows Clean Architecture with self-contained modules. See [Architecture](./docs/architecture.md) for detailed project structure and layer organization.

## Contributing

Contributions are welcome! Before contributing:
1. Read the [Development Guide](./docs/development.md)
2. Follow the [Architecture](./docs/architecture.md) principles
3. Run `make check` to ensure all checks pass

## License

MIT
