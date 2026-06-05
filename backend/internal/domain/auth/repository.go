package auth

import (
	"context"
	"time"
)

type UserRepository interface {
	// Upsert создаёт или обновляет пользователя. telegramID обновляется только если непустой.
	Upsert(ctx context.Context, email, telegramID string, isAdmin, isSuperAdmin bool) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, s Session) error
	Get(ctx context.Context, token string) (Session, error)
	Delete(ctx context.Context, token string) error
}

type OTPRepository interface {
	CreateOTP(ctx context.Context, o OTPCode) error
	// GetLatestUnused возвращает последний неиспользованный OTP для email.
	GetLatestUnused(ctx context.Context, email string) (OTPCode, error)
	// CountSince считает количество OTP-запросов для email после указанного момента.
	CountSince(ctx context.Context, email string, since time.Time) (int, error)
	MarkUsed(ctx context.Context, id string) error
}
