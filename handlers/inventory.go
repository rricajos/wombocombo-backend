package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/services"
)

type InventoryHandler struct {
	inventoryService *services.InventoryService
}

func NewInventoryHandler(inventoryService *services.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

func (h *InventoryHandler) ListItems(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	items, err := h.inventoryService.ListItems(playerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	entries := make([]dto.InventoryEntry, len(items))
	for i, item := range items {
		entries[i] = dto.InventoryEntry{
			ID:         item.ID,
			ItemType:   item.ItemType,
			UnlockedAt: item.UnlockedAt,
		}
	}

	return c.JSON(dto.DataResponse{Data: entries})
}

func (h *InventoryHandler) UnlockItem(c *fiber.Ctx) error {
	playerID := c.Locals("player_id").(string)

	var req dto.UnlockItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if err := validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "validation failed", "details": err.Error()})
	}

	item, err := h.inventoryService.UnlockItem(playerID, req.ItemType)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.Status(201).JSON(dto.DataResponse{Data: dto.InventoryEntry{
		ID:         item.ID,
		ItemType:   item.ItemType,
		UnlockedAt: item.UnlockedAt,
	}})
}

func (h *InventoryHandler) GetCatalog(c *fiber.Ctx) error {
	return c.JSON(dto.DataResponse{Data: h.inventoryService.GetCatalog()})
}
