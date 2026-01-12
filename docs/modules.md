# Modules

The system is organized into self-contained modules, each following Clean Architecture principles.

## Auth Module

**Purpose**: User authentication and authorization

**Components**:
- **Domain**: User entity, UserRepository interface
- **Application**: RegisterUseCase, LoginUseCase, ValidateTokenUseCase, RefreshTokenUseCase
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
  - `LeaderboardBackupRepository` interface
  - `LeaderboardRepository` interface
  - `UserRepository` interface (module-owned, not from auth module)
- **Application**: 
  - `SubmitScoreUseCase` - Updates score in PostgreSQL and Redis
  - `SyncFromPostgresUseCase` - Lazy loading from PostgreSQL to Redis
  - `GetFullLeaderboardUseCase` - Fetches leaderboard and enriches with usernames
- **Adapters**: 
  - `LeaderboardHandler` - Single HTTP handler for score update and leaderboard retrieval (SSE)
  - `LeaderboardBroadcast` - Broadcast service for real-time updates
- **Infrastructure**: 
  - PostgreSQL `LeaderboardBackupRepository` - Stores scores, `GetLeaderboard()` with username JOIN
  - Redis `LeaderboardRepository` - Sorted sets for real-time queries
  - PostgreSQL `UserRepository` - Batch username fetching for leaderboard module

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
  - Centralized service processes score updates once
  - Uses distributed lock for multi-instance deployments
  - Publishes full leaderboard to all connected clients
- **Module Independence**: Owns its `UserRepository` interface (no dependency on auth module)

For detailed flows, see [Architecture](./architecture.md) and [Redis Strategy](./redis-strategy.md).
