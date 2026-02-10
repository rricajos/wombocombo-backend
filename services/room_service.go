package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/models"
	"github.com/wombocombo/api-server/utils"
)

const roomTTL = 2 * time.Hour

type RoomService struct {
	redis *redis.Client
}

func NewRoomService(redis *redis.Client) *RoomService {
	return &RoomService{redis: redis}
}

func (s *RoomService) CreateRoom(ctx context.Context, hostID string, mapID string, maxPlayers int, isPublic bool) (*models.Room, error) {
	room := &models.Room{
		ID:         uuid.New().String(),
		JoinCode:   utils.GenerateRoomCode(),
		HostID:     hostID,
		MapID:      mapID,
		MaxPlayers: maxPlayers,
		IsPublic:   isPublic,
		Status:     "waiting",
		Players:    []string{hostID},
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(room)
	if err != nil {
		return nil, fmt.Errorf("marshaling room: %w", err)
	}

	// Store by ID and index by join_code
	pipe := s.redis.Pipeline()
	pipe.Set(ctx, "room:"+room.ID, data, roomTTL)
	pipe.Set(ctx, "room_code:"+room.JoinCode, room.ID, roomTTL)

	if isPublic {
		pipe.SAdd(ctx, "rooms:public", room.ID)
		pipe.Expire(ctx, "rooms:public", roomTTL)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("storing room in Redis: %w", err)
	}

	return room, nil
}

func (s *RoomService) GetByCode(ctx context.Context, code string) (*models.Room, error) {
	roomID, err := s.redis.Get(ctx, "room_code:"+code).Result()
	if err == redis.Nil {
		return nil, apperr.NotFound("room")
	}
	if err != nil {
		return nil, fmt.Errorf("fetching room by code: %w", err)
	}
	return s.GetByID(ctx, roomID)
}

func (s *RoomService) GetByID(ctx context.Context, id string) (*models.Room, error) {
	data, err := s.redis.Get(ctx, "room:"+id).Result()
	if err == redis.Nil {
		return nil, apperr.NotFound("room")
	}
	if err != nil {
		return nil, fmt.Errorf("fetching room: %w", err)
	}

	var room models.Room
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return nil, fmt.Errorf("unmarshaling room: %w", err)
	}
	return &room, nil
}

func (s *RoomService) ListPublic(ctx context.Context) ([]models.Room, error) {
	ids, err := s.redis.SMembers(ctx, "rooms:public").Result()
	if err != nil {
		return nil, fmt.Errorf("fetching public rooms: %w", err)
	}

	rooms := make([]models.Room, 0, len(ids))
	for _, id := range ids {
		room, err := s.GetByID(ctx, id)
		if err != nil {
			// Room expired, clean up
			s.redis.SRem(ctx, "rooms:public", id)
			continue
		}
		if room.Status == "waiting" {
			rooms = append(rooms, *room)
		}
	}
	return rooms, nil
}
