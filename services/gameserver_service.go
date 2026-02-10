package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/models"
	"github.com/wombocombo/api-server/utils"
	"gorm.io/gorm"
)

const (
	activeMatchTTL = 4 * time.Hour
	heartbeatTTL   = 5 * time.Minute
)

type GameServerService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewGameServerService(db *gorm.DB, redis *redis.Client) *GameServerService {
	return &GameServerService{db: db, redis: redis}
}

// ValidatePlayer checks a JWT token and returns player info + ban status.
// Called by the C++ game server when a player connects to a match.
func (s *GameServerService) ValidatePlayer(ctx context.Context, tokenStr, jwtSecret string) (*dto.PlayerValidation, error) {
	claims, err := utils.ValidateToken(tokenStr, jwtSecret)
	if err != nil {
		return &dto.PlayerValidation{Valid: false, Reason: "invalid or expired token"}, nil
	}

	// Check session exists
	exists, _ := s.redis.Exists(ctx, "session:"+claims.Subject).Result()
	if exists == 0 {
		return &dto.PlayerValidation{Valid: false, Reason: "session expired"}, nil
	}

	// Check ban status from cache or DB
	var player models.Player
	if err := s.db.Select("id, username, display_name, avatar_id, is_banned").
		Where("id = ?", claims.Subject).First(&player).Error; err != nil {
		return &dto.PlayerValidation{Valid: false, Reason: "player not found"}, nil
	}

	if player.IsBanned {
		return &dto.PlayerValidation{
			Valid:    false,
			PlayerID: player.ID,
			Reason:   "player is banned",
		}, nil
	}

	return &dto.PlayerValidation{
		Valid:       true,
		PlayerID:    player.ID,
		Username:    player.Username,
		DisplayName: player.DisplayName,
		AvatarID:    player.AvatarID,
	}, nil
}

// StartMatch records a match start in Redis for tracking.
func (s *GameServerService) StartMatch(ctx context.Context, req dto.MatchStartRequest) (*dto.MatchStartResponse, error) {
	if len(req.PlayerIDs) < 1 {
		return nil, apperr.BadRequest("at least one player required")
	}

	matchID := req.MatchID
	if matchID == "" {
		matchID = uuid.New().String()
	}

	activeMatch := dto.ActiveMatch{
		MatchID:   matchID,
		RoomID:    req.RoomID,
		MapID:     req.MapID,
		PlayerIDs: req.PlayerIDs,
		StartedAt: time.Now(),
		Status:    "playing",
	}

	data, err := json.Marshal(activeMatch)
	if err != nil {
		return nil, fmt.Errorf("marshaling active match: %w", err)
	}

	pipe := s.redis.Pipeline()
	pipe.Set(ctx, "active_match:"+matchID, data, activeMatchTTL)

	// Track each player's current match
	for _, pid := range req.PlayerIDs {
		pipe.Set(ctx, "player_match:"+pid, matchID, activeMatchTTL)
	}

	// Update room status to playing
	if req.RoomID != "" {
		s.updateRoomStatus(ctx, pipe, req.RoomID, "playing")
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("storing active match: %w", err)
	}

	log.Info().
		Str("match", matchID).
		Str("room", req.RoomID).
		Int("players", len(req.PlayerIDs)).
		Msg("match started")

	return &dto.MatchStartResponse{
		MatchID:   matchID,
		StartedAt: activeMatch.StartedAt,
	}, nil
}

