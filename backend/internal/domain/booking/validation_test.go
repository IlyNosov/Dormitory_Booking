package booking_test

import (
	"Dormitory_Booking/internal/domain/booking"
	"errors"
	"testing"
	"time"
)

func TestValidateBasic_OK(t *testing.T) {
	start := time.Now().Add(2 * time.Hour)
	start = time.Date(start.Year(), start.Month(), start.Day(), 12, 0, 0, 0, time.Local)
	end := start.Add(1 * time.Hour)

	b := booking.Booking{
		Start:      start,
		End:        end,
		Room:       booking.Room21,
		Title:      "OK",
		TelegramID: "123",
	}

	if err := b.ValidateBasic(); err != nil {
		t.Fatalf("ожидали nil, получили %v", err)
	}
}

func TestValidateBasic_InvalidPeriod(t *testing.T) {
	start := time.Now().Add(2 * time.Hour)
	start = time.Date(start.Year(), start.Month(), start.Day(), 12, 0, 0, 0, time.Local)

	// end раньше start, но оба в будущем => проверка "прошлое" не мешает
	end := start.Add(-10 * time.Minute)

	b := booking.Booking{
		Start:      start,
		End:        end,
		Room:       booking.Room21,
		Title:      "Bad",
		TelegramID: "123",
	}

	err := b.ValidateBasic()
	if !errors.Is(err, booking.ErrInvalidPeriod) {
		t.Fatalf("ожидали ErrInvalidPeriod, получили %v", err)
	}
}

func TestValidateBasic_InvalidRoom(t *testing.T) {
	b := booking.Booking{
		Start: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC),
		Room:  999,
	}

	if err := b.ValidateBasic(); !errors.Is(err, booking.ErrInvalidRoom) {
		t.Fatalf("ожидали ErrInvalidRoom, получили %v", err)
	}
}
