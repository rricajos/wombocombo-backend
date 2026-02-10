package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/services"
)

type FriendHandler struct {
	friendService *services.FriendService
}

func NewFriendHandler(friendService *services.FriendService) *FriendHandler {
	return &FriendHandler{friendService: friendService}
}

func (h *FriendHandler) SendRequest(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	var req dto.FriendRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation failed", "details": err.Error()})
	}

	if err := h.friendService.SendRequest(playerID, req.FriendID); err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(201).JSON(dto.MessageResponse{Message: "friend request sent"})
}

func (h *FriendHandler) AcceptRequest(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	var req dto.FriendRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation failed", "details": err.Error()})
	}

	if err := h.friendService.AcceptRequest(playerID, req.FriendID); err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.MessageResponse{Message: "friend request accepted"})
}

func (h *FriendHandler) RemoveFriend(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)
	friendID := c.Params("id")

	if err := h.friendService.RemoveFriend(playerID, friendID); err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.MessageResponse{Message: "friend removed"})
}

func (h *FriendHandler) BlockPlayer(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	var req dto.FriendRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation failed", "details": err.Error()})
	}

	if err := h.friendService.BlockPlayer(playerID, req.FriendID); err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.MessageResponse{Message: "player blocked"})
}

func (h *FriendHandler) ListFriends(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	friends, err := h.friendService.ListFriends(c.Context(), playerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	entries := make([]dto.FriendEntry, 0, len(friends))
	for _, f := range friends {
		entries = append(entries, dto.FriendEntry{
			PlayerID:    f.FriendID,
			Username:    f.Friend.Username,
			DisplayName: f.Friend.DisplayName,
			AvatarID:    f.Friend.AvatarID,
			Status:      string(f.Status),
			IsOnline:    h.friendService.IsOnline(c.Context(), f.FriendID),
			Since:       f.CreatedAt,
		})
	}

	return c.JSON(dto.DataResponse{Data: entries})
}

func (h *FriendHandler) ListPendingRequests(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	requests, err := h.friendService.ListPendingRequests(playerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	entries := make([]dto.FriendEntry, 0, len(requests))
	for _, r := range requests {
		entries = append(entries, dto.FriendEntry{
			PlayerID:    r.PlayerID,
			Username:    r.Friend.Username,
			DisplayName: r.Friend.DisplayName,
			AvatarID:    r.Friend.AvatarID,
			Status:      string(r.Status),
			IsOnline:    h.friendService.IsOnline(c.Context(), r.PlayerID),
			Since:       r.CreatedAt,
		})
	}

	return c.JSON(dto.DataResponse{Data: entries})
}
