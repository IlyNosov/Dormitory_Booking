package booking

import (
	"time"

	domain "Dormitory_Booking/internal/domain/booking"
)

type BookingDTO struct {
	ID          string    `json:"id"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Room        int       `json:"room"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	IsPrivate   bool      `json:"isPrivate"`
	UserEmail   string    `json:"userEmail,omitempty"`
	TelegramID  string    `json:"telegramId,omitempty"`
	CanManage   bool      `json:"canManage"`
}

func ToDTO(b domain.Booking, viewerEmail, viewerTG string, isAdmin bool) BookingDTO {
	canManage := isAdmin ||
		(viewerEmail != "" && b.UserEmail == viewerEmail) ||
		(viewerTG != "" && b.TelegramID == viewerTG)

	dto := BookingDTO{
		ID:        b.ID,
		Start:     b.Start,
		End:       b.End,
		Room:      b.Room,
		Title:     b.Title,
		IsPrivate: b.IsPrivate,
		CanManage: canManage,
	}
	if b.Description != "" {
		dto.Description = b.Description
	}
	// TelegramID is public — visible to everyone for contact purposes
	if b.TelegramID != "" {
		dto.TelegramID = b.TelegramID
	}
	if isAdmin || canManage {
		dto.UserEmail = b.UserEmail
	}
	return dto
}
