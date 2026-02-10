package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/wombocombo/api-server/utils"
)

func AuthRequired(jwtSecret string, redisClient *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{"error": "missing or invalid authorization header"})
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := utils.ValidateToken(tokenStr, jwtSecret)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		// Check session exists in Redis
		sessionKey := "session:" + claims.Subject
		exists, _ := redisClient.Exists(c.Context(), sessionKey).Result()
		if exists == 0 {
			return c.Status(401).JSON(fiber.Map{"error": "session expired"})
		}

		c.Locals("player_id", claims.Subject)
		c.Locals("username", claims.Username)
		return c.Next()
	}
}
