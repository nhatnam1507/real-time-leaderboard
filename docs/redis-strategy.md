# Redis Strategy

The leaderboard system uses Redis for real-time leaderboard queries and notifications. PostgreSQL stores the score per user as a backup/recovery mechanism.

## Sorted Sets for Leaderboard Storage

Redis sorted sets provide efficient leaderboard ranking:

- **Key**: `leaderboard:global` (single global leaderboard)
- **Score**: User's total score (as Redis score)
- **Member**: User ID

### Commands Used

- `ZADD` - Update/add user score
- `ZREVRANGE` - Get top N players (with scores)
- `ZCARD` - Get total players

### Performance Considerations

- Sorted sets provide O(log(N)) complexity for insertions and O(log(N)+M) for range queries
- All leaderboard operations are atomic
- Redis handles concurrent updates efficiently
- No need for application-level locking

## Pub/Sub for Real-Time Notifications

Redis pub/sub enables real-time leaderboard updates:

- **Score Update Topic**: `leaderboard:score:updates`
  - Published when scores are updated
  - Broadcast service subscribes to this topic

- **Viewer Update Topic**: `leaderboard:viewer:updates`
  - Published by broadcast service with full leaderboard JSON
  - SSE clients subscribe to this topic

### Broadcast Service

The `LeaderboardBroadcast` service handles real-time updates:

1. **Subscribes** to `leaderboard:score:updates` topic
2. **Acquires distributed lock** (`leaderboard:broadcast:lock`) to prevent duplicate processing in multi-instance deployments
3. **Fetches full leaderboard** from Redis
4. **Enriches entries** with usernames via batch fetch from PostgreSQL
5. **Publishes** full leaderboard to `leaderboard:viewer:updates` topic
6. **SSE clients** receive updates automatically

**Benefits**:
- Single processing: Only one instance processes each score update
- Efficient: Fetches leaderboard once, broadcasts to all clients
- Scalable: Works across multiple server instances

## Data Recovery and Lazy Loading

The system implements **lazy loading** to automatically sync PostgreSQL data to Redis:

- **Automatic Sync**: On first leaderboard request, if Redis is empty, the system automatically syncs all scores from PostgreSQL to Redis
- **Non-blocking**: Sync happens on-demand, not on startup, ensuring fast application startup
- **Resilient**: Handles Redis restarts gracefully - data is automatically restored on next request
- **Efficient**: Only syncs when needed (Redis is empty), avoiding unnecessary work

**Manual Recovery**:
If Redis data is lost, the system can rebuild the leaderboard from PostgreSQL:
- PostgreSQL stores one record per user with their current score (UPSERT pattern)
- `GetLeaderboard()` retrieves all scores with usernames via JOIN
- All users' scores can be restored to Redis via `ZADD` operations

## Username Enrichment

Leaderboard entries include usernames:

- **PostgreSQL**: `GetLeaderboard()` joins with `users` table to include usernames directly
- **Redis**: Entries fetched from Redis are enriched with usernames via batch fetch:
  - Application layer extracts user IDs from entries
  - Batch fetches usernames using `UserRepository.GetByIDs()`
  - Enriches entries before returning to clients

This design ensures:
- PostgreSQL queries are efficient (single JOIN query)
- Redis queries remain fast (no JOIN needed)
- Username enrichment happens in application layer (clean separation)

For complete system flow, see [Architecture](./architecture.md).
