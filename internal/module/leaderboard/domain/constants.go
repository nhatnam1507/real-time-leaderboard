// Package domain provides domain entities and constants for the leaderboard module.
package domain

const (
	// Redis pub/sub topics
	RedisScoreUpdateTopic  = "leaderboard:score:updates"  // Published when score updates
	RedisViewerUpdateTopic = "leaderboard:viewer:updates" // Published with full leaderboard data

	// Redis keys
	RedisLeaderboardKey  = "leaderboard:global"         // Redis sorted set key
	RedisBroadcastLockKey = "leaderboard:broadcast:lock" // Distributed lock key
)
