# API Documentation

## Swagger UI

Interactive API documentation is available via Swagger UI:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Swagger JSON**: http://localhost:8080/swagger/doc.json
- **Swagger YAML**: http://localhost:8080/swagger/swagger.yaml

The Swagger UI provides an interactive interface to test all API endpoints directly from your browser.

**Note:** Swagger documentation is generated from code annotations. To regenerate:
```bash
make swagger
```

Generated files are located in the `api/` directory.

## Authentication

All authenticated endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <access_token>
```

### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    "token": {
      "access_token": "jwt_token",
      "refresh_token": "jwt_token",
      "expires_in": 900
    }
  },
  "message": "User registered successfully"
}
```

### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "john_doe",
  "password": "password123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com"
    },
    "token": {
      "access_token": "jwt_token",
      "refresh_token": "jwt_token",
      "expires_in": 900
    }
  },
  "message": "Login successful"
}
```

### Refresh Token
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "jwt_refresh_token"
}
```

## Scores

### Submit Score
```http
POST /api/v1/scores
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "game_id": "game1",
  "score": 1000,
  "metadata": {
    "level": 5,
    "time": 120
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "score_uuid",
    "user_id": "user_uuid",
    "game_id": "game1",
    "score": 1000,
    "submitted_at": "2024-01-01T00:00:00Z",
    "metadata": {
      "level": 5,
      "time": 120
    }
  },
  "message": "Score submitted successfully"
}
```

### Get User Scores
```http
GET /api/v1/scores/me?game_id=game1&limit=10&offset=0
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `game_id` (optional): Filter by game ID
- `limit` (optional, default: 10, max: 100): Number of results per page
- `offset` (optional, default: 0): Number of results to skip

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "score_uuid",
      "user_id": "user_uuid",
      "game_id": "game1",
      "score": 1000,
      "submitted_at": "2024-01-01T00:00:00Z"
    }
  ],
  "message": "Scores retrieved successfully"
}
```

## Leaderboard

### Get Global Leaderboard
```http
GET /api/v1/leaderboard/global?limit=10&offset=0
```

**Query Parameters:**
- `limit` (optional, default: 10, max: 100): Number of top players to return
- `offset` (optional, default: 0): Number of results to skip

**Response:**
```json
{
  "success": true,
  "data": {
    "game_id": "global",
    "entries": [
      {
        "user_id": "user1",
        "score": 5000,
        "rank": 1
      },
      {
        "user_id": "user2",
        "score": 4500,
        "rank": 2
      }
    ],
    "total": 150
  },
  "message": "Leaderboard retrieved successfully"
}
```

### Get Game Leaderboard
```http
GET /api/v1/leaderboard/game/game1?limit=10&offset=0
```

**Query Parameters:**
- `limit` (optional, default: 10, max: 100): Number of top players to return
- `offset` (optional, default: 0): Number of results to skip

**Response:**
```json
{
  "success": true,
  "data": {
    "game_id": "game1",
    "entries": [
      {
        "user_id": "user1",
        "score": 3000,
        "rank": 1
      }
    ],
    "total": 50
  },
  "message": "Leaderboard retrieved successfully"
}
```

### Get User Rank
```http
GET /api/v1/leaderboard/rank/{user_id}?game_id=game1
```

**Path Parameters:**
- `user_id`: User ID to get rank for

**Query Parameters:**
- `game_id` (optional, default: "global"): Game ID (empty or "global" for global leaderboard)

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": "user123",
    "score": 2500,
    "rank": 5
  },
  "message": "User rank retrieved successfully"
}
```

## Reports

### Get Top Players Report
```http
GET /api/v1/reports/top-players?game_id=game1&start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z&limit=10&offset=0
```

**Query Parameters:**
- `game_id` (optional, default: "global"): Game ID filter
- `start_date` (optional): Start date in RFC3339 format
- `end_date` (optional): End date in RFC3339 format
- `limit` (optional, default: 10, max: 100): Number of top players to return
- `offset` (optional, default: 0): Number of results to skip

**Note:** When both `start_date` and `end_date` are provided, the system uses PostgreSQL for historical data. Otherwise, it uses Redis for current leaderboard data.

**Response:**
```json
{
  "success": true,
  "data": {
    "game_id": "game1",
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z",
    "players": [
      {
        "user_id": "user1",
        "score": 5000,
        "rank": 1,
        "game_id": "game1",
        "last_updated": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 100
  },
  "message": "Report generated successfully"
}
```

## WebSocket

### Real-Time Leaderboard Updates

Connect to the WebSocket endpoint for real-time leaderboard updates:

```
ws://localhost:8080/ws/leaderboard?game_id=game1
```

**Query Parameters:**
- `game_id` (optional): Game ID filter (currently broadcasts global leaderboard to all clients)

**Behavior:**
- Server automatically broadcasts leaderboard updates every 5 seconds
- Initial leaderboard is sent immediately upon connection
- Connection remains open for continuous updates
- Server sends ping messages every 54 seconds to keep connection alive

**Message Format:**
```json
{
  "game_id": "global",
  "entries": [
    {
      "user_id": "user1",
      "score": 5000,
      "rank": 1
    }
  ],
  "total": 150
}
```

## Error Responses

All error responses follow this format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Error description"
  }
}
```

**Common Error Codes:**
- `VALIDATION_ERROR`: Request validation failed
- `UNAUTHORIZED`: Authentication required or invalid token
- `NOT_FOUND`: Resource not found
- `CONFLICT`: Resource already exists (e.g., duplicate username)
- `INTERNAL_ERROR`: Internal server error

## Pagination

All list endpoints support pagination with the following parameters:

- `limit`: Number of items per page (default: 10, min: 1, max: 100)
- `offset`: Number of items to skip (default: 0, min: 0)

Example:
```
GET /api/v1/leaderboard/global?limit=20&offset=40
```
This returns items 41-60 (page 3 with 20 items per page).
