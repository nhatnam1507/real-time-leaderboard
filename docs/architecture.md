# Architecture

This document describes the architectural principles, project structure, and coding conventions used in this project.

## Clean Architecture Principles

The system follows **Clean Architecture** principles to achieve maintainability, testability, and independence of business logic from frameworks and infrastructure.
In Clean Architecture:
- Interfaces are defined inward
- Mocks live outward
- Dependencies point inward only
That rule alone prevents 90% of cyclic imports.

### Core Principles

1. **Dependency Rule**: Source code dependencies point inward. The domain layer has no dependencies on other layers.
2. **Independence**: Business logic is independent of frameworks, databases, and external services.
3. **Testability**: Business logic can be tested without external dependencies.
4. **Framework Independence**: Business logic doesn't depend on web frameworks, databases, or UI frameworks.

## Project Structure

The project is organized following Clean Architecture principles with a modular structure:

```
real-time-leaderboard/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── migrations/                     # Database migrations
│   ├── schema/                     # Core schema migrations (all environments)
│   └── dev/                        # Dev-only seed data migrations
├── internal/
│   ├── config/                     # Configuration management
│   │   └── config.go
│   ├── shared/                     # Shared utilities and infrastructure
│   │   ├── response/               # API response helpers and error definitions
│   │   ├── middleware/             # HTTP middleware
│   │   ├── logger/                 # Logger implementation
│   │   ├── validator/              # Request validation
│   │   ├── database/               # Database connections
│   │   └── redis/                  # Redis connections
│   └── module/                     # Self-contained modules
│       ├── auth/                   # Auth Module
│       │   ├── domain/            # Domain layer
│       │   ├── application/       # Application layer
│       │   ├── adapters/          # Adapters layer
│       │   └── infrastructure/    # Infrastructure layer
│       └── leaderboard/            # Leaderboard Module (score update + leaderboard)
│           ├── domain/            # Domain layer
│           ├── application/       # Application layer
│           ├── adapters/          # Adapters layer
│           └── infrastructure/    # Infrastructure layer
├── docs/                           # Documentation
├── scripts/                        # Utility scripts (shell scripts)
│   ├── init.sh                    # Initialize development environment (dev/ci modes)
│   ├── run.sh                     # Application startup script (dev/prod-like modes)
│   ├── migrate.sh                 # Database migration tool
│   └── validate-workflows.sh      # GitHub Actions workflow validation
├── tools/                          # Build tools (Go tools)
│   └── generate-openapi-json.go   # OpenAPI YAML to JSON converter
├── .githooks/                     # Git hooks (version controlled)
│   └── pre-push                   # Pre-push hook for code quality checks
├── .github/
│   ├── workflows/                 # GitHub Actions workflows
│   │   ├── pr.yml                 # PR workflow (lint + unit tests)
│   │   └── ci.yml                 # CI workflow (lint + unit tests + dockerize)
│   └── actions/                   # Reusable GitHub Actions
│       └── init/                  # Init action (Go setup + make init-ci)
├── docker/
│   ├── Dockerfile                 # Production Docker image
│   ├── docker-compose.dev.yml     # Dependency services for development (postgres, redis)
│   └── docker-compose.swarm.yml   # Full stack for Docker Swarm (production-like)
├── api/                           # OpenAPI 3.0 specifications
│   ├── v1/                        # API v1 specification
│   │   └── openapi.yaml           # OpenAPI 3.0 YAML specification
│   └── swagger-ui.html            # Swagger UI for viewing OpenAPI docs
├── .air.toml                      # Air configuration for hot reload
├── .golangci.yml                  # golangci-lint configuration
├── go.mod
├── Makefile                       # Development commands
└── README.md
```

## Module Organization

