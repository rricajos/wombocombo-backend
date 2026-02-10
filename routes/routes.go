package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/wombocombo/api-server/config"
	"github.com/wombocombo/api-server/handlers"
	"github.com/wombocombo/api-server/middleware"
	"github.com/wombocombo/api-server/services"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB, redisClient *redis.Client, cfg *config.Config) {
	// Services
	authService := services.NewAuthService(db, redisClient, cfg)
	playerService := services.NewPlayerService(db)
	roomService := services.NewRoomService(redisClient)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	playerHandler := handlers.NewPlayerHandler(playerService)
	roomHandler := handlers.NewRoomHandler(roomService)

	// Auth middleware
	auth := middleware.AuthRequired(cfg.JWTSecret, redisClient)

	// API group
	api := app.Group("/api")

	// Health
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Auth routes (public)
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/forgot-password", authHandler.ForgotPassword)
	authGroup.Post("/reset-password", authHandler.ResetPassword)

	// Auth routes (protected)
	authGroup.Post("/logout", auth, authHandler.Logout)

	// Player routes (protected)
	players := api.Group("/players", auth)
	players.Get("/me", playerHandler.GetMe)
	players.Patch("/me", playerHandler.UpdateMe)
	players.Get("/:id", playerHandler.GetPlayer)
	players.Get("/:id/stats", playerHandler.GetPlayerStats)

	// Room routes (protected)
	rooms := api.Group("/rooms", auth)
	rooms.Post("/", roomHandler.CreateRoom)
	rooms.Get("/public", roomHandler.ListPublicRooms)
	rooms.Get("/:code", roomHandler.GetRoom)
}
