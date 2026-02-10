package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/services"
)

type InternalHandler struct {
	gameService *services.GameServerService
	jwtSecret   string
}

func NewInternalHandler(gameService *services.GameServerService, jwtSecret string) *InternalHandler {
	return &InternalHandler{gameService: gameService, jwtSecret: jwtSecret}
}

// ValidatePlayer checks if a player's token is valid and returns their info.
// POST /internal/player/validate  {"token": "jwt_string"}
func (h *InternalHandler) ValidatePlayer(c *fiber.Ctx) error {
	var req dto.ValidatePlayerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Token == "" {
		return c.Status(400).JSON(fiber.Map{"error": "token required"})
	}

	result, err := h.gameService.ValidatePlayer(c.Context(), req.Token, h.jwtSecret)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: result})
}

// GetPlayerInfo returns player info for the game server.
// GET /internal/player/:id
func (h *InternalHandler) GetPlayerInfo(c *fiber.Ctx) error {
	playerID := c.Params("id")

	info, err := h.gameService.GetPlayerInfo(c.Context(), playerID)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: info})
}

// PlayerHeartbeat refreshes session TTL for connected players.
// POST /internal/player/heartbeat  {"player_ids": ["uuid1", "uuid2"]}
func (h *InternalHandler) PlayerHeartbeat(c *fiber.Ctx) error {
	var req dto.HeartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	refreshed, err := h.gameService.PlayerHeartbeat(c.Context(), req.PlayerIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: fiber.Map{
		"refreshed": refreshed,
		"total":     len(req.PlayerIDs),
	}})
}

// MatchStart notifies that a match has started.
// POST /internal/match/start
func (h *InternalHandler) MatchStart(c *fiber.Ctx) error {
	var req dto.MatchStartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation failed", "details": err.Error()})
	}

	result, err := h.gameService.StartMatch(c.Context(), req)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(201).JSON(dto.DataResponse{Data: result})
}

// MatchEnd submits a match result for immediate processing.
// POST /internal/match/end
func (h *InternalHandler) MatchEnd(c *fiber.Ctx) error {
	var req dto.MatchResultInternal
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.MatchID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "match_id required"})
	}

	if err := h.gameService.EndMatch(c.Context(), req); err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.MessageResponse{Message: "match result processed"})
}

// GetActiveMatch returns info about a match in progress.
// GET /internal/match/:id
func (h *InternalHandler) GetActiveMatch(c *fiber.Ctx) error {
	matchID := c.Params("id")

	match, err := h.gameService.GetActiveMatch(c.Context(), matchID)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: match})
}

// UpdateRoomStatus updates a room's status.
// PATCH /internal/room/status
func (h *InternalHandler) UpdateRoomStatus(c *fiber.Ctx) error {
	var req dto.RoomStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.RoomID == "" || req.Status == "" {
		return c.Status(400).JSON(fiber.Map{"error": "room_id and status required"})
	}

	if err := h.gameService.UpdateRoomStatus(c.Context(), req.RoomID, req.Status); err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.MessageResponse{Message: "room status updated"})
}
