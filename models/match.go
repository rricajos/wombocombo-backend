package models

import "time"

type Match struct {
	ID              string    `gorm:"type:uuid;primaryKey" json:"id"`
	RoomID          string    `gorm:"size:32;index" json:"room_id"`
	StartedAt       time.Time `json:"started_at"`
	EndedAt         time.Time `json:"ended_at"`
	RoundsCompleted int       `json:"rounds_completed"`
	MapID           string    `gorm:"size:32" json:"map_id"`

	Players []MatchPlayer `gorm:"foreignKey:MatchID" json:"players,omitempty"`
}
