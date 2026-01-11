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
- **Domain**: Score entity, LeaderboardEntry entity, LeaderboardBackupRepository interface, LeaderboardRepository interface
- **Application**: SubmitScoreUseCase, GetLeaderboardUseCase
- **Adapters**: HTTP handlers for score update and leaderboard retrieval (SSE)
- **Infrastructure**: PostgreSQL LeaderboardBackupRepository (stores score per user), Redis LeaderboardRepository (sorted sets for real-time queries)

**Endpoints**:
- `PUT /api/v1/score` - Update score (authenticated)
- `GET /api/v1/leaderboard` - SSE stream for leaderboard (real-time updates)

**Overview**:
- Score updates use UPSERT pattern (creates if not exists, updates if exists)
- PostgreSQL stores score per user as backup/recovery for Redis
- Redis stores leaderboard in sorted sets for real-time queries
- Real-time updates via Server-Sent Events (SSE) with Redis pub/sub notifications

For detailed flows and data storage strategy, see [Architecture](./architecture.md) and [Redis Strategy](./redis-strategy.md).
