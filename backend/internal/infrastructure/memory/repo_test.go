package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"Dormitory_Booking/internal/domain/booking"
	"Dormitory_Booking/internal/infrastructure/memory"
)

func newBooking() booking.Booking {
	return booking.Booking{
		Start:      time.Now(),
		End:        time.Now().Add(time.Hour),
		Room:       booking.Room21,
		Title:      "Test",
		TelegramID: "123",
	}
}

func TestMemoryRepo_CreateAndGet(t *testing.T) {
	r := memory.NewInMemoryBookingRepo()
	ctx := context.Background()

	b := newBooking()

	created, err := r.Create(ctx, b)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	got, err := r.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if got.Title != b.Title {
		t.Fatalf("ожидался заголовок %s, получили %s", b.Title, got.Title)
	}
}

func TestMemoryRepo_Delete(t *testing.T) {
	r := memory.NewInMemoryBookingRepo()
	ctx := context.Background()

	b := newBooking()
	created, _ := r.Create(ctx, b)

	err := r.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка при удалении: %v", err)
	}

	_, err = r.Get(ctx, created.ID)
	if !errors.Is(err, booking.ErrNotFound) {
		t.Fatalf("ожидалось booking.ErrNotFound, получили %v", err)
	}
}

func TestMemoryRepo_List(t *testing.T) {
	r := memory.NewInMemoryBookingRepo()
	ctx := context.Background()

	r.Create(ctx, newBooking())
	r.Create(ctx, newBooking())

	list, err := r.List(ctx)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("ожидалось 2, получили %d", len(list))
	}
}
