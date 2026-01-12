# Architecture

The system follows **Clean Architecture** principles with clear layer separation:

- **Domain Layer**: Core entities and repository interfaces (no dependencies)
- **Application Layer**: Use cases and business logic orchestration
- **Adapters Layer**: HTTP handlers (input adapters)
- **Infrastructure Layer**: Repository implementations, external services (output adapters)

## System Architecture Diagram

```mermaid
graph TB
    subgraph client[Client Applications]
        WebClient[Web Client]
        MobileClient[Mobile Client]
    end
    
    subgraph api[API Gateway Layer]
        Router[HTTP Router]
    end
    
    subgraph authModule[Auth Module - Self Contained]
        AuthAdapter[Auth Adapters HTTP]
        AuthApplication[Auth Application Use Cases]
        AuthDomain[Auth Domain Entities]
        AuthDomainRepo[Auth Repository Interface]
        AuthInfra[Auth Infrastructure]
    end
    
    subgraph leaderboardModule[Leaderboard Module - Self Contained]
        ScoreAdapter[Score Adapters HTTP]
        LeaderboardAdapter[Leaderboard Adapters HTTP]
        ScoreApplication[Score Application Use Cases]
        LeaderboardApplication[Leaderboard Application Use Cases]
        LeaderboardDomain[Leaderboard Domain Entities]
        BackupDomainRepo[Leaderboard Backup Repository Interface]
        LeaderboardDomainRepo[Leaderboard Repository Interface]
        BackupInfra[PostgreSQL Backup Infrastructure]
        LeaderboardInfra[Redis Leaderboard Infrastructure]
    end
    
    subgraph shared[Shared Components]
        Response[Response Helpers & Errors]
        Middleware[Middleware]
        Logger[Logger]
        Validator[Validator]
        Database[Database Connection]
        Redis[Redis Connection]
    end
    
    subgraph infra[External Services]
        Postgres[(PostgreSQL)]
        Redis[(Redis)]
    end
    
    WebClient --> Router
    MobileClient --> Router
    
    Router --> AuthAdapter
    Router --> ScoreAdapter
    Router --> LeaderboardAdapter
    
    AuthAdapter --> AuthApplication
    AuthApplication --> AuthDomain
    AuthApplication --> AuthDomainRepo
    AuthDomainRepo --> AuthInfra
    AuthInfra --> Database
    Database --> Postgres
    
    ScoreAdapter --> ScoreApplication
    ScoreApplication --> LeaderboardDomain
    ScoreApplication --> BackupDomainRepo
    BackupDomainRepo --> BackupInfra
    BackupInfra --> Database
    ScoreApplication --> LeaderboardDomainRepo
    LeaderboardDomainRepo --> LeaderboardInfra
    LeaderboardInfra --> RedisClient
    Database --> Postgres
    RedisClient --> Redis
    
    LeaderboardAdapter --> LeaderboardApplication
    LeaderboardApplication --> LeaderboardDomain
    LeaderboardApplication --> LeaderboardDomainRepo
    LeaderboardDomainRepo --> LeaderboardInfra
    LeaderboardInfra --> RedisClient
    RedisClient --> Redis
    
    AuthAdapter --> Middleware
    ScoreAdapter --> Middleware
    LeaderboardAdapter --> Middleware
    
    AuthApplication --> Response
    ScoreApplication --> Response
    LeaderboardApplication --> Response
    
    AuthAdapter --> Response
    ScoreAdapter --> Response
    LeaderboardAdapter --> Response
    
    AuthApplication --> Logger
    ScoreApplication --> Logger
    LeaderboardApplication --> Logger
```

## Data Flow Example

### Score Update and Real-Time Leaderboard Updates

```mermaid
sequenceDiagram
    participant Client
    participant ScoreAdapter
    participant ScoreApplication
    participant ScoreInfra
    participant Redis
    participant LeaderboardAdapter
    participant SSEClient
    
    Note over Client,SSEClient: Score Update Flow
    Client->>ScoreAdapter: PUT /api/v1/score
    ScoreAdapter->>ScoreApplication: SubmitScoreUseCase.Execute(score)
    ScoreApplication->>BackupInfra: UpsertScore(userId, point)
    BackupInfra->>Postgres: INSERT ... ON CONFLICT (upsert)
    ScoreApplication->>Redis: ZADD leaderboard:global point userId
    ScoreApplication->>Redis: PUBLISH leaderboard:updates "updated"
    ScoreApplication-->>ScoreAdapter: Score entity
    ScoreAdapter-->>Client: 200 OK (Score entity as JSON)
    
    Note over Client,SSEClient: User Registration Flow
    Client->>AuthAdapter: POST /api/v1/auth/register
    AuthAdapter->>AuthApplication: RegisterUseCase.Execute()
    AuthApplication->>AuthInfra: CreateUser()
    AuthInfra->>Postgres: INSERT INTO users
    
    Note over SSEClient,Redis: Real-Time Leaderboard Updates (SSE)
    SSEClient->>LeaderboardAdapter: GET /api/v1/leaderboard?limit=50 (SSE)
    LeaderboardAdapter->>LeaderboardApplication: SyncFromPostgres(ctx)
    LeaderboardApplication->>Redis: ZCARD leaderboard:global
    alt Redis is empty
        LeaderboardApplication->>BackupInfra: GetAllScores()
        BackupInfra->>Postgres: SELECT * FROM leaderboard
        BackupInfra-->>LeaderboardApplication: scores[]
        loop For each score
            LeaderboardApplication->>Redis: ZADD leaderboard:global
        end
    end
    LeaderboardAdapter->>LeaderboardApplication: WatchLeaderboard(ctx, listReq)
    LeaderboardApplication->>Redis: ZREVRANGE leaderboard:global 0 49
    LeaderboardApplication-->>LeaderboardAdapter: channel <- Leaderboard (initial)
    LeaderboardAdapter-->>SSEClient: data: {leaderboard JSON}\n\n (initial)
    LeaderboardApplication->>Redis: SUBSCRIBE leaderboard:updates
    
    Note over Redis,SSEClient: When score is updated...
    Redis-->>LeaderboardApplication: PUBLISH notification received
    LeaderboardApplication->>Redis: ZREVRANGE leaderboard:global 0 49
    LeaderboardApplication-->>LeaderboardAdapter: channel <- Leaderboard (updated)
    LeaderboardAdapter-->>SSEClient: data: {updated leaderboard JSON}\n\n
```

