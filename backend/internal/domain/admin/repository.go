package admin

import "context"

type Repository interface {
	List(ctx context.Context) ([]Admin, error)
	IsAdmin(ctx context.Context, email string) (bool, error)
	Add(ctx context.Context, email, addedBy string) error
	Remove(ctx context.Context, email string) error
}
