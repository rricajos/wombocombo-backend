package services

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/models"
	"gorm.io/gorm"
)

type FriendService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewFriendService(db *gorm.DB, redis *redis.Client) *FriendService {
	return &FriendService{db: db, redis: redis}
}

func (s *FriendService) SendRequest(playerID, friendID string) error {
	if playerID == friendID {
		return apperr.BadRequest("cannot add yourself as a friend")
	}

	// Check target player exists
	var count int64
	s.db.Model(&models.Player{}).Where("id = ?", friendID).Count(&count)
	if count == 0 {
		return apperr.NotFound("player")
	}

	// Check if blocked by the other player
	var blocked models.Friend
	err := s.db.Where("player_id = ? AND friend_id = ? AND status = ?",
		friendID, playerID, models.FriendBlocked).First(&blocked).Error
	if err == nil {
		return apperr.Forbidden("cannot send friend request to this player")
	}

	// Check existing relationship
	var existing models.Friend
	err = s.db.Where("player_id = ? AND friend_id = ?", playerID, friendID).First(&existing).Error
	if err == nil {
		switch existing.Status {
		case models.FriendAccepted:
			return apperr.Conflict("already friends")
		case models.FriendPending:
			return apperr.Conflict("friend request already sent")
		case models.FriendBlocked:
			return apperr.BadRequest("unblock this player first")
		}
	}

	// Check if the other player already sent us a request → auto-accept
	var incoming models.Friend
	err = s.db.Where("player_id = ? AND friend_id = ? AND status = ?",
		friendID, playerID, models.FriendPending).First(&incoming).Error
	if err == nil {
		// Mutual request → accept both directions
		return s.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&incoming).Update("status", models.FriendAccepted).Error; err != nil {
				return err
			}
			reverse := models.Friend{
				PlayerID: playerID,
				FriendID: friendID,
				Status:   models.FriendAccepted,
			}
			return tx.Create(&reverse).Error
		})
	}

	// Create pending request
	friend := models.Friend{
		PlayerID: playerID,
		FriendID: friendID,
		Status:   models.FriendPending,
	}
	if err := s.db.Create(&friend).Error; err != nil {
		return fmt.Errorf("creating friend request: %w", err)
	}

	return nil
}

func (s *FriendService) AcceptRequest(playerID, friendID string) error {
	// Find the incoming pending request
	var request models.Friend
	err := s.db.Where("player_id = ? AND friend_id = ? AND status = ?",
		friendID, playerID, models.FriendPending).First(&request).Error
	if err != nil {
		return apperr.NotFound("friend request")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update incoming request to accepted
		if err := tx.Model(&request).Update("status", models.FriendAccepted).Error; err != nil {
			return err
		}
		// Create reverse relationship
		reverse := models.Friend{
			PlayerID: playerID,
			FriendID: friendID,
			Status:   models.FriendAccepted,
		}
		return tx.Create(&reverse).Error
	})
}

func (s *FriendService) RemoveFriend(playerID, friendID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where(
			"(player_id = ? AND friend_id = ?) OR (player_id = ? AND friend_id = ?)",
			playerID, friendID, friendID, playerID,
		).Delete(&models.Friend{})

		if result.RowsAffected == 0 {
			return apperr.NotFound("friend relationship")
		}
		return result.Error
	})
}

func (s *FriendService) BlockPlayer(playerID, friendID string) error {
	if playerID == friendID {
		return apperr.BadRequest("cannot block yourself")
	}

	// Check target exists
	var count int64
	s.db.Model(&models.Player{}).Where("id = ?", friendID).Count(&count)
	if count == 0 {
		return apperr.NotFound("player")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Remove any existing relationships in both directions
		tx.Where(
			"(player_id = ? AND friend_id = ?) OR (player_id = ? AND friend_id = ?)",
			playerID, friendID, friendID, playerID,
		).Delete(&models.Friend{})

		// Create block entry
		block := models.Friend{
			PlayerID: playerID,
			FriendID: friendID,
			Status:   models.FriendBlocked,
		}
		return tx.Create(&block).Error
	})
}

func (s *FriendService) ListFriends(ctx context.Context, playerID string) ([]models.Friend, error) {
	var friends []models.Friend
	err := s.db.Where("player_id = ? AND status = ?", playerID, models.FriendAccepted).
		Preload("Friend").
		Find(&friends).Error
	if err != nil {
		return nil, fmt.Errorf("fetching friends: %w", err)
	}
	return friends, nil
}

func (s *FriendService) ListPendingRequests(playerID string) ([]models.Friend, error) {
	var requests []models.Friend
	// Incoming requests: someone sent request to us
	err := s.db.Where("friend_id = ? AND status = ?", playerID, models.FriendPending).
		Preload("Friend").
		Find(&requests).Error
	if err != nil {
		return nil, fmt.Errorf("fetching pending requests: %w", err)
	}

	// Swap: the "Friend" preload loaded the requester via FriendID FK,
	// but for incoming requests we want to show the player who sent it (PlayerID).
	// Re-query with proper direction.
	var incoming []models.Friend
	err = s.db.Table("friends").
		Select("friends.*").
		Where("friends.friend_id = ? AND friends.status = ?", playerID, models.FriendPending).
		Find(&incoming).Error
	if err != nil {
		return nil, fmt.Errorf("fetching incoming requests: %w", err)
	}

	// Load the sender player data
	if len(incoming) > 0 {
		senderIDs := make([]string, len(incoming))
		for i, r := range incoming {
			senderIDs[i] = r.PlayerID
		}
		var senders []models.Player
		s.db.Where("id IN ?", senderIDs).Find(&senders)
		senderMap := make(map[string]models.Player)
		for _, p := range senders {
			senderMap[p.ID] = p
		}
		for i := range incoming {
			if p, ok := senderMap[incoming[i].PlayerID]; ok {
				incoming[i].Friend = p
			}
		}
	}

	return incoming, nil
}

func (s *FriendService) IsOnline(ctx context.Context, playerID string) bool {
	exists, _ := s.redis.Exists(ctx, "session:"+playerID).Result()
	return exists > 0
}
