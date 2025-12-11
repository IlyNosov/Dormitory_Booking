package booking

// В этом файле описана доменная модель бронирования.

import "time"

// Room - тип комнаты, чтобы не таскать магические числа.
type Room int

const (
	Room21  Room = 21
	Room132 Room = 132
	Room256 Room = 256
)

// Booking - основная модель бронирования.
type Booking struct {
	ID          string    `json:"id"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
	Room        Room      `json:"room"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"` // опциональное описание, показываем по кнопке "Подробнее"
	TelegramID  string    `json:"telegramId"`
	IsPrivate   bool      `json:"isPrivate"`
}

// IsValidRoom проверяет, что номер комнаты один из разрешённых.
func IsValidRoom(r Room) bool {
	switch r {
	case Room21, Room132, Room256:
		return true
	default:
		return false
	}
}
