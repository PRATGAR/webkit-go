package models

import (
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	Department   string    `json:"department"`
	Balance      float64   `json:"balance"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserProfile struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatar_url"`
	Website     string `json:"website"`
}

type UserSearchParams struct {
	Query      string `json:"query"`
	Department string `json:"department"`
	Role       string `json:"role"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

type TransferRequest struct {
	FromUserID int64   `json:"from_user_id"`
	ToUserID   int64   `json:"to_user_id"`
	Amount     float64 `json:"amount"`
	Note       string  `json:"note"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	ExpiresAt time.Time `json:"expires_at"`
}
