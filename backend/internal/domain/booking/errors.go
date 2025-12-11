package booking

import "errors"

var (
	ErrNotFound      = errors.New("booking not found")
	ErrOverlap       = errors.New("booking overlaps with existing one")
	ErrInvalidPeriod = errors.New("end must be after start")
	ErrForbidden     = errors.New("operation is not allowed for this user")
	ErrInvalidRoom   = errors.New("invalid room number")

	ErrInvalidTime         = errors.New("booking is not allowed at this time")
	ErrPrivateDailyLimit   = errors.New("private bookings daily limit exceeded")
	ErrPrivateEveningLimit = errors.New("private evening bookings limit exceeded")
	ErrTooLongDuration     = errors.New("booking duration exceeds maximum allowed")
)
