package services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/models"
	"gorm.io/gorm"
)

type AdminService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewAdminService(db *gorm.DB, redis *redis.Client) *AdminService {
	return &AdminService{db: db, redis: redis}
}

func (s *AdminService) BanPlayer(ctx context.Context, playerID string) error {
	result := s.db.Model(&models.Player{}).Where("id = ?", playerID).Update("is_banned", true)
	if result.RowsAffected == 0 {
		return apperr.NotFound("player")
	}
	if result.Error != nil {
		return fmt.Errorf("banning player: %w", result.Error)
	}

	// Kill their session and cache
	s.redis.Del(ctx, "session:"+playerID)
	s.redis.Del(ctx, "player_cache:"+playerID)
	s.redis.Del(ctx, "heartbeat:"+playerID)

	return nil
}

func (s *AdminService) UnbanPlayer(playerID string) error {
	result := s.db.Model(&models.Player{}).Where("id = ?", playerID).Update("is_banned", false)
	if result.RowsAffected == 0 {
		return apperr.NotFound("player")
	}
	if result.Error != nil {
		return fmt.Errorf("unbanning player: %w", result.Error)
	}
	return nil
}

func (s *AdminService) GetServerStats(ctx context.Context) (*dto.ServerStatsResponse, error) {
	var stats dto.ServerStatsResponse

	// DB stats
	s.db.Model(&models.Player{}).Count(&stats.TotalPlayers)
	s.db.Model(&models.Player{}).Where("is_banned = ?", true).Count(&stats.BannedPlayers)
	s.db.Model(&models.Match{}).Count(&stats.TotalMatches)

	// Try live stats from heartbeat worker first (fast path)
	live, err := s.redis.HGetAll(ctx, "server:stats:live").Result()
	if err == nil && len(live) > 0 {
		if v, ok := live["sessions"]; ok {
			stats.ActiveSessions, _ = strconv.ParseInt(v, 10, 64)
		}
		if v, ok := live["in_game"]; ok {
			stats.PlayersInGame, _ = strconv.ParseInt(v, 10, 64)
		}
		if v, ok := live["active_matches"]; ok {
			stats.ActiveMatches, _ = strconv.ParseInt(v, 10, 64)
		}
	} else {
		// Fallback: count directly (slower)
		var cursor uint64
		for {
			keys, newCursor, err := s.redis.Scan(ctx, cursor, "session:*", 1000).Result()
			if err != nil {
				break
			}
			stats.ActiveSessions += int64(len(keys))
			cursor = newCursor
			if cursor == 0 {
				break
			}
		}
	}

	// Count active rooms
	roomCount, _ := s.redis.SCard(ctx, "rooms:public").Result()
	stats.ActiveRooms = roomCount

	return &stats, nil
}
