# Redis Sorted Sets Strategy

The leaderboard uses Redis sorted sets for efficient ranking:

- **Key Pattern**: `leaderboard:{game_id}` for per-game leaderboards
- **Key Pattern**: `leaderboard:global` for global leaderboard
- **Score**: User's total score (as Redis score)
- **Member**: User ID

## Commands Used

- `ZADD` - Update/add user score
- `ZREVRANGE` - Get top N players
- `ZREVRANK` - Get user's rank
- `ZSCORE` - Get user's score
- `ZCARD` - Get total players

## Performance Considerations

- Sorted sets provide O(log(N)) complexity for insertions and O(log(N)+M) for range queries
- All leaderboard operations are atomic
- Redis handles concurrent updates efficiently
- No need for application-level locking

