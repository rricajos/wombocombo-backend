package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/services"
	"github.com/wombocombo/api-server/utils"
)

type StatsHandler struct {
	statsService *services.StatsService
}

func NewStatsHandler(statsService *services.StatsService) *StatsHandler {
	return &StatsHandler{statsService: statsService}
}

func (h *StatsHandler) GetMatchHistory(c *fiber.Ctx) error {
	playerID := c.Params("id")
	if playerID == "me" {
		playerID = c.Locals("player_id").(string)
	}

	pg := utils.ParsePagination(c)

	entries, total, err := h.statsService.GetMatchHistory(playerID, pg)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.PaginatedResponse{
		Data:    entries,
		Page:    pg.Page,
		PerPage: pg.PerPage,
		Total:   total,
	})
}

func (h *StatsHandler) GetLeaderboard(c *fiber.Ctx) error {
	sortBy := c.Query("sort", "kills")
	pg := utils.ParsePagination(c)

	entries, total, err := h.statsService.GetLeaderboard(sortBy, pg)
	if err != nil {
		if appErr, ok := err.(*apperr.AppError); ok {
			return apperr.SendError(c, appErr)
		}
		return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(dto.PaginatedResponse{
		Data:    entries,
		Page:    pg.Page,
		PerPage: pg.PerPage,
		Total:   total,
	})
}
