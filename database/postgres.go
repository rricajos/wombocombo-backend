package database

import (
	"github.com/rs/zerolog/log"
	"github.com/wombocombo/api-server/config"
	"github.com/wombocombo/api-server/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectPostgres(cfg *config.Config) *gorm.DB {
	logLevel := logger.Silent
	if cfg.IsDev() {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get sql.DB")
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	log.Info().Str("host", cfg.DBHost).Str("db", cfg.DBName).Msg("connected to PostgreSQL")
	return db
}

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.Player{},
		&models.PlayerStats{},
		&models.Match{},
		&models.MatchPlayer{},
		&models.Friend{},
		&models.InventoryItem{},
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to auto-migrate")
	}
	log.Info().Msg("database migrated successfully")
}
