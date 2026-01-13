# Modules

The system is organized into self-contained modules, each following Clean Architecture principles.

## Auth Module

**Purpose**: User authentication and authorization

**Components**:
- **Domain**: User entity
- **Application**: UserRepository interface, RegisterUseCase, LoginUseCase, ValidateTokenUseCase, RefreshTokenUseCase
- **Adapters**: HTTP handlers for registration, login, token refresh
- **Infrastructure**: PostgreSQL UserRepository, JWT token manager

**Endpoints**:
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token

## Leaderboard Module

**Purpose**: Score update and real-time leaderboard queries via Server-Sent Events (SSE)

**Components**:
- **Domain**: 
  - `LeaderboardEntry` (user_id, username, score, rank)
  - `Leaderboard` (entries, total)
  - Domain constants (Redis keys, topics)
- **Application**: 
  - `LeaderboardBackupRepository` interface
  - `LeaderboardRepository` interface
  - `UserRepository` interface (module-owned, not from auth module)
  - `BroadcastService` interface - For real-time leaderboard broadcasting (publishes notifications)
  - `ScoreUseCase` - Updates score in PostgreSQL and Redis, publishes notifications via broadcast service
  - `LeaderboardUseCase` - Handles leaderboard queries, enrichment, and broadcasting:
    - `SyncFromPostgres()` - Lazy loading from PostgreSQL to Redis
    - `GetFullLeaderboard()` - Fetches leaderboard and enriches with usernames
    - `StartBroadcasting()` - Listens to score updates and broadcasts leaderboard
    - `SubscribeToLeaderboardUpdates()` - Provides subscription channel for SSE handlers
- **Adapters**: 
  - `LeaderboardHandler` - HTTP handler for score update and leaderboard retrieval (SSE)
    - `RegisterPublicRoutes()` - Registers public routes (SSE stream)
    - `RegisterProtectedRoutes()` - Registers protected routes (score submission)
- **Infrastructure**: 
  - PostgreSQL `LeaderboardBackupRepository` - Stores scores, `GetLeaderboard()` with username JOIN
  - Redis `LeaderboardRepository` - Sorted sets for real-time queries
  - PostgreSQL `UserRepository` - Batch username fetching for leaderboard module
  - Redis `BroadcastService` (`RedisBroadcastService`) - Implements pub/sub for real-time leaderboard broadcasting

**Endpoints**:
- `PUT /api/v1/leaderboard/score` - Update score (authenticated)
- `GET /api/v1/leaderboard/stream?limit=50` - SSE stream for leaderboard (real-time updates)

**Key Features**:
- **Score Updates**: UPSERT pattern (creates if not exists, updates if exists)
- **PostgreSQL Backup**: Stores score per user as backup/recovery for Redis
- **Lazy Loading**: Automatic sync from PostgreSQL to Redis when Redis is empty
- **Real-Time Updates**: Server-Sent Events (SSE) with Redis pub/sub notifications
- **Username Enrichment**: 
  - PostgreSQL `GetLeaderboard()` includes usernames via JOIN
  - Redis entries enriched with usernames via batch fetch in application layer
- **Broadcast Service**: 
  - Publishes score update notifications (triggered by use cases, not repositories)
  - Centralized service processes score updates once
  - Uses distributed lock for multi-instance deployments
  - Publishes full leaderboard to all connected clients
  - Repositories only update data, broadcast service handles all pub/sub concerns
- **Module Independence**: Owns its `UserRepository` interface (no dependency on auth module)

For detailed flows, see [Architecture](./architecture.md) and [Application Features](./application.md).
