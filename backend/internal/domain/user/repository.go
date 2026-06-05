package user

import "context"

type Repository interface {
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
	Upsert(ctx context.Context, email string) (User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, token, userID string) error
	Get(ctx context.Context, token string) (User, error)
	Delete(ctx context.Context, token string) error
}

type OTPRepository interface {
	Create(ctx context.Context, email, code string) error
	Verify(ctx context.Context, email, code string) (bool, error)
	CountRecent(ctx context.Context, email string) (int, error)
}
