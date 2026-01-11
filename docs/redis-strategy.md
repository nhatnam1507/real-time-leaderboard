# Redis Strategy

The leaderboard system uses Redis for real-time leaderboard queries and notifications. PostgreSQL stores the score per user as a backup/recovery mechanism.

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

### Data Recovery

If Redis data is lost, the system can rebuild the leaderboard from PostgreSQL:
- PostgreSQL stores one record per user with their current score (UPSERT pattern)
- All users' scores can be queried and restored to Redis
- This ensures data durability while maintaining Redis performance for real-time queries

For complete system flow, see [Architecture](./architecture.md).

### Benefits

- **Real-time**: Updates pushed immediately when scores change (no polling)
- **Efficient**: Only fetches data when there's an actual update
- **Decoupled**: Score and leaderboard modules communicate via Redis (shared infrastructure)
- **Scalable**: Works across multiple server instances
