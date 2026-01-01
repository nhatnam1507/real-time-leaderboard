# API Documentation

## Authentication

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

### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "john_doe",
  "password": "password123"
}
```

Response:
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
  }
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

### Get User Scores
```http
GET /api/v1/scores/me?game_id=game1&limit=10&offset=0
Authorization: Bearer <access_token>
```

## Leaderboard

### Get Global Leaderboard
```http
GET /api/v1/leaderboard/global?limit=10
```

### Get Game Leaderboard
```http
GET /api/v1/leaderboard/game/game1?limit=10
```

### Get User Rank
```http
GET /api/v1/leaderboard/rank/{user_id}?game_id=game1
```

## Reports

### Get Top Players Report
```http
GET /api/v1/reports/top-players?game_id=game1&start_date=2024-01-01T00:00:00Z&end_date=2024-01-31T23:59:59Z&limit=10
```

## WebSocket

Connect to `ws://localhost:8080/ws/leaderboard?game_id=game1` for real-time leaderboard updates.

