package models

import "time"

// Room lives only in Redis, not persisted in Postgres.
type Room struct {
	ID        string    `json:"id"`
	JoinCode  string    `json:"join_code"`
	HostID    string    `json:"host_id"`
	MapID     string    `json:"map_id"`
	MaxPlayers int      `json:"max_players"`
	IsPublic  bool      `json:"is_public"`
	Status    string    `json:"status"` // waiting, playing, finished
	Players   []string  `json:"players"`
	CreatedAt time.Time `json:"created_at"`
}
