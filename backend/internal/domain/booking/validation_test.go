package booking_test

import (
	"Dormitory_Booking/internal/domain/booking"
	"errors"
	"testing"
	"time"
)

func TestValidateBasic_OK(t *testing.T) {
	b := booking.Booking{
		Start: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC),
		Room:  booking.Room21,
	}

	if err := b.ValidateBasic(); err != nil {
		t.Fatalf("ожидали nil, получили %v", err)
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

func TestValidateBasic_InvalidPeriod(t *testing.T) {
	b := booking.Booking{
		Start: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		End:   time.Date(2025, 1, 1, 11, 0, 0, 0, time.UTC),
		Room:  21,
	}

	if err := b.ValidateBasic(); !errors.Is(err, booking.ErrInvalidPeriod) {
		t.Fatalf("ожидали ErrInvalidPeriod, получили %v", err)
	}
}
