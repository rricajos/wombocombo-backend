package services

import (
	"context"
	"fmt"

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

	// Kill their session
	s.redis.Del(ctx, "session:"+playerID)
	s.redis.Del(ctx, "player_cache:"+playerID)

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

	s.db.Model(&models.Player{}).Count(&stats.TotalPlayers)
	s.db.Model(&models.Player{}).Where("is_banned = ?", true).Count(&stats.BannedPlayers)
	s.db.Model(&models.Match{}).Count(&stats.TotalMatches)

	// Count active sessions from Redis
	var cursor uint64
	var sessionCount int64
	for {
		keys, newCursor, err := s.redis.Scan(ctx, cursor, "session:*", 1000).Result()
		if err != nil {
			break
		}
		sessionCount += int64(len(keys))
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
	stats.ActiveSessions = sessionCount

	// Count active rooms
	roomCount, _ := s.redis.SCard(ctx, "rooms:public").Result()
	stats.ActiveRooms = roomCount

	return &stats, nil
}
