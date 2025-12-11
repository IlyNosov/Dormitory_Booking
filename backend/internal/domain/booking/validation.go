package booking

// В этом файле простые функции валидации. Они завязаны только на Booking и не требуют внешних зависимостей.

func (b Booking) ValidateBasic() error {
	if !IsValidRoom(b.Room) {
		return ErrInvalidRoom
	}
	if !b.End.After(b.Start) {
		return ErrInvalidPeriod
	}
	return nil
}
