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
    
    subgraph scoreModule[Score Module - Self Contained]
        ScoreAdapter[Score Adapters HTTP]
        ScoreApplication[Score Application Use Cases]
        ScoreDomain[Score Domain Entities]
        ScoreDomainRepo[Score Repository Interface]
        ScoreInfra[Score Infrastructure]
    end
    
    subgraph leaderboardModule[Leaderboard Module - Self Contained]
        LeaderboardAdapter[Leaderboard Adapters HTTP]
        LeaderboardApplication[Leaderboard Application Use Cases]
        LeaderboardDomain[Leaderboard Domain Entities]
        LeaderboardDomainRepo[Leaderboard Repository Interface]
        LeaderboardInfra[Leaderboard Infrastructure]
    end
    
    subgraph reportModule[Report Module - Self Contained]
        ReportAdapter[Report Adapters HTTP]
        ReportApplication[Report Application Use Cases]
        ReportDomain[Report Domain Entities]
        ReportDomainRepo[Report Repository Interface]
        ReportInfra[Report Infrastructure]
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
    Router --> ReportAdapter
    
    AuthAdapter --> AuthApplication
    AuthApplication --> AuthDomain
    AuthApplication --> AuthDomainRepo
    AuthDomainRepo --> AuthInfra
    AuthInfra --> Database
    Database --> Postgres
    
    ScoreAdapter --> ScoreApplication
    ScoreApplication --> ScoreDomain
    ScoreApplication --> ScoreDomainRepo
    ScoreDomainRepo --> ScoreInfra
    ScoreInfra --> Database
    ScoreInfra --> RedisClient
    Database --> Postgres
    RedisClient --> Redis
    
    LeaderboardAdapter --> LeaderboardApplication
    LeaderboardApplication --> LeaderboardDomain
    LeaderboardApplication --> LeaderboardDomainRepo
    LeaderboardDomainRepo --> LeaderboardInfra
    LeaderboardInfra --> RedisClient
    RedisClient --> Redis
    
    ReportAdapter --> ReportApplication
    ReportApplication --> ReportDomain
    ReportApplication --> ReportDomainRepo
    ReportDomainRepo --> ReportInfra
    ReportInfra --> RedisClient
    ReportInfra --> Database
    RedisClient --> Redis
    Database --> Postgres
    
    AuthAdapter --> Middleware
    ScoreAdapter --> Middleware
    LeaderboardAdapter --> Middleware
    ReportAdapter --> Middleware
    
    AuthApplication --> Response
    ScoreApplication --> Response
    LeaderboardApplication --> Response
    ReportApplication --> Response
    
    AuthAdapter --> Response
    ScoreAdapter --> Response
    LeaderboardAdapter --> Response
    ReportAdapter --> Response
    
    AuthApplication --> Logger
    ScoreApplication --> Logger
    LeaderboardApplication --> Logger
    ReportApplication --> Logger
```

## Data Flow Example

### Score Submission and Real-Time Leaderboard Updates

```mermaid
sequenceDiagram
    participant Client
    participant ScoreAdapter
    participant ScoreApplication
    participant ScoreInfra
    participant Redis
    participant LeaderboardAdapter
    participant SSEClient
    
    Note over Client,SSEClient: Score Submission Flow
    Client->>ScoreAdapter: POST /api/v1/scores
    ScoreAdapter->>ScoreApplication: SubmitScoreUseCase.Execute(score)
    ScoreApplication->>ScoreInfra: SaveScore(score)
    ScoreInfra->>Postgres: INSERT INTO scores
    ScoreApplication->>Redis: ZADD leaderboard:global score userId
    ScoreApplication->>Redis: ZADD leaderboard:{gameId} score userId
    ScoreApplication->>Redis: PUBLISH leaderboard:updates:global "updated"
    ScoreApplication->>Redis: PUBLISH leaderboard:updates:{gameId} "updated"
    ScoreApplication-->>ScoreAdapter: Score entity
    ScoreAdapter-->>Client: 200 OK (Score entity as JSON)
    
    Note over SSEClient,Redis: Real-Time Leaderboard Updates (SSE)
    SSEClient->>LeaderboardAdapter: GET /api/v1/leaderboard/global (SSE)
    LeaderboardAdapter->>Redis: SUBSCRIBE leaderboard:updates:global
    LeaderboardAdapter->>Redis: ZREVRANGE leaderboard:global 0 99
    LeaderboardAdapter-->>SSEClient: data: {leaderboard JSON}\n\n (initial)
    
    Note over Redis,SSEClient: When score is submitted...
    Redis-->>LeaderboardAdapter: PUBLISH notification received
    LeaderboardAdapter->>Redis: ZREVRANGE leaderboard:global 0 99
    LeaderboardAdapter-->>SSEClient: data: {updated leaderboard JSON}\n\n
```

**Key Points**:
- Score module updates Redis sorted sets and publishes notifications
- Leaderboard SSE handlers subscribe to Redis pub/sub channels
- No polling - updates are pushed immediately when scores change
- No cross-module dependencies - communication via Redis (shared infrastructure)

### Detailed Flow

1. **Score Submission**:
   - User submits score via `POST /api/v1/scores`
   - Score is saved to PostgreSQL for historical records
   - Score is added/updated in Redis sorted sets (`ZADD` command)
   - Notification is published to Redis pub/sub channels (`PUBLISH` command)

2. **Real-Time Updates**:
   - SSE clients connect to leaderboard endpoints (`GET /api/v1/leaderboard/global` or `/game/:game_id`)
   - SSE handlers subscribe to corresponding Redis pub/sub channels (`SUBSCRIBE` command)
   - Initial leaderboard data is fetched from Redis and sent to client
   - When a notification is received from pub/sub:
     - Handler fetches fresh leaderboard data from Redis sorted sets (`ZREVRANGE` command)
     - Updated leaderboard is sent to client via SSE (`data: {json}\n\n` format)

3. **Benefits**:
   - **Real-time**: Updates pushed immediately when scores change (no polling delay)
   - **Efficient**: Only fetches data when there's an actual update
   - **Decoupled**: Score and leaderboard modules communicate via Redis (shared infrastructure)
   - **Scalable**: Works across multiple server instances - all instances receive pub/sub notifications

## Architecture Principles

### Clean Architecture Layers

Each module follows Clean Architecture with four distinct layers:

1. **Domain Layer** (`domain/`): Contains core business entities and repository interfaces. This layer has no external dependencies and represents the business rules.

2. **Application Layer** (`application/`): Contains use cases that orchestrate business logic. It depends only on the domain layer and defines interfaces for infrastructure.

3. **Adapters Layer** (`adapters/`): Contains HTTP handlers that translate external requests into domain entities and use cases. This is the input adapter layer.

4. **Infrastructure Layer** (`infrastructure/`): Contains implementations of repositories and external service integrations (database, Redis, etc.). This is the output adapter layer.

### Module Independence

Each module (auth, score, leaderboard, report) is self-contained with its own:
- Domain entities and business rules
- Use cases and application logic
- HTTP adapters
- Infrastructure implementations

This design allows each module to be extracted into a separate microservice if needed. See [Microservice Migration Guide](./microservice-migration.md) for details.

### Shared Components

The `internal/shared/` directory contains cross-cutting concerns used by all modules:
- **Response**: Standardized API responses and error handling
- **Middleware**: HTTP middleware (authentication, logging, recovery)
- **Logger**: Centralized logging
- **Validator**: Request validation utilities
- **Database**: PostgreSQL connection and migrations
- **Redis**: Redis client connection

These shared components follow the dependency inversion principle - modules depend on abstractions, not concrete implementations.

