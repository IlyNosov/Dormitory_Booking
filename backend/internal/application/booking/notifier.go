package booking

import (
	"context"
	domain "Dormitory_Booking/internal/domain/booking"
)

// Notifier sends notifications when bookings are created or deleted.
type Notifier interface {
	NotifyNewBooking(ctx context.Context, b domain.Booking) error
	NotifyDeletedBooking(ctx context.Context, b domain.Booking) error
}
