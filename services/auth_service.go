package services

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wombocombo/api-server/config"
	"github.com/wombocombo/api-server/dto"
	apperr "github.com/wombocombo/api-server/errors"
	"github.com/wombocombo/api-server/models"
	"github.com/wombocombo/api-server/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	db    *gorm.DB
	redis *redis.Client
	cfg   *config.Config
}

func NewAuthService(db *gorm.DB, redis *redis.Client, cfg *config.Config) *AuthService {
	return &AuthService{db: db, redis: redis, cfg: cfg}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*models.Player, string, string, error) {
	// Check unique username
	var count int64
	s.db.Model(&models.Player{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, "", "", apperr.Conflict("username already taken")
	}

	// Check unique email
	s.db.Model(&models.Player{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, "", "", apperr.Conflict("email already registered")
	}

	// Hash password
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, "", "", fmt.Errorf("hashing password: %w", err)
	}

	player := models.Player{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&player).Error; err != nil {
			return fmt.Errorf("creating player: %w", err)
		}
		stats := models.PlayerStats{PlayerID: player.ID}
		if err := tx.Create(&stats).Error; err != nil {
			return fmt.Errorf("creating player stats: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, "", "", err
	}

	// Generate tokens
	token, err := utils.GenerateToken(player.ID, player.Username, s.cfg.JWTSecret, s.cfg.JWTExpiration)
	if err != nil {
		return nil, "", "", fmt.Errorf("generating token: %w", err)
	}

	refreshToken, err := utils.GenerateToken(player.ID, player.Username, s.cfg.JWTSecret, s.cfg.JWTRefreshExpiry)
	if err != nil {
		return nil, "", "", fmt.Errorf("generating refresh token: %w", err)
	}

	// Store session in Redis
	s.storeSession(ctx, player.ID, token)

	// Cache player for game server
	s.cachePlayer(ctx, player)

	return &player, token, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*models.Player, string, string, error) {
	var player models.Player
	if err := s.db.Where("email = ?", req.Email).First(&player).Error; err != nil {
		return nil, "", "", apperr.New(401, "invalid email or password")
	}

	if player.IsBanned {
		return nil, "", "", apperr.New(403, "account is banned")
	}

	if !utils.CheckPassword(req.Password, player.PasswordHash) {
		return nil, "", "", apperr.New(401, "invalid email or password")
	}

	token, err := utils.GenerateToken(player.ID, player.Username, s.cfg.JWTSecret, s.cfg.JWTExpiration)
	if err != nil {
		return nil, "", "", fmt.Errorf("generating token: %w", err)
	}

	refreshToken, err := utils.GenerateToken(player.ID, player.Username, s.cfg.JWTSecret, s.cfg.JWTRefreshExpiry)
	if err != nil {
		return nil, "", "", fmt.Errorf("generating refresh token: %w", err)
	}

	s.storeSession(ctx, player.ID, token)
	s.cachePlayer(ctx, player)

	return &player, token, refreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, playerID string) error {
	s.redis.Del(ctx, "session:"+playerID)
	s.redis.Del(ctx, "player_cache:"+playerID)
	return nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshTokenStr string) (string, error) {
	claims, err := utils.ValidateToken(refreshTokenStr, s.cfg.JWTSecret)
	if err != nil {
		return "", apperr.New(401, "invalid refresh token")
	}

	newToken, err := utils.GenerateToken(claims.Subject, claims.Username, s.cfg.JWTSecret, s.cfg.JWTExpiration)
	if err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}

	s.storeSession(ctx, claims.Subject, newToken)
	return newToken, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	var player models.Player
	if err := s.db.Where("email = ?", email).First(&player).Error; err != nil {
		// Don't reveal if email exists
		return nil
	}

	resetToken := utils.GenerateRandomHex(32)
	s.redis.Set(ctx, "reset:"+resetToken, player.ID, 1*time.Hour)

	// TODO: send email in production. For now, log it.
	fmt.Printf("[DEV] Password reset token for %s: %s\n", email, resetToken)
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	playerID, err := s.redis.Get(ctx, "reset:"+token).Result()
	if err != nil {
		return apperr.New(400, "invalid or expired reset token")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.db.Model(&models.Player{}).Where("id = ?", playerID).Update("password_hash", hash).Error; err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	s.redis.Del(ctx, "reset:"+token)
	s.redis.Del(ctx, "session:"+playerID) // Force re-login
	return nil
}

func (s *AuthService) storeSession(ctx context.Context, playerID, token string) {
	key := "session:" + playerID
	s.redis.Set(ctx, key, token, s.cfg.JWTExpiration)
}

func (s *AuthService) cachePlayer(ctx context.Context, player models.Player) {
	key := "player_cache:" + player.ID
	s.redis.HSet(ctx, key,
		"id", player.ID,
		"username", player.Username,
		"display_name", player.DisplayName,
		"avatar_id", player.AvatarID,
	)
	s.redis.Expire(ctx, key, 24*time.Hour)
}
