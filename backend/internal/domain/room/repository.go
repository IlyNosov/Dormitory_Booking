package room

import "context"

type Repository interface {
	List(ctx context.Context) ([]Room, error)
	Get(ctx context.Context, number int) (Room, error)
	Create(ctx context.Context, r Room) (Room, error)
	Delete(ctx context.Context, number int) error
}
