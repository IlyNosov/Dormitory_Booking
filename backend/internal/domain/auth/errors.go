package auth

import "errors"

var (
	ErrUserNotFound    = errors.New("пользователь не найден")
	ErrSessionNotFound = errors.New("сессия не найдена или истекла")
	ErrInvalidOTP      = errors.New("неверный или истёкший код")
	ErrTooManyRequests = errors.New("слишком много OTP-запросов, попробуйте через час")
	ErrInvalidEmail    = errors.New("email должен быть корпоративным (edu.hse.ru)")
	ErrUnauthorized    = errors.New("требуется авторизация")
)
