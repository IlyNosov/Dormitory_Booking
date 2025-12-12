package booking

import "errors"

var (
	ErrNotFound            = errors.New("Бронь не найдена.")
	ErrOverlap             = errors.New("Бронь пересекается с существующей.")
	ErrInvalidPeriod       = errors.New("Время окончания должно быть позже времени начала.")
	ErrForbidden           = errors.New("Операция запрещена для этого пользователя.")
	ErrInvalidRoom         = errors.New("Недопустимый номер комнаты.")
	ErrInPast              = errors.New("Нельзя создавать бронь в прошлом.")
	ErrInvalidTime         = errors.New("Бронирование не разрешено в это время.")
	ErrPrivateDailyLimit   = errors.New("Превышен суточный лимит частных бронирований.")
	ErrPrivateEveningLimit = errors.New("Превышен вечерний лимит частных бронирований.")
	ErrTooLongDuration     = errors.New("Длительность бронирования превышает максимально допустимую.")
)
