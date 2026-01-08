# Redis Strategy

The leaderboard system uses Redis for both storage and real-time notifications.

## Sorted Sets for Leaderboard Storage

Redis sorted sets provide efficient leaderboard ranking:

- **Key Pattern**: `leaderboard:{game_id}` for per-game leaderboards
- **Key Pattern**: `leaderboard:global` for global leaderboard
- **Score**: User's total score (as Redis score)
- **Member**: User ID

### Commands Used

- `ZADD` - Update/add user score
- `ZREVRANGE` - Get top N players
- `ZREVRANK` - Get user's rank
- `ZSCORE` - Get user's score
- `ZCARD` - Get total players

### Performance Considerations

- Sorted sets provide O(log(N)) complexity for insertions and O(log(N)+M) for range queries
- All leaderboard operations are atomic
- Redis handles concurrent updates efficiently
- No need for application-level locking

## Pub/Sub for Real-Time Notifications

Redis pub/sub is used to notify SSE clients when leaderboard data changes:

- **Channel Pattern**: `leaderboard:updates:{game_id}` for game-specific updates
- **Channel Pattern**: `leaderboard:updates:global` for global leaderboard updates
- **Publisher**: Score module publishes notifications when scores are updated
- **Subscriber**: Leaderboard SSE handlers subscribe to channels and fetch fresh data on notification

### Flow

1. Score is submitted → stored in Redis sorted set
2. Score module publishes notification to Redis pub/sub channel
3. SSE handlers subscribed to the channel receive notification
4. Handlers fetch fresh leaderboard data from Redis sorted sets
5. Updated leaderboard sent to clients via Server-Sent Events (SSE)

### Benefits

- **Real-time**: Updates pushed immediately when scores change (no polling)
- **Efficient**: Only fetches data when there's an actual update
- **Decoupled**: Score and leaderboard modules communicate via Redis (shared infrastructure)
- **Scalable**: Works across multiple server instances
