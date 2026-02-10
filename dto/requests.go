package dto

// Auth
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=32,alphanum"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// Player
type UpdatePlayerRequest struct {
	DisplayName *string `json:"display_name" validate:"omitempty,min=1,max=64"`
	AvatarID    *string `json:"avatar_id" validate:"omitempty,min=1,max=32"`
}

// Room
type CreateRoomRequest struct {
	MapID      string `json:"map_id" validate:"required,max=32"`
	MaxPlayers int    `json:"max_players" validate:"required,min=2,max=4"`
	IsPublic   bool   `json:"is_public"`
}

// Friends
type FriendRequest struct {
	FriendID string `json:"friend_id" validate:"required,uuid"`
}

// Inventory
type UnlockItemRequest struct {
	ItemType string `json:"item_type" validate:"required,min=1,max=64"`
}

// Admin
type BanPlayerRequest struct {
	PlayerID string `json:"player_id" validate:"required,uuid"`
	Reason   string `json:"reason" validate:"omitempty,max=255"`
}
