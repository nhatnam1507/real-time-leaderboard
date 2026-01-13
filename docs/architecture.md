# Architecture

This document describes the architectural principles, project structure, and coding conventions used in this project.

## Clean Architecture Principles

The system follows **Clean Architecture** principles to achieve maintainability, testability, and independence of business logic from frameworks and infrastructure.

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
│   ├── run.sh                     # Application startup script (dev/all modes)
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
│   ├── docker-compose.deps.yml    # Dependency services (postgres, redis)
│   └── docker-compose.yml         # Full compose file (includes deps + app)
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

Each module is self-contained and follows a consistent structure with four layers:

```
module/
├── domain/          # Business entities and domain services (if needed)
├── application/     # Use cases, business logic orchestration, and repository interfaces
├── adapters/        # HTTP handlers and external interface adapters
└── infrastructure/  # Repository implementations and external services
```

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
- Domain entities (business concepts only, no infrastructure concerns)
- Domain services (if needed for complex business logic)
- Domain constants (e.g., Redis keys, topics)

**Rules**:
- **Zero external dependencies** - no imports from other layers
- No database IDs, timestamps, or infrastructure concerns
- No repository interfaces - these belong to the application layer
- Pure Go code - no framework dependencies
- Contains only core business objects and domain logic

### Application Layer (`application/`)

**Purpose**: Orchestrates business logic and use cases

**Contains**:
- Use case structs and methods
- Repository interfaces (application needs, not domain concepts)
- Business logic orchestration
- Data enrichment (combining data from multiple sources)

**Rules**:
- Depends only on domain layer
- Defines repository interfaces that represent application needs
- Defines service interfaces for external systems (pub/sub, messaging, etc.)
- No direct infrastructure access - uses repository and service interfaces
- Contains business logic, not infrastructure details
- Can call multiple repositories to compose results

### Adapters Layer (`adapters/`)

**Purpose**: Translates external requests into domain operations

**Contains**:
- HTTP handlers (REST API)
- Request/response transformation
- Connection lifecycle management (e.g., SSE)

**Rules**:
- No business logic - delegates to application layer
- Handles protocol-specific concerns (HTTP, SSE, WebSocket)
- Transforms external formats to domain entities
- Manages connection lifecycle only

### Infrastructure Layer (`infrastructure/`)

**Purpose**: Implements technical details and external integrations

**Contains**:
- Repository implementations (PostgreSQL, Redis)
- Service implementations (Redis pub/sub, messaging, etc.)
- DTOs (Data Transfer Objects) for database operations
- External service clients

**Rules**:
- Implements application repository interfaces
- Implements application service interfaces (e.g., BroadcastService)
- Uses DTOs internally (with `db` tags, not `json` tags)
- Maps DTOs to domain objects when returning
- Database concerns (IDs, timestamps) stay here
- Infrastructure-specific details (Redis commands, pub/sub) stay here

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

- **Domain entities**: One file per entity or related entities (e.g., `leaderboard.go`)
- **Repository interfaces**: Grouped in `application/repository.go` or split by concern
- **Use cases**: One file per use case or related use cases
- **Handlers**: One file per handler or version group
- **Repository implementations**: One file per repository implementation

### Domain Layer Rules

1. **No Infrastructure Concerns**: Domain entities should represent pure business concepts, not database structures. Avoid database IDs, timestamps, or infrastructure-specific fields.
   ```go
   // ❌ Bad - contains database ID and timestamps
   type Score struct {
       ID        string    `json:"id"`
       CreatedAt time.Time  `json:"created_at"`
   }
   
   // ✅ Good - pure business concept
   type LeaderboardEntry struct {
       UserID   string `json:"user_id"`
       Username string `json:"username"`
       Score    int64  `json:"score"`
       Rank     int64  `json:"rank"`
   }
   ```

2. **Repository Interfaces in Application Layer**: Repository interfaces are defined in the application layer (representing application needs), not in the domain layer. They return domain objects, not DTOs.
   ```go
   // ✅ Good - repository interface in application layer, returns domain object
   // application/repository.go
   type LeaderboardRepository interface {
       GetTopPlayers(ctx context.Context, limit, offset int64) ([]domain.LeaderboardEntry, error)
   }
   
   // ❌ Bad - returns DTO
   type LeaderboardRepository interface {
       GetTopPlayers(ctx context.Context, limit, offset int64) ([]ScoreDTO, error)
   }
   ```

3. **Constants in Domain**: Domain constants (like Redis keys, topics) belong in the domain layer as they represent business concepts.
   ```go
   // ✅ Good - domain constants
   const (
       RedisLeaderboardKey = "leaderboard:global"
       RedisScoreUpdateTopic = "leaderboard:score:updates"
   )
   ```

### Infrastructure Layer Rules

1. **Use DTOs with Database Tags**: Infrastructure layer uses Data Transfer Objects (DTOs) with database tags (`db:`) for database operations. These are internal to the infrastructure layer.
   ```go
   // ✅ Good - DTO with db tags
   type Score struct {
       ID        string    `db:"id"`
       UserID    string    `db:"user_id"`
       Point     int64     `db:"point"`
       CreatedAt time.Time `db:"created_at"`
   }
   ```

