package dto

import "time"

// Auth responses
type AuthResponse struct {
	Player       PlayerResponse `json:"player"`
	Token        string         `json:"token"`
	RefreshToken string         `json:"refresh_token"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

// Player responses
type PlayerResponse struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"display_name"`
	AvatarID    string    `json:"avatar_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type PlayerPublicResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarID    string `json:"avatar_id"`
}

type PlayerStatsResponse struct {
	RoundsPlayed   int `json:"rounds_played"`
	RoundsSurvived int `json:"rounds_survived"`
	TotalKills     int `json:"total_kills"`
	TotalDeaths    int `json:"total_deaths"`
	BestRound      int `json:"best_round"`
	TotalPlaytime  int `json:"total_playtime_secs"`
	CoopRevives    int `json:"coop_revives"`
	Currency       int `json:"currency"`
}

// Room responses
type RoomResponse struct {
	ID         string   `json:"id"`
	JoinCode   string   `json:"join_code"`
	HostID     string   `json:"host_id"`
	MapID      string   `json:"map_id"`
	MaxPlayers int      `json:"max_players"`
	IsPublic   bool     `json:"is_public"`
	Status     string   `json:"status"`
	Players    []string `json:"players"`
}

// Match result (written by game server to Redis)
type MatchResult struct {
	MatchID         string              `json:"match_id"`
	RoomID          string              `json:"room_id"`
	StartedAt       time.Time           `json:"started_at"`
	EndedAt         time.Time           `json:"ended_at"`
	RoundsCompleted int                 `json:"rounds_completed"`
	MapID           string              `json:"map_id"`
	Players         []MatchPlayerResult `json:"players"`
}

type MatchPlayerResult struct {
	PlayerID       string `json:"player_id"`
	Kills          int    `json:"kills"`
	Deaths         int    `json:"deaths"`
	Score          int    `json:"score"`
	RoundsSurvived int    `json:"rounds_survived"`
}

// Generic
type DataResponse struct {
	Data interface{} `json:"data"`
}

type PaginatedResponse struct {
	Data    interface{} `json:"data"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
	Total   int64       `json:"total"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// Match history
type MatchHistoryEntry struct {
	MatchID         string              `json:"match_id"`
	RoomID          string              `json:"room_id"`
	MapID           string              `json:"map_id"`
	RoundsCompleted int                 `json:"rounds_completed"`
	StartedAt       time.Time           `json:"started_at"`
	EndedAt         time.Time           `json:"ended_at"`
	MyStats         MatchPlayerResult   `json:"my_stats"`
	Players         []MatchPlayerResult `json:"players"`
}

// Leaderboard
type LeaderboardEntry struct {
	Rank        int    `json:"rank"`
	PlayerID    string `json:"player_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarID    string `json:"avatar_id"`
	Value       int    `json:"value"`
}

// Friends
type FriendEntry struct {
	PlayerID    string    `json:"player_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarID    string    `json:"avatar_id"`
	Status      string    `json:"status"`
	IsOnline    bool      `json:"is_online"`
	Since       time.Time `json:"since"`
}

// Inventory
type InventoryEntry struct {
	ID         uint      `json:"id"`
	ItemType   string    `json:"item_type"`
	UnlockedAt time.Time `json:"unlocked_at"`
}

// Admin
type ServerStatsResponse struct {
	TotalPlayers   int64 `json:"total_players"`
	BannedPlayers  int64 `json:"banned_players"`
	ActiveSessions int64 `json:"active_sessions"`
	ActiveRooms    int64 `json:"active_rooms"`
	TotalMatches   int64 `json:"total_matches"`
}
