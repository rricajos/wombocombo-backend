package services

import (
	"fmt"

	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/models"
	"github.com/wombocombo/api-server/utils"
	"gorm.io/gorm"
)

type StatsService struct {
	db *gorm.DB
}

func NewStatsService(db *gorm.DB) *StatsService {
	return &StatsService{db: db}
}

func (s *StatsService) GetMatchHistory(playerID string, pg utils.Pagination) ([]dto.MatchHistoryEntry, int64, error) {
	var total int64
	s.db.Model(&models.MatchPlayer{}).Where("player_id = ?", playerID).Count(&total)

	if total == 0 {
		return []dto.MatchHistoryEntry{}, 0, nil
	}

	// Fetch match IDs for this player, ordered by match start time desc
	var matchPlayers []models.MatchPlayer
	err := s.db.Where("player_id = ?", playerID).
		Order("match_id DESC").
		Offset(pg.Offset()).Limit(pg.PerPage).
		Find(&matchPlayers).Error
	if err != nil {
		return nil, 0, fmt.Errorf("fetching match history: %w", err)
	}

	matchIDs := make([]string, len(matchPlayers))
	myStatsMap := make(map[string]models.MatchPlayer)
	for i, mp := range matchPlayers {
		matchIDs[i] = mp.MatchID
		myStatsMap[mp.MatchID] = mp
	}

	// Fetch full match data with all players
	var matches []models.Match
	err = s.db.Where("id IN ?", matchIDs).
		Preload("Players").
		Preload("Players.Player").
		Order("started_at DESC").
		Find(&matches).Error
	if err != nil {
		return nil, 0, fmt.Errorf("fetching matches: %w", err)
	}

	entries := make([]dto.MatchHistoryEntry, 0, len(matches))
	for _, m := range matches {
		my := myStatsMap[m.ID]
		players := make([]dto.MatchPlayerResult, 0, len(m.Players))
		for _, p := range m.Players {
			players = append(players, dto.MatchPlayerResult{
				PlayerID:       p.PlayerID,
				Kills:          p.Kills,
				Deaths:         p.Deaths,
				Score:          p.Score,
				RoundsSurvived: p.RoundsSurvived,
			})
		}

		entries = append(entries, dto.MatchHistoryEntry{
			MatchID:         m.ID,
			RoomID:          m.RoomID,
			MapID:           m.MapID,
			RoundsCompleted: m.RoundsCompleted,
			StartedAt:       m.StartedAt,
			EndedAt:         m.EndedAt,
			MyStats: dto.MatchPlayerResult{
				PlayerID:       my.PlayerID,
				Kills:          my.Kills,
				Deaths:         my.Deaths,
				Score:          my.Score,
				RoundsSurvived: my.RoundsSurvived,
			},
			Players: players,
		})
	}

	return entries, total, nil
}

func (s *StatsService) GetLeaderboard(sortBy string, pg utils.Pagination) ([]dto.LeaderboardEntry, int64, error) {
	// Validate sort column
	allowedColumns := map[string]string{
		"kills":    "total_kills",
		"deaths":   "total_deaths",
		"rounds":   "rounds_played",
		"survived": "rounds_survived",
		"playtime": "total_playtime_secs",
		"revives":  "coop_revives",
	}

	column, ok := allowedColumns[sortBy]
	if !ok {
		return nil, 0, apperr.BadRequest("invalid sort parameter. allowed: kills, deaths, rounds, survived, playtime, revives")
	}

	var total int64
	s.db.Model(&models.PlayerStats{}).Count(&total)

	var results []struct {
		models.PlayerStats
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		AvatarID    string `json:"avatar_id"`
	}

	err := s.db.Table("player_stats").
		Select("player_stats.*, players.username, players.display_name, players.avatar_id").
		Joins("JOIN players ON players.id = player_stats.player_id").
		Where("players.is_banned = ?", false).
		Order(column + " DESC").
		Offset(pg.Offset()).Limit(pg.PerPage).
		Find(&results).Error
	if err != nil {
		return nil, 0, fmt.Errorf("fetching leaderboard: %w", err)
	}

	entries := make([]dto.LeaderboardEntry, len(results))
	for i, r := range results {
		var value int
		switch sortBy {
		case "kills":
			value = r.TotalKills
		case "deaths":
			value = r.TotalDeaths
		case "rounds":
			value = r.RoundsPlayed
		case "survived":
			value = r.RoundsSurvived
		case "playtime":
			value = r.TotalPlaytime
		case "revives":
			value = r.CoopRevives
		}

		entries[i] = dto.LeaderboardEntry{
			Rank:        pg.Offset() + i + 1,
			PlayerID:    r.PlayerID,
			Username:    r.Username,
			DisplayName: r.DisplayName,
			AvatarID:    r.AvatarID,
			Value:       value,
		}
	}

	return entries, total, nil
}
