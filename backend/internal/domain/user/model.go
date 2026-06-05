package user

import "time"

type User struct {
	ID         string
	Email      string
	TelegramID string
	CreatedAt  time.Time
}
