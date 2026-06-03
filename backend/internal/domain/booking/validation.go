package booking

import "time"

// В этом файле простые функции валидации.

func (b Booking) ValidateBasic() error {
	if !IsValidRoom(b.Room) {
		return ErrInvalidRoom
	}
	if !b.Start.After(time.Now()) {
		return ErrInPast
	}
	if !b.End.After(b.Start) {
		return ErrInvalidPeriod
	}
	return nil
}
