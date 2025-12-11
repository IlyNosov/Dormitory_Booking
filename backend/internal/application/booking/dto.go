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
	Description string    `json:"description,omitempty"` // если пусто - фронт не увидит и не рисует кнопку "Подробнее"
	IsPrivate   bool      `json:"isPrivate"`
	TelegramID  string    `json:"telegramId"`
	CanManage   bool      `json:"canManage"`
}

func ToDTO(b domain.Booking, viewerID string, isAdmin bool) BookingDTO {
	return BookingDTO{
		ID:          b.ID,
		Start:       b.Start,
		End:         b.End,
		Room:        int(b.Room),
		Title:       b.Title,
		Description: b.Description,
		IsPrivate:   b.IsPrivate,
		TelegramID:  b.TelegramID,
		CanManage:   isAdmin || viewerID == b.TelegramID,
	}
}
