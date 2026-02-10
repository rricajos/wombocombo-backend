package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/models"
	"github.com/wombocombo/api-server/services"
)

type PlayerHandler struct {
	playerService *services.PlayerService
}

func NewPlayerHandler(playerService *services.PlayerService) *PlayerHandler {
	return &PlayerHandler{playerService: playerService}
}

func (h *PlayerHandler) GetMe(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	player, err := h.playerService.GetByID(playerID)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: toPlayerResponse(player)})
}

func (h *PlayerHandler) UpdateMe(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	var req dto.UpdatePlayerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation failed", "details": err.Error()})
	}

	player, err := h.playerService.UpdateProfile(playerID, req)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: toPlayerResponse(player)})
}

func (h *PlayerHandler) GetPlayer(c *fiber.Ctx) error {
	id := c.Params("id")

	player, err := h.playerService.GetByID(id)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: dto.PlayerPublicResponse{
		ID:          player.ID,
		Username:    player.Username,
		DisplayName: player.DisplayName,
		AvatarID:    player.AvatarID,
	}})
}

func (h *PlayerHandler) GetPlayerStats(c *fiber.Ctx) error {
	id := c.Params("id")

	stats, err := h.playerService.GetStats(id)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: stats})
}

// toPlayerResponse maps a Player model to the API response (no password hash).
func toPlayerResponse(p *models.Player) dto.PlayerResponse {
	return dto.PlayerResponse{
		ID:          p.ID,
		Username:    p.Username,
		Email:       p.Email,
		DisplayName: p.DisplayName,
		AvatarID:    p.AvatarID,
		CreatedAt:   p.CreatedAt,
	}
}
