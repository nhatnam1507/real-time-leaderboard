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

## Score Module

**Purpose**: Score submission and management

**Components**:
- **Domain**: Score entity, ScoreRepository interface, LeaderboardRepository interface
- **Application**: SubmitScoreUseCase, GetUserScoresUseCase
- **Adapters**: HTTP handlers for score submission and retrieval
- **Infrastructure**: PostgreSQL ScoreRepository, Redis LeaderboardRepository

**Endpoints**:
- `POST /api/v1/scores` - Submit score (authenticated)
- `GET /api/v1/scores/me` - Get user's score history (authenticated)

**Real-Time Notifications**:
- When a score is submitted, it's stored in Redis sorted sets for leaderboard ranking
- The module publishes notifications to Redis pub/sub channels to notify leaderboard clients
- Channels: `leaderboard:updates:{gameID}` for game-specific updates, `leaderboard:updates:global` for global updates
- No direct dependencies on leaderboard module - uses Redis (shared infrastructure) for communication

## Leaderboard Module

**Purpose**: Real-time leaderboard queries via Server-Sent Events (SSE)

**Components**:
- **Domain**: LeaderboardEntry entity, LeaderboardRepository interface
- **Application**: GetGlobalLeaderboardUseCase, GetGameLeaderboardUseCase, GetUserRankUseCase
- **Adapters**: HTTP handlers with SSE support
- **Infrastructure**: Redis LeaderboardRepository (sorted sets)

**Endpoints**:
- `GET /api/v1/leaderboard/global` - SSE stream for global leaderboard (real-time updates)
- `GET /api/v1/leaderboard/game/:game_id` - SSE stream for game-specific leaderboard (real-time updates)
- `GET /api/v1/leaderboard/rank/:user_id` - Get user's rank (regular REST endpoint)

**Real-Time Updates**:
- When scores are submitted, the score module publishes notifications to Redis pub/sub channels
- SSE handlers subscribe to Redis pub/sub channels (`leaderboard:updates:global` and `leaderboard:updates:{gameID}`)
- On receiving notifications, handlers fetch fresh leaderboard data from Redis and send updates to connected clients
- No polling - updates are pushed immediately when scores change

## Report Module

**Purpose**: Top players reports and analytics

**Components**:
- **Domain**: TopPlayer entity, ReportRepository interface
- **Application**: GetTopPlayersReportUseCase
- **Adapters**: HTTP handler for report generation
- **Infrastructure**: Composite repository (Redis + PostgreSQL)

**Endpoints**:
- `GET /api/v1/reports/top-players` - Top players report (with optional date range)