**Key Points**:
- Score update upserts in PostgreSQL and updates Redis sorted sets, then publishes notifications
- Leaderboard SSE handlers subscribe to Redis pub/sub channel
- No polling - updates are pushed immediately when scores change
- Single leaderboard - no game-specific separation

### Detailed Flow

1. **User Registration**:
   - User registers via `POST /api/v1/auth/register`
   - User record is created in PostgreSQL

2. **Score Update**:
   - User updates score via `PUT /api/v1/score`
   - System upserts score in PostgreSQL (creates if not exists, updates if exists)
     - PostgreSQL serves as backup/recovery mechanism for Redis
     - Uses UPSERT pattern to handle both new and existing users automatically
   - Score is updated in Redis sorted sets (`ZADD` command)
   - Notification is published to Redis pub/sub channel (`PUBLISH leaderboard:updates`)

3. **Real-Time Updates**:
   - SSE clients connect to leaderboard endpoint (`GET /api/v1/leaderboard?limit=50`)
   - Handler calls `SyncFromPostgres()` for lazy loading - syncs PostgreSQL data to Redis if Redis is empty
   - Handler calls `WatchLeaderboard()` which:
     - Fetches initial leaderboard from Redis and sends to channel
     - Subscribes to Redis pub/sub channel (`SUBSCRIBE leaderboard:updates`) in application layer
     - On each pub/sub notification, fetches fresh leaderboard and sends to channel
   - Handler reads from channel and streams to client via SSE (`data: {json}\n\n` format)
   - Handler only manages SSE connection lifecycle - all business logic is in application layer

### Data Storage Strategy

- **PostgreSQL**: Stores one record per user with their current score
  - Uses UPSERT pattern - creates record if user doesn't exist, updates if exists
  - Serves as backup/recovery mechanism if Redis data is lost
  - Can be queried to rebuild Redis leaderboard if needed
  - **Lazy Loading**: On first leaderboard request, if Redis is empty, data is automatically synced from PostgreSQL

- **Redis**: Stores leaderboard in sorted sets for efficient real-time queries
  - Primary source of truth for leaderboard rankings
  - Provides O(log(N)) complexity for insertions and range queries
  - Automatically populated from PostgreSQL on first request if empty (lazy loading)
  - See [Redis Strategy](./redis-strategy.md) for detailed implementation

### Benefits

- **Real-time**: Updates pushed immediately when scores change (no polling delay)
- **Efficient**: Only fetches data when there's an actual update
- **Scalable**: Works across multiple server instances - all instances receive pub/sub notifications
- **Durable**: PostgreSQL backup ensures data can be recovered if Redis fails

## Architecture Principles

### Clean Architecture Layers

Each module follows Clean Architecture with four distinct layers:

1. **Domain Layer** (`domain/`): Contains core business entities and repository interfaces. This layer has no external dependencies and represents the business rules.

2. **Application Layer** (`application/`): Contains use cases that orchestrate business logic. It depends only on the domain layer and defines interfaces for infrastructure.

3. **Adapters Layer** (`adapters/`): Contains HTTP handlers that translate external requests into domain entities and use cases. This is the input adapter layer.

4. **Infrastructure Layer** (`infrastructure/`): Contains implementations of repositories and external service integrations (database, Redis, etc.). This is the output adapter layer.

### Module Independence

Each module (auth, leaderboard) is self-contained with its own:
- Domain entities and business rules
- Use cases and application logic
- HTTP adapters
- Infrastructure implementations

The leaderboard module combines score update and leaderboard retrieval functionality. This design allows each module to be extracted into a separate microservice if needed. See [Microservice Migration Guide](./microservice-migration.md) for details.

### Shared Components

The `internal/shared/` directory contains cross-cutting concerns used by all modules:
- **Response**: Standardized API responses and error handling
- **Middleware**: HTTP middleware (authentication, logging, recovery)
- **Logger**: Centralized logging
- **Validator**: Request validation utilities
- **Database**: PostgreSQL connection and migrations
- **Redis**: Redis client connection

These shared components follow the dependency inversion principle - modules depend on abstractions, not concrete implementations.

