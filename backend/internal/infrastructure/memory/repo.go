package memory

// В этом файле лежит in-memory репозиторий для бронирований.

import (
	"context"
	"sync"
	"time"

	"Dormitory_Booking/internal/domain/booking"

	"github.com/google/uuid"
)

type InMemoryBookingRepo struct {
	mu       sync.RWMutex
	bookings map[string]booking.Booking
}

func NewInMemoryBookingRepo() *InMemoryBookingRepo {
	return &InMemoryBookingRepo{
		bookings: make(map[string]booking.Booking),
	}
}

// List возвращает все брони.
func (r *InMemoryBookingRepo) List(ctx context.Context) ([]booking.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]booking.Booking, 0, len(r.bookings))
	for _, b := range r.bookings {
		out = append(out, b)
	}

	return out, nil
}

// Get возвращает бронь по ID.
func (r *InMemoryBookingRepo) Get(ctx context.Context, id string) (booking.Booking, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.bookings[id]
	if !ok {
		return booking.Booking{}, booking.ErrNotFound
	}
	return b, nil
}

// Create создаёт бронь. Если у брони нет ID, генерируем новый UUID.
func (r *InMemoryBookingRepo) Create(ctx context.Context, b booking.Booking) (booking.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	if b.Start.IsZero() {
		b.Start = time.Now()
	}
	if b.End.IsZero() {
		b.End = b.Start.Add(time.Hour)
	}

	r.bookings[b.ID] = b
	return b, nil
}

// Delete удаляет бронь, если она существует.
func (r *InMemoryBookingRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.bookings[id]; !ok {
		return booking.ErrNotFound
	}
	delete(r.bookings, id)
	return nil
}
