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
	// ── Services ──────────────────────────────────────────────────
	authService := services.NewAuthService(db, redisClient, cfg)
	playerService := services.NewPlayerService(db)
	roomService := services.NewRoomService(redisClient)
	statsService := services.NewStatsService(db)
	friendService := services.NewFriendService(db, redisClient)
	inventoryService := services.NewInventoryService(db)
	adminService := services.NewAdminService(db, redisClient)

	// ── Handlers ──────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authService)
	playerHandler := handlers.NewPlayerHandler(playerService)
	roomHandler := handlers.NewRoomHandler(roomService)
	statsHandler := handlers.NewStatsHandler(statsService)
	friendHandler := handlers.NewFriendHandler(friendService)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService)
	adminHandler := handlers.NewAdminHandler(adminService)

	// ── Middleware ─────────────────────────────────────────────────
	auth := middleware.AuthRequired(cfg.JWTSecret, redisClient)
	admin := middleware.AdminRequired(db)

	// ── API group ─────────────────────────────────────────────────
	api := app.Group("/api")

	// Health
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ── Auth routes (public) ──────────────────────────────────────
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/forgot-password", authHandler.ForgotPassword)
	authGroup.Post("/reset-password", authHandler.ResetPassword)

	// Auth routes (protected)
	authGroup.Post("/logout", auth, authHandler.Logout)

	// ── Player routes (protected) ─────────────────────────────────
	players := api.Group("/players", auth)
	players.Get("/me", playerHandler.GetMe)
	players.Patch("/me", playerHandler.UpdateMe)
	players.Get("/:id", playerHandler.GetPlayer)
	players.Get("/:id/stats", playerHandler.GetPlayerStats)

	// ── Room routes (protected) ───────────────────────────────────
	rooms := api.Group("/rooms", auth)
	rooms.Post("/", roomHandler.CreateRoom)
	rooms.Get("/public", roomHandler.ListPublicRooms)
	rooms.Get("/:code", roomHandler.GetRoom)

	// ── Stats routes (protected) ──────────────────────────────────
	stats := api.Group("/stats", auth)
	stats.Get("/history/:id", statsHandler.GetMatchHistory) // :id or "me"
	stats.Get("/leaderboard", statsHandler.GetLeaderboard)  // ?sort=kills&page=1&per_page=20

	// ── Friends routes (protected) ────────────────────────────────
	friends := api.Group("/friends", auth)
	friends.Get("/", friendHandler.ListFriends)
	friends.Get("/pending", friendHandler.ListPendingRequests)
	friends.Post("/request", friendHandler.SendRequest)
	friends.Post("/accept", friendHandler.AcceptRequest)
	friends.Delete("/:id", friendHandler.RemoveFriend)
	friends.Post("/block", friendHandler.BlockPlayer)

	// ── Inventory routes (protected) ──────────────────────────────
	inventory := api.Group("/inventory", auth)
	inventory.Get("/", inventoryHandler.ListItems)
	inventory.Post("/unlock", inventoryHandler.UnlockItem)
	inventory.Get("/catalog", inventoryHandler.GetCatalog)

	// ── Admin routes (protected + admin) ──────────────────────────
	adminGroup := api.Group("/admin", auth, admin)
	adminGroup.Post("/ban", adminHandler.BanPlayer)
	adminGroup.Post("/unban", adminHandler.UnbanPlayer)
	adminGroup.Get("/stats", adminHandler.GetServerStats)
}
