package dto

import "time"

// ── Game Server Requests ──────────────────────────────────────────

// ValidatePlayerRequest is sent by the game server to validate a player's JWT.
type ValidatePlayerRequest struct {
	Token string `json:"token"`
}

// MatchStartRequest is sent by the game server when a match begins.
type MatchStartRequest struct {
	MatchID   string   `json:"match_id"`                                    // Optional, generated if empty
	RoomID    string   `json:"room_id" validate:"required"`
	MapID     string   `json:"map_id" validate:"required"`
	PlayerIDs []string `json:"player_ids" validate:"required,min=1,dive,uuid"`
}

// MatchResultInternal is the match result format the game server sends.
// Extends the base MatchResult with coop_revives per player.
type MatchResultInternal struct {
	MatchID         string                      `json:"match_id"`
	RoomID          string                      `json:"room_id"`
	StartedAt       time.Time                   `json:"started_at"`
	EndedAt         time.Time                   `json:"ended_at"`
	RoundsCompleted int                         `json:"rounds_completed"`
	MapID           string                      `json:"map_id"`
	Players         []MatchPlayerResultInternal `json:"players"`
}

// MatchPlayerResultInternal extends MatchPlayerResult with coop revive tracking.
type MatchPlayerResultInternal struct {
	PlayerID       string `json:"player_id"`
	Kills          int    `json:"kills"`
	Deaths         int    `json:"deaths"`
	Score          int    `json:"score"`
	RoundsSurvived int    `json:"rounds_survived"`
	CoopRevives    int    `json:"coop_revives"`
}

// HeartbeatRequest is sent periodically by the game server for connected players.
type HeartbeatRequest struct {
	PlayerIDs []string `json:"player_ids"`
}

// RoomStatusRequest updates a room's status.
type RoomStatusRequest struct {
	RoomID string `json:"room_id"`
	Status string `json:"status"` // waiting, playing, finished
}

// ── Game Server Responses ─────────────────────────────────────────

// PlayerValidation is returned when the game server validates a player token.
type PlayerValidation struct {
	Valid       bool   `json:"valid"`
	PlayerID    string `json:"player_id,omitempty"`
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	AvatarID    string `json:"avatar_id,omitempty"`
	Reason      string `json:"reason,omitempty"` // Only set if Valid=false
}

// MatchStartResponse is returned when a match is registered.
type MatchStartResponse struct {
	MatchID   string    `json:"match_id"`
	StartedAt time.Time `json:"started_at"`
}

// ActiveMatch represents a match currently in progress (stored in Redis).
type ActiveMatch struct {
	MatchID   string    `json:"match_id"`
	RoomID    string    `json:"room_id"`
	MapID     string    `json:"map_id"`
	PlayerIDs []string  `json:"player_ids"`
	StartedAt time.Time `json:"started_at"`
	Status    string    `json:"status"` // playing, finished
}

// GamePlayerInfo is the player info returned to the game server.
type GamePlayerInfo struct {
	PlayerID    string `json:"player_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarID    string `json:"avatar_id"`
	IsBanned    bool   `json:"is_banned"`
}
