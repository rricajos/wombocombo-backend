package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/wombocombo/api-server/config"
)

func ConnectRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Str("addr", cfg.RedisAddr).Msg("failed to connect to Redis")
	}

	log.Info().Str("addr", cfg.RedisAddr).Msg("connected to Redis")
	return client
}

// PublishJWTSecret writes the JWT secret to Redis so the game server can validate tokens.
func PublishJWTSecret(ctx context.Context, client *redis.Client, secret string) {
	if err := client.Set(ctx, "jwt:secret", secret, 0).Err(); err != nil {
		log.Error().Err(err).Msg("failed to publish JWT secret to Redis")
		return
	}
	log.Info().Msg("JWT secret published to Redis")
}
