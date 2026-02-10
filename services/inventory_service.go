package services

import (
	"fmt"
	"time"

	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/models"
	"gorm.io/gorm"
)

// ItemCatalog defines available items and their costs.
var ItemCatalog = map[string]int{
	"skin_zombie_01":  100,
	"skin_zombie_02":  200,
	"skin_zombie_03":  500,
	"skin_survivor_01": 100,
	"skin_survivor_02": 200,
	"skin_survivor_03": 500,
	"emote_dance":      50,
	"emote_taunt":      50,
	"emote_wave":       25,
	"trail_fire":       300,
	"trail_ice":        300,
	"trail_neon":       400,
}

type InventoryService struct {
	db *gorm.DB
}

func NewInventoryService(db *gorm.DB) *InventoryService {
	return &InventoryService{db: db}
}

func (s *InventoryService) ListItems(playerID string) ([]models.InventoryItem, error) {
	var items []models.InventoryItem
	err := s.db.Where("player_id = ?", playerID).Order("unlocked_at DESC").Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("fetching inventory: %w", err)
	}
	return items, nil
}

func (s *InventoryService) UnlockItem(playerID, itemType string) (*models.InventoryItem, error) {
	// Check item exists in catalog
	cost, exists := ItemCatalog[itemType]
	if !exists {
		return nil, apperr.BadRequest("unknown item type")
	}

	// Check not already owned
	var count int64
	s.db.Model(&models.InventoryItem{}).Where("player_id = ? AND item_type = ?", playerID, itemType).Count(&count)
	if count > 0 {
		return nil, apperr.Conflict("item already unlocked")
	}

	// Check currency
	var stats models.PlayerStats
	if err := s.db.Where("player_id = ?", playerID).First(&stats).Error; err != nil {
		return nil, apperr.NotFound("player stats")
	}
	if stats.Currency < cost {
		return nil, apperr.BadRequest(fmt.Sprintf("insufficient currency: need %d, have %d", cost, stats.Currency))
	}

	var item models.InventoryItem
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Deduct currency
		if err := tx.Model(&models.PlayerStats{}).
			Where("player_id = ? AND currency >= ?", playerID, cost).
			Update("currency", gorm.Expr("currency - ?", cost)).Error; err != nil {
			return err
		}

		// Create item
		item = models.InventoryItem{
			PlayerID:   playerID,
			ItemType:   itemType,
			UnlockedAt: time.Now(),
		}
		return tx.Create(&item).Error
	})
	if err != nil {
		return nil, fmt.Errorf("unlocking item: %w", err)
	}

	return &item, nil
}

func (s *InventoryService) GetCatalog() map[string]int {
	return ItemCatalog
}
