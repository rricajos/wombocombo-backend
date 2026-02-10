package models

import "time"

type FriendStatus string

const (
	FriendPending  FriendStatus = "pending"
	FriendAccepted FriendStatus = "accepted"
	FriendBlocked  FriendStatus = "blocked"
)

type Friend struct {
	PlayerID  string       `gorm:"type:uuid;primaryKey" json:"player_id"`
	FriendID  string       `gorm:"type:uuid;primaryKey;index" json:"friend_id"`
	Status    FriendStatus `gorm:"size:16;default:'pending'" json:"status"`
	CreatedAt time.Time    `json:"created_at"`

	Friend Player `gorm:"foreignKey:FriendID" json:"friend,omitempty"`
}
