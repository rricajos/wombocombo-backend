package middleware

import (
	"crypto/subtle"

	"github.com/gofiber/fiber/v2"
)

// InternalAuth validates requests from the game server using a shared secret.
// The game server sends the secret in the X-Server-Key header.
func InternalAuth(serverSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-Server-Key")
		if key == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing server key"})
		}

		// Constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(key), []byte(serverSecret)) != 1 {
			return c.Status(403).JSON(fiber.Map{"error": "invalid server key"})
		}

		return c.Next()
	}
}
