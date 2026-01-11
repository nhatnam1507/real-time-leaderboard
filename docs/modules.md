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

**Purpose**: Score submission and real-time leaderboard queries via Server-Sent Events (SSE)

**Components**:
- **Domain**: Score entity, LeaderboardEntry entity, ScoreRepository interface, LeaderboardRepository interface
- **Application**: SubmitScoreUseCase, GetLeaderboardUseCase
- **Adapters**: HTTP handlers for score submission and leaderboard retrieval (SSE)
- **Infrastructure**: PostgreSQL ScoreRepository, Redis LeaderboardRepository (sorted sets)

**Endpoints**:
- `POST /api/v1/score` - Submit score (authenticated)
- `GET /api/v1/leaderboard` - SSE stream for leaderboard (real-time updates)

**Real-Time Updates**:
- When a score is submitted, it's stored in PostgreSQL and updated in Redis sorted sets for leaderboard ranking
- The module publishes notifications to Redis pub/sub channel (`leaderboard:updates`) to notify leaderboard clients
- SSE handlers subscribe to Redis pub/sub channel and fetch fresh leaderboard data when updates occur
- No polling - updates are pushed immediately when scores change

