package services

import (
	"fmt"

	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/models"
	"gorm.io/gorm"
)

type PlayerService struct {
	db *gorm.DB
}

func NewPlayerService(db *gorm.DB) *PlayerService {
	return &PlayerService{db: db}
}

func (s *PlayerService) GetByID(id string) (*models.Player, error) {
	var player models.Player
	if err := s.db.Where("id = ?", id).First(&player).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.NotFound("player")
		}
		return nil, fmt.Errorf("fetching player: %w", err)
	}
	return &player, nil
}

func (s *PlayerService) GetStats(playerID string) (*models.PlayerStats, error) {
	var stats models.PlayerStats
	if err := s.db.Where("player_id = ?", playerID).First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperr.NotFound("player stats")
		}
		return nil, fmt.Errorf("fetching stats: %w", err)
	}
	return &stats, nil
}

func (s *PlayerService) UpdateProfile(playerID string, req dto.UpdatePlayerRequest) (*models.Player, error) {
	updates := map[string]interface{}{}
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.AvatarID != nil {
		updates["avatar_id"] = *req.AvatarID
	}

	if len(updates) == 0 {
		return s.GetByID(playerID)
	}

	if err := s.db.Model(&models.Player{}).Where("id = ?", playerID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("updating player: %w", err)
	}

	return s.GetByID(playerID)
}
