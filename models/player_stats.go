package models

import "time"

type PlayerStats struct {
	PlayerID       string    `gorm:"type:uuid;primaryKey" json:"player_id"`
	RoundsPlayed   int       `gorm:"default:0" json:"rounds_played"`
	RoundsSurvived int       `gorm:"default:0" json:"rounds_survived"`
	TotalKills     int       `gorm:"default:0" json:"total_kills"`
	TotalDeaths    int       `gorm:"default:0" json:"total_deaths"`
	BestRound      int       `gorm:"default:0" json:"best_round"`
	TotalPlaytime  int       `gorm:"default:0" json:"total_playtime_secs"`
	CoopRevives    int       `gorm:"default:0" json:"coop_revives"`
	Currency       int       `gorm:"default:0" json:"currency"`
	UpdatedAt      time.Time `json:"updated_at"`
}
