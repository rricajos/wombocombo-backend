package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Player struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:32;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	DisplayName  string    `gorm:"size:64" json:"display_name"`
	AvatarID     string    `gorm:"size:32;default:'avatar_01'" json:"avatar_id"`
	IsBanned     bool      `gorm:"default:false" json:"is_banned"`
	IsAdmin      bool      `gorm:"default:false" json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Stats PlayerStats `gorm:"foreignKey:PlayerID" json:"stats,omitempty"`
}

func (p *Player) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	if p.DisplayName == "" {
		p.DisplayName = p.Username
	}
	return nil
}