Each module is self-contained with four layers (see [Project Structure](#project-structure) above for full structure):

**Key Points**:
- Each module is independent - no cross-module imports
- Layers depend inward: Adapters → Application → Domain ← Infrastructure
- Domain layer has zero external dependencies - contains only core business objects
- Repository interfaces are defined in application layer (application needs, not domain concepts)
- Infrastructure implements application repository interfaces

## Layer Structure & Responsibilities

The architecture consists of four concentric layers, each with specific responsibilities:

```
┌─────────────────────────────────────┐
│      Adapters Layer (Input)         │  ← HTTP handlers, external interfaces
├─────────────────────────────────────┤
│      Application Layer              │  ← Use cases, business logic orchestration
├─────────────────────────────────────┤
│      Domain Layer                   │  ← Entities (core business objects, no deps)
├─────────────────────────────────────┤
│      Infrastructure Layer (Output)  │  ← Repository implementations, external services
└─────────────────────────────────────┘
```

**Dependency Direction**: Adapters → Application → Domain ← Infrastructure

### Domain Layer (`domain/`)

**Purpose**: Contains pure business logic and core business objects

**Contains**: 
- Domain entities (one file per entity, no timestamps/infrastructure fields)
- Domain errors (sentinel errors)
- Domain constants (Redis keys, topics)

**Rules**:
- Zero external dependencies
- No repository interfaces (belong in application layer)
- Pure business objects only

### Application Layer (`application/`)

**Purpose**: Orchestrates business logic and use cases

**Contains**:
- Use cases, repository interfaces, service interfaces (pub/sub, messaging)
- Business logic orchestration and data enrichment

**Rules**:
- Depends only on domain layer
- Defines interfaces (repositories, services) - implemented in infrastructure
- Orchestrates business logic via interfaces

### Adapters Layer (`adapters/`)

**Purpose**: Translates external requests into domain operations

**Contains**:
- HTTP handlers, error mappers, request/response transformation
- Connection lifecycle management (SSE, WebSocket)

**Rules**:
- Delegates business logic to application layer
- Handles protocol concerns only (HTTP headers, SSE lifecycle)
- Maps errors to APIError via module-specific error mappers

### Infrastructure Layer (`infrastructure/`)

**Purpose**: Implements technical details and external integrations

**Contains**:
- Repository implementations, service implementations (pub/sub, messaging)
- DTOs (with `db` tags, unexported)

**Rules**:
- Implements application layer interfaces
- Maps DTOs to domain objects when returning
- Database concerns (IDs, timestamps) stay here

## Dependency Rules

1. **Domain Independence**: Domain layer has zero dependencies on other layers - contains only core business objects
2. **Interface Segregation**: Repository interfaces are defined in application layer (application needs), implemented in infrastructure
3. **Dependency Inversion**: High-level modules (application) depend on abstractions (repository interfaces), not concrete implementations
4. **Module Independence**: Each module owns its interfaces - no cross-module dependencies

## Shared Components

The `internal/shared/` directory provides cross-cutting concerns:
- **Response**: Standardized API responses and error handling
- **Middleware**: HTTP middleware (auth, logging, recovery, CORS)
- **Logger**: Centralized structured logging
- **Validator**: Request validation utilities
- **Database**: PostgreSQL connection and migrations
- **Redis**: Redis client connection

These follow dependency inversion - modules depend on abstractions, not concrete implementations.

## Coding Conventions

### Naming Conventions

- **Packages**: Lowercase, single word, descriptive (e.g., `domain`, `application`, `repository`)
- **Interfaces**: Descriptive names ending with the role (e.g., `UserRepository`, `LeaderboardRepository`)
- **Structs**: PascalCase (e.g., `LeaderboardEntry`, `ScoreUseCase`)
- **Functions**: PascalCase for exported, camelCase for unexported
- **Constants**: PascalCase with descriptive names (e.g., `RedisLeaderboardKey`)

### File Organization

**Naming Conventions**:
- **Domain entities**: One file per entity (e.g., `user.go`, `token.go`, `leaderboard_entry.go`)
- **Domain errors**: `errors.go` in domain layer
- **Repository interfaces**: Grouped in `application/repository.go` or split by concern
- **Use cases**: One file per use case (e.g., `auth_usecase.go`, `leaderboard_usecase.go`, `score_usecase.go`)
- **Handlers**: One file per handler or version group (e.g., `handler.go`)
- **Error mappers**: `error_mapper.go` in adapters layer (module-specific error mapping)
- **Repository implementations**: One file per repository implementation
- **DTOs**: One file per DTO or related DTOs (e.g., `user_dto.go`)

**File Structure Example**:
```
module/
├── domain/
│   ├── user.go           # User entity (no timestamps)
│   ├── token.go          # TokenPair entity
│   ├── errors.go         # Domain errors (sentinel errors)
│   └── constants.go      # Domain constants (Redis keys, topics)
├── application/
│   ├── auth_usecase.go   # Use case (returns wrapped errors)
│   ├── repository.go     # Repository interfaces
│   └── auth_usecase_test.go  # Tests
├── adapters/
│   ├── rest/v1/
│   │   ├── handler.go    # HTTP handlers (logs errors, maps to APIError)
│   │   └── error_mapper.go  # Error mapping (domain errors → APIError)
│   └── mocks/            # Mocks for use case interfaces
│       └── auth_usecase_mock.go
└── infrastructure/
    ├── repository/
    │   ├── postgres.go   # PostgreSQL implementation
    │   └── user_dto.go   # DTOs (with timestamps, db tags)
    ├── jwt/              # Service implementations
    │   └── jwt.go
    └── mocks/            # Mocks for repository/service interfaces
        ├── user_repository_mock.go
        └── jwt_manager_mock.go
```

### Domain Layer Rules

**No Infrastructure Concerns**: Domain entities represent pure business concepts.
```go
// ❌ Bad - contains timestamps (infrastructure concern)
type User struct {
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// ✅ Good - pure business concept
type User struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}
```

Timestamps are handled in infrastructure DTOs only.

**Repository Interfaces in Application Layer**: Interfaces defined in application layer, return domain objects.
```go
// ✅ Good - application layer interface, returns domain
type LeaderboardRepository interface {
    GetTopPlayers(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, error)
}
```

**Constants in Domain**: Business constants (Redis keys, topics) belong in domain layer.

### Infrastructure Layer Rules

**DTOs**: Use DTOs with `db` tags (unexported), map to domain objects when returning.
```go
// ✅ Good - DTO with db tags, unexported
type score struct {
    ID        string    `db:"id"`
    CreatedAt time.Time `db:"created_at"`
}

// Map to domain when returning
func (r *Repo) GetLeaderboard(...) ([]domain.LeaderboardEntry, error) {
    // ... query with DTO
    return mapToDomain(dto), nil
}
```

**Service Interfaces**: Implements application layer service interfaces (e.g., `BroadcastService`).

### Application Layer Rules

**Orchestrate Business Logic**: Use cases orchestrate by calling multiple repositories/services.
```go
// ✅ Good - orchestrates, enriches domain objects
func (uc *UseCase) GetLeaderboard(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, int64, error) {
    entries, _ := uc.cacheRepo.GetTopPlayers(ctx, limit, offset)
    usernames, _ := uc.userRepo.GetByIDs(ctx, extractUserIDs(entries))
    enrichEntries(entries, usernames)
    return entries, total, nil
}
```

**Service Interfaces**: Define service interfaces (e.g., `BroadcastService`) in application layer, implemented in infrastructure.

### Adapters Layer Rules

**Delegate to Application**: Handlers delegate business logic to use cases, handle protocol concerns only.
```go
// ✅ Good - delegates, maps errors, handles protocol
func (h *Handler) SubmitScore(c *gin.Context) {
    if err := h.useCase.SubmitScore(ctx, userID, req); err != nil {
        apiErr := toAPIError(err)  // Module-specific mapper
        h.logger.Err(ctx, err).Msg("Request error")
        response.Error(c, apiErr)
        return
    }
    response.Success(c, data, "Success")
}
```

**Error Mapping**: Each module has `error_mapper.go` in adapters layer (domain/validation errors → APIError).

### Error Handling

**Strategy**:
- **Application**: Returns wrapped errors (`fmt.Errorf("...: %w", err)`) - no `APIError`
- **Adapters**: Logs errors, maps to `APIError` via module-specific error mappers
- **Domain**: Sentinel errors (e.g., `ErrUserNotFound`) in `domain/errors.go`
- **Validator**: Returns `ValidationError` - mapped to `APIError` in adapters

**Error Types**:
- `APIError` (`internal/shared/response/error.go`) - API error response
- `ValidationError` (`internal/shared/validator/validator.go`) - Validation error
- Domain errors (`domain/errors.go`) - Business logic errors

### Dependency Injection

**Constructor Pattern**: Accept interfaces, return structs (use cases) or interfaces (repositories).
```go
// ✅ Use case: accepts interfaces, returns concrete struct
func NewLeaderboardUseCase(
    cacheRepo LeaderboardCacheRepository,  // interface
    logger *logger.Logger,
) *leaderboardUseCase {                    // concrete struct
    return &leaderboardUseCase{...}
}

// ✅ Repository: returns interface
func NewPostgresRepository(pool *pgxpool.Pool) application.UserRepository {
    return &PostgresUserRepository{pool: pool}
}
```

Dependencies wired in `cmd/server/main.go`. Enables easy testing with mocks.

### Mock Organization

**Core Principle**: Mocks belong to the layer that **depends on** the interface, not where it's defined.

**Rules**:
- **Infrastructure layer mocks**: Interfaces used by infrastructure implementations (repositories, services) → `infrastructure/mocks/`
- **Adapters layer mocks**: Interfaces used by adapters/handlers (use cases) → `adapters/mocks/`
- Mocks live **outward** from where interfaces are defined
- Prevents import cycles (mocks don't need to import application layer)

See [File Structure Example](#file-organization) above for complete structure.

## Module Independence

Each module is self-contained:

- **Own Domain**: Defines its own entities (core business objects)
- **No Cross-Module Dependencies**: Modules don't import from other modules
- **Own Interfaces**: If a module needs user data, it defines its own `UserRepository` interface in the application layer
- **Extractable**: Each module can be extracted to a separate microservice

This design ensures:
- Modules can evolve independently
- Easy to extract to microservices
- Clear boundaries and responsibilities
- Testable in isolation

For module-specific details, see [Modules](./modules.md).
