package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/services"
)

type RoomHandler struct {
	roomService *services.RoomService
}

func NewRoomHandler(roomService *services.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

func (h *RoomHandler) CreateRoom(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	var req dto.CreateRoomRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation failed", "details": err.Error()})
	}

	room, err := h.roomService.CreateRoom(c.Context(), playerID, req.MapID, req.MaxPlayers, req.IsPublic)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create room"})
	}

	return c.Status(201).JSON(dto.DataResponse{Data: room})
}

func (h *RoomHandler) GetRoom(c *fiber.Ctx) error {
	code := c.Params("code")

	room, err := h.roomService.GetByCode(c.Context(), code)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: room})
}

func (h *RoomHandler) ListPublicRooms(c *fiber.Ctx) error {
	rooms, err := h.roomService.ListPublic(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to list rooms"})
	}

	return c.JSON(dto.DataResponse{Data: rooms})
}
