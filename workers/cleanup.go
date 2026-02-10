package workers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// StartCleanup periodically removes expired rooms from the public rooms set.
func StartCleanup(ctx context.Context, redisClient *redis.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Info().Msg("cleanup worker started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("cleanup worker stopped")
			return
		case <-ticker.C:
			cleanExpiredRooms(ctx, redisClient)
		}
	}
}

func cleanExpiredRooms(ctx context.Context, redisClient *redis.Client) {
	ids, err := redisClient.SMembers(ctx, "rooms:public").Result()
	if err != nil {
		return
	}

	for _, id := range ids {
		exists, _ := redisClient.Exists(ctx, "room:"+id).Result()
		if exists == 0 {
			redisClient.SRem(ctx, "rooms:public", id)
			log.Debug().Str("room", id).Msg("cleaned expired room from public set")
		}
	}
}
