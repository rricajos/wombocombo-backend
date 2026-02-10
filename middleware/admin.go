package middleware

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AdminRequired(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		playerID := c.Locals("player_id").(string)

		var isAdmin bool
		err := db.Table("players").
			Select("is_admin").
			Where("id = ?", playerID).
			Scan(&isAdmin).Error
		if err != nil || !isAdmin {
			return c.Status(403).JSON(fiber.Map{"error": "admin access required"})
		}

		return c.Next()
	}
}
