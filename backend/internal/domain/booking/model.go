package booking

import "time"

type Booking struct {
	ID          string
	Start       time.Time
	End         time.Time
	Room        int
	Title       string
	Description string
	UserEmail   string
	TelegramID  string
	IsPrivate   bool
	TgMsgID     int // ID сообщения в Telegram для привязки отмены
}
