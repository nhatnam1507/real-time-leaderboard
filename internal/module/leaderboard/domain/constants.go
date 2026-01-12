// Package domain provides domain entities and constants for the leaderboard module.
package domain

const (
	// RedisScoreUpdateTopic is the Redis pub/sub topic published when score updates occur.
	RedisScoreUpdateTopic = "leaderboard:score:updates"
	// RedisViewerUpdateTopic is the Redis pub/sub topic published with full leaderboard data for viewers.
	RedisViewerUpdateTopic = "leaderboard:viewer:updates"

	// RedisLeaderboardKey is the Redis sorted set key for the global leaderboard.
	RedisLeaderboardKey = "leaderboard:global"
	// RedisBroadcastLockKey is the Redis key used for distributed locking in broadcast service.
	RedisBroadcastLockKey = "leaderboard:broadcast:lock"
)
