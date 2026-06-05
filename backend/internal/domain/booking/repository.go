package booking

// В этом файле описан интерфейс хранилища бронирований

import "context"

// Repository описывает интерфейс слоя хранения для модели Booking
type Repository interface {
	List(ctx context.Context) ([]Booking, error)
	Get(ctx context.Context, id string) (Booking, error)
	Create(ctx context.Context, b Booking) (Booking, error)
	Update(ctx context.Context, b Booking) (Booking, error)
	Delete(ctx context.Context, id string) error
}