// EndMatch processes a match result immediately (direct path, no Redis polling).
// This is the preferred path for the game server — faster than the worker poll cycle.
func (s *GameServerService) EndMatch(ctx context.Context, result dto.MatchResultInternal) error {
	if result.MatchID == "" {
		return apperr.BadRequest("match_id required")
	}

	// Persist to Postgres in a transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		match := models.Match{
			ID:              result.MatchID,
			RoomID:          result.RoomID,
			StartedAt:       result.StartedAt,
			EndedAt:         result.EndedAt,
			RoundsCompleted: result.RoundsCompleted,
			MapID:           result.MapID,
		}
		if err := tx.Create(&match).Error; err != nil {
			return fmt.Errorf("creating match: %w", err)
		}

		matchDurationSecs := int(result.EndedAt.Sub(result.StartedAt).Seconds())
		if matchDurationSecs < 0 {
			matchDurationSecs = 0
		}

		for _, p := range result.Players {
			mp := models.MatchPlayer{
				MatchID:        result.MatchID,
				PlayerID:       p.PlayerID,
				Kills:          p.Kills,
				Deaths:         p.Deaths,
				Score:          p.Score,
				RoundsSurvived: p.RoundsSurvived,
			}
			if err := tx.Create(&mp).Error; err != nil {
				return fmt.Errorf("creating match_player: %w", err)
			}

			// Calculate currency reward
			currency := calculateCurrencyReward(p, result.RoundsCompleted)

			// Update player stats + award currency + track playtime
			if err := tx.Exec(`
				UPDATE player_stats SET
					rounds_played = rounds_played + ?,
					rounds_survived = rounds_survived + ?,
					total_kills = total_kills + ?,
					total_deaths = total_deaths + ?,
					best_round = GREATEST(best_round, ?),
					total_playtime_secs = total_playtime_secs + ?,
					coop_revives = coop_revives + ?,
					currency = currency + ?,
					updated_at = NOW()
				WHERE player_id = ?
			`, result.RoundsCompleted, p.RoundsSurvived, p.Kills, p.Deaths,
				result.RoundsCompleted, matchDurationSecs, p.CoopRevives,
				currency, p.PlayerID).Error; err != nil {
				return fmt.Errorf("updating player stats: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("persisting match result: %w", err)
	}

	// Clean up Redis tracking data
	pipe := s.redis.Pipeline()
	pipe.Del(ctx, "active_match:"+result.MatchID)
	for _, p := range result.Players {
		pipe.Del(ctx, "player_match:"+p.PlayerID)
	}

	// Update room status to finished
	if result.RoomID != "" {
		s.updateRoomStatus(ctx, pipe, result.RoomID, "finished")
	}

	// Publish match_end event for any subscribers (e.g., notification service)
	eventData, _ := json.Marshal(map[string]interface{}{
		"event":    "match_end",
		"match_id": result.MatchID,
		"room_id":  result.RoomID,
		"players":  len(result.Players),
	})
	pipe.Publish(ctx, "game:events", eventData)

	pipe.Exec(ctx)

	log.Info().
		Str("match", result.MatchID).
		Int("players", len(result.Players)).
		Int("rounds", result.RoundsCompleted).
		Msg("match ended and persisted")

	return nil
}

// UpdateRoomStatus updates a room's status in Redis.
func (s *GameServerService) UpdateRoomStatus(ctx context.Context, roomID, status string) error {
	validStatuses := map[string]bool{"waiting": true, "playing": true, "finished": true}
	if !validStatuses[status] {
		return apperr.BadRequest("invalid status: must be waiting, playing, or finished")
	}

	data, err := s.redis.Get(ctx, "room:"+roomID).Result()
	if err == redis.Nil {
		return apperr.NotFound("room")
	}
	if err != nil {
		return fmt.Errorf("fetching room: %w", err)
	}

	var room models.Room
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return fmt.Errorf("unmarshaling room: %w", err)
	}

	room.Status = status
	updated, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("marshaling room: %w", err)
	}

	ttl, _ := s.redis.TTL(ctx, "room:"+roomID).Result()
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	s.redis.Set(ctx, "room:"+roomID, updated, ttl)

	// Remove from public set if finished
	if status == "finished" {
		s.redis.SRem(ctx, "rooms:public", roomID)
	}

	return nil
}

