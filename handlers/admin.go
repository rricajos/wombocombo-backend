package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/services"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) BanPlayer(c *fiber.Ctx) error {
	var req dto.BanPlayerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation failed", "details": err.Error()})
	}

	if err := h.adminService.BanPlayer(c.Context(), req.PlayerID); err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.MessageResponse{Message: "player banned"})
}

func (h *AdminHandler) UnbanPlayer(c *fiber.Ctx) error {
	var req dto.BanPlayerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation failed", "details": err.Error()})
	}

	if err := h.adminService.UnbanPlayer(req.PlayerID); err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.MessageResponse{Message: "player unbanned"})
}

func (h *AdminHandler) GetServerStats(c *fiber.Ctx) error {
	stats, err := h.adminService.GetServerStats(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.DataResponse{Data: stats})
}
