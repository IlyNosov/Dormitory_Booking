package auth

import "time"

type User struct {
	ID           string
	Email        string
	TelegramID   string
	IsAdmin      bool
	IsSuperAdmin bool
	CreatedAt    time.Time
}

type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}

type OTPCode struct {
	ID        string
	Email     string
	Code      string
	CreatedAt time.Time
	ExpiresAt time.Time
	Used      bool
}
