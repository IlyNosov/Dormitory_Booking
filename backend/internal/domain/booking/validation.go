package booking

import "time"

// ValidateBasic проверяет порядок времён и наличие хотя бы одного идентификатора
func (b Booking) ValidateBasic() error {
	if !b.Start.After(time.Now()) {
		return ErrInPast
	}
	if !b.End.After(b.Start) {
		return ErrInvalidPeriod
	}
	if b.UserEmail == "" && b.TelegramID == "" {
		return ErrNoIdentifier
	}
	return nil
}

// ValidateTimeOrder проверяет только что end > start, используется для форс-пуша
func (b Booking) ValidateTimeOrder() error {
	if !b.End.After(b.Start) {
		return ErrInvalidPeriod
	}
	return nil
}
