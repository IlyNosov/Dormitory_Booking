package booking_test

import (
	"testing"
	"time"

	app "Dormitory_Booking/internal/application/booking"
	domain "Dormitory_Booking/internal/domain/booking"
)

func TestToDTO_TGOwnerCanManage(t *testing.T) {
	b := domain.Booking{
		ID:          "42",
		Start:       time.Now(),
		End:         time.Now().Add(time.Hour),
		Room:        21,
		Title:       "Тестовая бронь",
		Description: "Описание",
		TelegramID:  "111",
		IsPrivate:   true,
	}

	dto := app.ToDTO(b, "", "111", false)

	if !dto.CanManage {
		t.Fatalf("владелец по TG должен иметь CanManage=true")
	}
	if dto.Description != b.Description {
		t.Fatalf("описание должно прокидываться в DTO")
	}
	if dto.Room != b.Room {
		t.Fatalf("номер комнаты должен совпадать")
	}
}

func TestToDTO_EmailOwnerCanManage(t *testing.T) {
	b := domain.Booking{
		ID:        "43",
		Title:     "Бронь по email",
		UserEmail: "student@edu.hse.ru",
	}

	dto := app.ToDTO(b, "student@edu.hse.ru", "", false)

	if !dto.CanManage {
		t.Fatalf("владелец по email должен иметь CanManage=true")
	}
}

func TestToDTO_NonOwnerCannotManage(t *testing.T) {
	b := domain.Booking{
		ID:         "1",
		Title:      "Бронь",
		TelegramID: "owner",
	}

	dto := app.ToDTO(b, "stranger@edu.hse.ru", "other", false)

	if dto.CanManage {
		t.Fatalf("чужой не должен иметь CanManage=true")
	}
}

func TestToDTO_AdminCanManage(t *testing.T) {
	b := domain.Booking{
		ID:         "1",
		Title:      "Бронь",
		TelegramID: "owner",
	}

	dto := app.ToDTO(b, "whoever@edu.hse.ru", "", true)

	if !dto.CanManage {
		t.Fatalf("админ всегда должен иметь CanManage=true")
	}
}
