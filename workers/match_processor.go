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
			resultJSON, err := redisClient.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var result dto.MatchResult
			if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
				log.Error().Err(err).Str("key", key).Msg("parsing match result")
				continue
			}

			err = db.Transaction(func(tx *gorm.DB) error {
				match := models.Match{
					ID:              result.MatchID,
					RoomID:          result.RoomID,
					StartedAt:       result.StartedAt,
					EndedAt:         result.EndedAt,
					RoundsCompleted: result.RoundsCompleted,
					MapID:           result.MapID,
				}
				if err := tx.Create(&match).Error; err != nil {
					return err
				}

				for _, p := range result.Players {
					mp := models.MatchPlayer{
						MatchID:        result.MatchID,
						PlayerID:       p.PlayerID,
						Kills:          p.Kills,
						Deaths:         p.Deaths,
						Score:          p.Score,
						RoundsSurvived: p.RoundsSurvived,
					}
					if err := tx.Create(&mp).Error; err != nil {
						return err
					}

					if err := tx.Exec(`
						UPDATE player_stats SET
							rounds_played = rounds_played + ?,
							rounds_survived = rounds_survived + ?,
							total_kills = total_kills + ?,
							total_deaths = total_deaths + ?,
							best_round = GREATEST(best_round, ?),
							updated_at = NOW()
						WHERE player_id = ?
					`, result.RoundsCompleted, p.RoundsSurvived, p.Kills, p.Deaths,
						result.RoundsCompleted, p.PlayerID).Error; err != nil {
						return err
					}
				}
				return nil
			})

			if err != nil {
				log.Error().Err(err).Str("match", result.MatchID).Msg("persisting match")
				continue
			}

			redisClient.Del(ctx, key)
			log.Info().Str("match", result.MatchID).Msg("match persisted successfully")
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
}
