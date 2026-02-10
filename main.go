package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/wombocombo/api-server/config"
	"github.com/wombocombo/api-server/database"
	mw "github.com/wombocombo/api-server/middleware"
	"github.com/wombocombo/api-server/routes"
	"github.com/wombocombo/api-server/workers"
)

func main() {
	// Load config
	cfg := config.Load()

	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if cfg.IsDev() {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().Str("env", cfg.Env).Msg("starting API server")

	// Connect databases
	db := database.ConnectPostgres(cfg)
	redisClient := database.ConnectRedis(cfg)

	// Auto-migrate in development
	if cfg.IsDev() {
		database.AutoMigrate(db)
	}

	// Publish JWT secret to Redis for game server
	database.PublishJWTSecret(context.Background(), redisClient, cfg.JWTSecret)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
		DisableStartupMessage: !cfg.IsDev(),
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(mw.RequestLogger())
	app.Use(mw.CORSConfig(os.Getenv("CORS_ORIGINS")))
	app.Use(mw.RateLimit(cfg.RateLimitMax, cfg.RateLimitWindow))

	// Register routes
	routes.Setup(app, db, redisClient, cfg)

	// Start background workers
	workerCtx, workerCancel := context.WithCancel(context.Background())
	if cfg.WorkersEnabled {
		go workers.StartCleanup(workerCtx, redisClient)
		go workers.StartMatchProcessor(workerCtx, redisClient, db)
		go workers.StartSessionHeartbeat(workerCtx, redisClient)
		log.Info().Msg("background workers started")
	} else {
		log.Warn().Msg("background workers disabled")
	}

	// Start HTTP server
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	// Stop workers first
	workerCancel()

	// Stop HTTP server
	app.Shutdown()

	// Close database connections
	sqlDB, _ := db.DB()
	sqlDB.Close()
	redisClient.Close()

	log.Info().Msg("server stopped")
}
