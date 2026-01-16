// Package domain provides domain entities and constants for the leaderboard module.
package domain

const (
	// RedisViewerUpdateTopic is the Redis pub/sub topic published with leaderboard entry delta updates for viewers.
	RedisViewerUpdateTopic = "leaderboard:viewer:updates"

	// RedisLeaderboardKey is the Redis sorted set key for the global leaderboard.
	RedisLeaderboardKey = "leaderboard:global"

	// MaxBroadcastRank is the maximum rank for which entry updates are broadcasted.
	// Entries ranked higher than this will not trigger broadcasts to reduce unnecessary network traffic.
	// This threshold should be higher than any client's typical limit (e.g., 1000 covers clients showing top 5/10/50/100).
	MaxBroadcastRank = 1000
)
