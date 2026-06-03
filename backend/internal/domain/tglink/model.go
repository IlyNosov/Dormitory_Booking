package tglink

import "errors"

// Link хранит состояние привязки сессии сайта к Telegram-аккаунту.
type Link struct {
	SessionID  string
	TelegramID string // числовой ID пользователя в Telegram, заполняется после подтверждения
	Token      string // одноразовый токен, который пользователь отправляет боту
	Confirmed  bool
}

var (
	ErrNotFound      = errors.New("привязка не найдена")
	ErrAlreadyLinked = errors.New("сессия уже привязана к Telegram")
	ErrTokenExpired  = errors.New("токен не найден или устарел")
)

// Repository — интерфейс хранилища ссылок.
type Repository interface {
	// Upsert создаёт или обновляет запись для sessionID.
	Upsert(link Link) error
	// GetBySession возвращает запись по sessionID.
	GetBySession(sessionID string) (Link, error)
	// GetByToken возвращает запись по токену (только неподтверждённые).
	GetByToken(token string) (Link, error)
	// Confirm помечает запись подтверждённой и сохраняет telegramID.
	Confirm(token string, telegramID string) error
	// Unlink сбрасывает привязку для сессии.
	Unlink(sessionID string) error
}
