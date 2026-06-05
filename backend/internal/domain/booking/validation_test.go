package booking_test

import (
	"Dormitory_Booking/internal/domain/booking"
	"errors"
	"testing"
	"time"
)

func TestValidateBasic_OK(t *testing.T) {
	start := time.Now().Add(2 * time.Hour)
	end := start.Add(1 * time.Hour)

	b := booking.Booking{
		Start:      start,
		End:        end,
		Room:       21,
		Title:      "OK",
		TelegramID: "123",
	}

	if err := b.ValidateBasic(); err != nil {
		t.Fatalf("ожидали nil, получили %v", err)
	}
}

func TestValidateBasic_InvalidPeriod(t *testing.T) {
	start := time.Now().Add(2 * time.Hour)
	end := start.Add(-10 * time.Minute)

	b := booking.Booking{
		Start:      start,
		End:        end,
		Room:       21,
		Title:      "Bad",
		TelegramID: "123",
	}

	err := b.ValidateBasic()
	if !errors.Is(err, booking.ErrInvalidPeriod) {
		t.Fatalf("ожидали ErrInvalidPeriod, получили %v", err)
	}
}

func TestValidateBasic_NoIdentifier(t *testing.T) {
	start := time.Now().Add(2 * time.Hour)
	b := booking.Booking{
		Start: start,
		End:   start.Add(time.Hour),
		Room:  21,
		Title: "No owner",
	}

	if err := b.ValidateBasic(); !errors.Is(err, booking.ErrNoIdentifier) {
		t.Fatalf("ожидали ErrNoIdentifier, получили %v", err)
	}
}
