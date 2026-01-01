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

## Leaderboard Module

**Purpose**: Leaderboard queries and real-time updates

**Components**:
- **Domain**: LeaderboardEntry entity, LeaderboardRepository interface
- **Application**: GetGlobalLeaderboardUseCase, GetGameLeaderboardUseCase, GetUserRankUseCase
- **Adapters**: HTTP handlers + WebSocket handler
- **Infrastructure**: Redis LeaderboardRepository (sorted sets)

**Endpoints**:
- `GET /api/v1/leaderboard/global` - Get global leaderboard
- `GET /api/v1/leaderboard/game/:game_id` - Get game-specific leaderboard
- `GET /api/v1/leaderboard/rank/:user_id` - Get user's rank
- `WS /ws/leaderboard` - WebSocket for real-time updates

## Report Module

**Purpose**: Top players reports and analytics

**Components**:
- **Domain**: TopPlayer entity, ReportRepository interface
- **Application**: GetTopPlayersReportUseCase
- **Adapters**: HTTP handler for report generation
- **Infrastructure**: Composite repository (Redis + PostgreSQL)

**Endpoints**:
- `GET /api/v1/reports/top-players` - Top players report (with optional date range)

