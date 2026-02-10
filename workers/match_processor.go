package workers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/wombocombo/api-server/dto"
	"github.com/wombocombo/api-server/models"
	"gorm.io/gorm"
)

// StartMatchProcessor polls Redis for match results and persists them to Postgres.
// This is the fallback path — the preferred path is POST /internal/match/end.
// Match results written by the game server to Redis keys like "match:<id>:result"
// are picked up here, processed, and deleted.
func StartMatchProcessor(ctx context.Context, redisClient *redis.Client, db *gorm.DB) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Info().Msg("match_processor started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("match_processor stopped")
			return
		case <-ticker.C:
			processMatchResults(ctx, redisClient, db)
		}
	}
}

func processMatchResults(ctx context.Context, redisClient *redis.Client, db *gorm.DB) {
	var cursor uint64
	for {
		keys, newCursor, err := redisClient.Scan(ctx, cursor, "match:*:result", 100).Result()
		if err != nil {
			log.Error().Err(err).Msg("scanning match results")
			return
		}

		for _, key := range keys {
			processOneResult(ctx, redisClient, db, key)
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
}

func processOneResult(ctx context.Context, redisClient *redis.Client, db *gorm.DB, key string) {
	resultJSON, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return
	}

	// Try parsing as internal format first (with coop_revives)
	var internal dto.MatchResultInternal
	if err := json.Unmarshal([]byte(resultJSON), &internal); err != nil {
		log.Error().Err(err).Str("key", key).Msg("parsing match result")
		return
	}

	// If no internal players, try standard format
	if len(internal.Players) == 0 {
		var standard dto.MatchResult
		if err := json.Unmarshal([]byte(resultJSON), &standard); err != nil {
			log.Error().Err(err).Str("key", key).Msg("parsing match result (standard)")
			return
		}
		internal = dto.MatchResultInternal{
			MatchID:         standard.MatchID,
			RoomID:          standard.RoomID,
			StartedAt:       standard.StartedAt,
			EndedAt:         standard.EndedAt,
			RoundsCompleted: standard.RoundsCompleted,
			MapID:           standard.MapID,
		}
		for _, p := range standard.Players {
			internal.Players = append(internal.Players, dto.MatchPlayerResultInternal{
				PlayerID:       p.PlayerID,
				Kills:          p.Kills,
				Deaths:         p.Deaths,
				Score:          p.Score,
				RoundsSurvived: p.RoundsSurvived,
			})
		}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		match := models.Match{
			ID:              internal.MatchID,
			RoomID:          internal.RoomID,
			StartedAt:       internal.StartedAt,
			EndedAt:         internal.EndedAt,
			RoundsCompleted: internal.RoundsCompleted,
			MapID:           internal.MapID,
		}
		if err := tx.Create(&match).Error; err != nil {
			return err
		}

		matchDurationSecs := int(internal.EndedAt.Sub(internal.StartedAt).Seconds())
		if matchDurationSecs < 0 {
			matchDurationSecs = 0
		}

		for _, p := range internal.Players {
			mp := models.MatchPlayer{
				MatchID:        internal.MatchID,
				PlayerID:       p.PlayerID,
				Kills:          p.Kills,
				Deaths:         p.Deaths,
				Score:          p.Score,
				RoundsSurvived: p.RoundsSurvived,
			}
			if err := tx.Create(&mp).Error; err != nil {
				return err
			}

			// Currency reward: 10/round + 5/kill + 2/revive + 15/round survived
			currency := internal.RoundsCompleted*10 + p.Kills*5 + p.CoopRevives*2 + p.RoundsSurvived*15
			if currency < 0 {
				currency = 0
			}

			if err := tx.Exec(`
				UPDATE player_stats SET
					rounds_played = rounds_played + ?,
					rounds_survived = rounds_survived + ?,
					total_kills = total_kills + ?,
					total_deaths = total_deaths + ?,
					best_round = GREATEST(best_round, ?),
					total_playtime_secs = total_playtime_secs + ?,
					coop_revives = coop_revives + ?,
					currency = currency + ?,
					updated_at = NOW()
				WHERE player_id = ?
			`, internal.RoundsCompleted, p.RoundsSurvived, p.Kills, p.Deaths,
				internal.RoundsCompleted, matchDurationSecs, p.CoopRevives,
				currency, p.PlayerID).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Error().Err(err).Str("match", internal.MatchID).Msg("persisting match")
		return
	}

	// Clean up Redis tracking
	pipe := redisClient.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, "active_match:"+internal.MatchID)
	for _, p := range internal.Players {
		pipe.Del(ctx, "player_match:"+p.PlayerID)
	}
	pipe.Exec(ctx)

	log.Info().
		Str("match", internal.MatchID).
		Int("players", len(internal.Players)).
		Msg("match persisted via worker")
}
