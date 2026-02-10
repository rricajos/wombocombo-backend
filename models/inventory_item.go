package models

import "time"

type InventoryItem struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PlayerID   string    `gorm:"type:uuid;index" json:"player_id"`
	ItemType   string    `gorm:"size:64" json:"item_type"`
	UnlockedAt time.Time `json:"unlocked_at"`
}
