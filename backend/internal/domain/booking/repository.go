package booking

// В этом файле описан интерфейс хранилища бронирований.

import "context"

// Repository описывает, что умеет слой работы с данными для модели Booking.
type Repository interface {
	List(ctx context.Context) ([]Booking, error)
	Get(ctx context.Context, id string) (Booking, error)
	Create(ctx context.Context, b Booking) (Booking, error)
	Delete(ctx context.Context, id string) error
}