2. **Map DTOs to Domain Objects**: Infrastructure implementations map DTOs to domain objects when returning data to the application layer. Domain objects are the contract between layers.
   ```go
   // ✅ Good - maps DTO to domain
   func (r *PostgresRepository) GetLeaderboard(ctx context.Context) (*domain.Leaderboard, error) {
       // Query with DTO
       var dto ScoreDTO
       // ... scan into dto
       
       // Map to domain
       return &domain.Leaderboard{
           Entries: mapToDomain(dto),
       }, nil
   }
   ```

3. **Keep DTOs Private**: DTOs should be unexported (lowercase) as they are internal implementation details of the infrastructure layer.
   ```go
   // ✅ Good - DTO is internal to infrastructure
   type score struct {  // lowercase = unexported
       ID string `db:"id"`
   }
   ```

4. **Implement Service Interfaces**: Infrastructure implements service interfaces (like `BroadcastService`) defined in the application layer, handling all infrastructure-specific details (Redis commands, pub/sub connections, etc.).

### Application Layer Rules

1. **Enrich Domain Objects**: Application layer orchestrates data from multiple sources to enrich domain objects. For example, combining leaderboard rankings with user information.
   ```go
   // ✅ Good - enriches domain objects with data from multiple sources
   func (uc *UseCase) GetFullLeaderboard(ctx context.Context) (*domain.Leaderboard, error) {
       entries, _ := uc.leaderboardRepo.GetTopPlayers(ctx, 1000, 0)
       usernames, _ := uc.userRepo.GetByIDs(ctx, extractUserIDs(entries))
       enrichEntries(entries, usernames)
       return &domain.Leaderboard{Entries: entries}, nil
   }
   ```

2. **Orchestrate, Don't Implement**: Use cases orchestrate business logic by calling multiple repositories/services. They don't implement low-level details - those belong in infrastructure.
   ```go
   // ✅ Good - orchestrates business logic and coordinates services
   func (uc *ScoreUseCase) SubmitScore(ctx context.Context, userID string, req SubmitScoreRequest) error {
       // Update data in repositories
       if err := uc.backupRepo.UpsertScore(ctx, userID, req.Score); err != nil {
           return err
       }
       if err := uc.leaderboardRepo.UpdateScore(ctx, userID, req.Score); err != nil {
           return err
       }
       // Publish notification via broadcast service (not repository)
       return uc.broadcastService.PublishScoreUpdate(ctx)
   }
   ```

3. **Define Service Interfaces for External Systems**: Service interfaces (like `BroadcastService` for pub/sub) are defined in the application layer. Use cases depend on these interfaces, not concrete implementations. Infrastructure implements these interfaces.
   ```go
   // ✅ Good - service interface in application layer
   // application/broadcast_service.go
   type BroadcastService interface {
       PublishScoreUpdate(ctx context.Context) error  // Publish notifications
       SubscribeToScoreUpdates(ctx context.Context) (<-chan struct{}, error)
       BroadcastLeaderboard(ctx context.Context, leaderboard *domain.Leaderboard) error
       SubscribeToLeaderboardUpdates(ctx context.Context) (<-chan *domain.Leaderboard, error)
   }
   
   // ✅ Good - use case uses interface, not concrete implementation
   type ScoreUseCase struct {
       broadcastService BroadcastService  // interface, not *RedisBroadcastService
       leaderboardRepo LeaderboardRepository
       // ... other dependencies
   }
   
   // ✅ Good - repository only updates data, use case publishes notifications
   func (uc *ScoreUseCase) SubmitScore(...) error {
       uc.leaderboardRepo.UpdateScore(...)  // Only updates data
       uc.broadcastService.PublishScoreUpdate(...)  // Publishes notification
   }
   
   // Infrastructure implements the interface
   // infrastructure/broadcast/redis_broadcast_service.go
   type RedisBroadcastService struct { ... }
   func (s *RedisBroadcastService) PublishScoreUpdate(...) error { ... }
   ```

### Adapters Layer Rules

1. **Delegate to Application Layer**: Adapters delegate all business logic to the application layer (use cases). They handle protocol-specific concerns only.
   ```go
   // ✅ Good - delegates to use case
   func (h *Handler) SubmitScore(c *gin.Context) {
       if err := h.useCase.SubmitScore(ctx, userID, req); err != nil {
           response.Error(c, err)
           return
       }
       response.Success(c, data, "Success")
   }
   ```

2. **Handle Protocol Concerns Only**: Adapters manage connection lifecycles (SSE, WebSocket), request/response transformation, and protocol-specific headers. Business logic stays in the application layer.
   ```go
   // ✅ Good - handles SSE connection lifecycle
   func (h *Handler) GetLeaderboard(c *gin.Context) {
       c.Header("Content-Type", "text/event-stream")
       updateCh := h.useCase.SubscribeToLeaderboardUpdates(ctx)
       // ... manage connection
   }
   ```

### Error Handling

- Use domain-specific errors in application layer
- Use shared error helpers from `internal/shared/response`
- Log errors with context in application layer
- Return errors, don't panic (except in unrecoverable situations)

### Dependency Injection

- All dependencies injected via constructors
- Repository interfaces defined in application layer, implementations in infrastructure
- Wire dependencies in `cmd/server/main.go`
- Enables easy testing with mocks

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
