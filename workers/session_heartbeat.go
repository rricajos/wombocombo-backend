package workers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// StartSessionHeartbeat periodically counts online players and cleans stale heartbeats.
// It also publishes online player count to Redis for monitoring.
func StartSessionHeartbeat(ctx context.Context, redisClient *redis.Client) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	log.Info().Msg("session_heartbeat worker started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("session_heartbeat worker stopped")
			return
		case <-ticker.C:
			processHeartbeats(ctx, redisClient)
		}
	}
}

func processHeartbeats(ctx context.Context, redisClient *redis.Client) {
	// Count active sessions
	var sessionCount int64
	var cursor uint64
	for {
		keys, newCursor, err := redisClient.Scan(ctx, cursor, "session:*", 1000).Result()
		if err != nil {
			break
		}
		sessionCount += int64(len(keys))
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	// Count players with recent heartbeats (connected to game server)
	var heartbeatCount int64
	cursor = 0
	for {
		keys, newCursor, err := redisClient.Scan(ctx, cursor, "heartbeat:*", 1000).Result()
		if err != nil {
			break
		}
		heartbeatCount += int64(len(keys))
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	// Count active matches
	var matchCount int64
	cursor = 0
	for {
		keys, newCursor, err := redisClient.Scan(ctx, cursor, "active_match:*", 1000).Result()
		if err != nil {
			break
		}
		matchCount += int64(len(keys))
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	// Publish stats to Redis for monitoring dashboards
	pipe := redisClient.Pipeline()
	pipe.HSet(ctx, "server:stats:live",
		"sessions", sessionCount,
		"in_game", heartbeatCount,
		"active_matches", matchCount,
		"updated_at", time.Now().Unix(),
	)
	pipe.Expire(ctx, "server:stats:live", 5*time.Minute)
	pipe.Exec(ctx)

	log.Debug().
		Int64("sessions", sessionCount).
		Int64("in_game", heartbeatCount).
		Int64("active_matches", matchCount).
		Msg("heartbeat stats updated")
}
