package booking_test

import (
	"testing"
	"time"

	app "Dormitory_Booking/internal/application/booking"
	domain "Dormitory_Booking/internal/domain/booking"
)

func TestToDTO_OwnerCanManage(t *testing.T) {
	b := domain.Booking{
		ID:          "42",
		Start:       time.Now(),
		End:         time.Now().Add(time.Hour),
		Room:        domain.Room21,
		Title:       "Тестовая бронь",
		Description: "Описание тестовой брони",
		TelegramID:  "111",
		IsPrivate:   true,
	}

	dto := app.ToDTO(b, "111", false)

	if !dto.CanManage {
		t.Fatalf("владелец должен иметь CanManage=true")
	}
	if dto.Description != b.Description {
		t.Fatalf("описание должно прокидываться в DTO")
	}
	if dto.Room != int(b.Room) {
		t.Fatalf("номер комнаты должен совпадать")
	}
}

func TestToDTO_NonOwnerCannotManage(t *testing.T) {
	b := domain.Booking{
		ID:         "1",
		Title:      "Бронь",
		TelegramID: "owner",
	}

	dto := app.ToDTO(b, "stranger", false)

	if dto.CanManage {
		t.Fatalf("человек без прав не должен иметь CanManage=true")
	}
}

func TestToDTO_AdminCanManage(t *testing.T) {
	b := domain.Booking{
		ID:         "1",
		Title:      "Бронь",
		TelegramID: "owner",
	}

	dto := app.ToDTO(b, "whoever", true)

	if !dto.CanManage {
		t.Fatalf("админ всегда должен иметь CanManage=true")
	}
}