// PlayerHeartbeat refreshes session TTL for connected players.
// The game server calls this periodically for all connected players.
func (s *GameServerService) PlayerHeartbeat(ctx context.Context, playerIDs []string) (int, error) {
	if len(playerIDs) == 0 {
		return 0, nil
	}

	refreshed := 0
	pipe := s.redis.Pipeline()

	for _, pid := range playerIDs {
		pipe.Expire(ctx, "session:"+pid, 24*time.Hour)
		pipe.Set(ctx, "heartbeat:"+pid, time.Now().Unix(), heartbeatTTL)
	}

	results, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, fmt.Errorf("refreshing heartbeats: %w", err)
	}

	// Count successful refreshes (Expire returns true if key exists)
	for i := 0; i < len(results); i += 2 {
		if results[i].Err() == nil {
			refreshed++
		}
	}

	return refreshed, nil
}

// GetPlayerInfo returns cached player info for the game server.
func (s *GameServerService) GetPlayerInfo(ctx context.Context, playerID string) (*dto.GamePlayerInfo, error) {
	// Try cache first
	cached, err := s.redis.HGetAll(ctx, "player_cache:"+playerID).Result()
	if err == nil && len(cached) > 0 {
		return &dto.GamePlayerInfo{
			PlayerID:    cached["id"],
			Username:    cached["username"],
			DisplayName: cached["display_name"],
			AvatarID:    cached["avatar_id"],
		}, nil
	}

	// Fallback to DB
	var player models.Player
	if err := s.db.Select("id, username, display_name, avatar_id, is_banned").
		Where("id = ?", playerID).First(&player).Error; err != nil {
		return nil, apperr.NotFound("player")
	}

	// Re-cache
	s.redis.HSet(ctx, "player_cache:"+playerID,
		"id", player.ID,
		"username", player.Username,
		"display_name", player.DisplayName,
		"avatar_id", player.AvatarID,
	)
	s.redis.Expire(ctx, "player_cache:"+playerID, 24*time.Hour)

	return &dto.GamePlayerInfo{
		PlayerID:    player.ID,
		Username:    player.Username,
		DisplayName: player.DisplayName,
		AvatarID:    player.AvatarID,
		IsBanned:    player.IsBanned,
	}, nil
}

// GetActiveMatch returns info about an in-progress match.
func (s *GameServerService) GetActiveMatch(ctx context.Context, matchID string) (*dto.ActiveMatch, error) {
	data, err := s.redis.Get(ctx, "active_match:"+matchID).Result()
	if err == redis.Nil {
		return nil, apperr.NotFound("active match")
	}
	if err != nil {
		return nil, fmt.Errorf("fetching active match: %w", err)
	}

	var match dto.ActiveMatch
	if err := json.Unmarshal([]byte(data), &match); err != nil {
		return nil, fmt.Errorf("unmarshaling active match: %w", err)
	}
	return &match, nil
}

// updateRoomStatus is a helper that queues room status update in a pipeline.
func (s *GameServerService) updateRoomStatus(ctx context.Context, pipe redis.Pipeliner, roomID, status string) {
	data, err := s.redis.Get(ctx, "room:"+roomID).Result()
	if err != nil {
		return
	}

	var room models.Room
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return
	}

	room.Status = status
	updated, err := json.Marshal(room)
	if err != nil {
		return
	}

	ttl, _ := s.redis.TTL(ctx, "room:"+roomID).Result()
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	pipe.Set(ctx, "room:"+roomID, updated, ttl)

	if status == "finished" {
		pipe.SRem(ctx, "rooms:public", roomID)
	}
}

// calculateCurrencyReward computes currency earned from a match.
// Base: 10 per round + 5 per kill + 2 per revive + 15 survival bonus per round survived.
func calculateCurrencyReward(p dto.MatchPlayerResultInternal, totalRounds int) int {
	reward := totalRounds*10 + p.Kills*5 + p.CoopRevives*2 + p.RoundsSurvived*15
	if reward < 0 {
		reward = 0
	}
	return reward
}
