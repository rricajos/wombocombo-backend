package models

type MatchPlayer struct {
	MatchID        string `gorm:"type:uuid;primaryKey" json:"match_id"`
	PlayerID       string `gorm:"type:uuid;primaryKey;index" json:"player_id"`
	Kills          int    `json:"kills"`
	Deaths         int    `json:"deaths"`
	Score          int    `json:"score"`
	RoundsSurvived int    `json:"rounds_survived"`

	Player Player `gorm:"foreignKey:PlayerID" json:"player,omitempty"`
}
