package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret          string
	JWTExpiration      time.Duration
	JWTRefreshExpiry   time.Duration

	// Rate Limit
	RateLimitMax    int
	RateLimitWindow time.Duration
}

func Load() *Config {
	return &Config{
		Port: getEnv("PORT", "3000"),
		Env:  getEnv("ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "gameuser"),
		DBPassword: getEnv("DB_PASSWORD", "gamepass"),
		DBName:     getEnv("DB_NAME", "gamedb"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiration:    time.Duration(getEnvInt("JWT_EXP_HOURS", 24)) * time.Hour,
		JWTRefreshExpiry: time.Duration(getEnvInt("JWT_REFRESH_EXP_DAYS", 7)) * 24 * time.Hour,

		RateLimitMax:    getEnvInt("RATE_LIMIT_MAX", 60),
		RateLimitWindow: time.Duration(getEnvInt("RATE_LIMIT_WINDOW_SECS", 60)) * time.Second,
	}
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode +
		" TimeZone=UTC"
}

func (c *Config) IsDev() bool {
	return c.Env == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
