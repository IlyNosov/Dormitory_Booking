package booking

// В этом файле лежит сервис для работы с бронированиями.
// Его дергают HTTP-слой, телеграм-бот и всё остальное.
// Здесь реализованы правила по времени, частным посиделкам, ограничениям по длительности и графику работы комнат.

import (
	"context"
	"time"

	domain "Dormitory_Booking/internal/domain/booking"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// CreateBookingInput - данные от HTTP/бота/парсера для создания брони.
type CreateBookingInput struct {
	Start       time.Time   // время начала брони
	End         time.Time   // время конца брони
	Room        domain.Room // комната
	Title       string      // название события
	Description string      // описание события
	TelegramID  string      // кто бронирует (Telegram ID)
	IsPrivate   bool        // частная посиделка или нет
}

// ListBookings возвращает все брони.
func (s *Service) ListBookings(ctx context.Context) ([]domain.Booking, error) {
	return s.repo.List(ctx)
}

// GetBooking возвращает бронь по ID.
func (s *Service) GetBooking(ctx context.Context, id string) (domain.Booking, error) {
	return s.repo.Get(ctx, id)
}

// DeleteBooking - удалить бронь может только владелец или админ.
func (s *Service) DeleteBooking(ctx context.Context, id string, requesterID string, isAdmin bool) error {
	b, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if !isAdmin && b.TelegramID != requesterID {
		return domain.ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}

// CreateBooking создаёт новую бронь с учётом всех правил.
func (s *Service) CreateBooking(ctx context.Context, in CreateBookingInput) (domain.Booking, error) {
	b := domain.Booking{
		Start:       in.Start,
		End:         in.End,
		Room:        in.Room,
		Title:       in.Title,
		Description: in.Description,
		TelegramID:  in.TelegramID,
		IsPrivate:   in.IsPrivate,
	}

	if err := b.ValidateBasic(); err != nil {
		return domain.Booking{}, err
	}

	// общие ограничения по длительности
	if err := validateDuration(b); err != nil {
		return domain.Booking{}, err
	}

	// ограничения по графику работы комнаты
	if err := validateRoomSchedule(b); err != nil {
		return domain.Booking{}, err
	}

	// частные посиделки: ночь, лимиты на день/вечер
	if b.IsPrivate {
		if err := s.validatePrivateRules(ctx, b); err != nil {
			return domain.Booking{}, err
		}
	}

	// проверка пересечений по времени в той же комнате
	existing, err := s.repo.List(ctx)
	if err != nil {
		return domain.Booking{}, err
	}
	for _, e := range existing {
		if e.Room != b.Room {
			continue
		}
		if timesOverlap(b.Start, b.End, e.Start, e.End) {
			return domain.Booking{}, domain.ErrOverlap
		}
	}

	return s.repo.Create(ctx, b)
}

// timesOverlap проверяет пересечение двух временных интервалов.
func timesOverlap(s1, e1, s2, e2 time.Time) bool {
	return s1.Before(e2) && e1.After(s2)
}

// Ограничения по длительности брони.

const maxBookingDuration = 3 * time.Hour

func validateDuration(b domain.Booking) error {
	dur := b.End.Sub(b.Start)
	if dur <= 0 {
		return domain.ErrInvalidPeriod
	}
	if dur > maxBookingDuration {
		return domain.ErrTooLongDuration
	}
	return nil
}

// График работы комнат.

// Для удобства храним часы работы в виде смещений от полуночи.
// Закрытие может быть позже 23:59 (например, 25:00 = 01:00 следующего дня).
type roomSchedule struct {
	WeekdayOpen  int // часы с 0 до 24
	WeekdayClose int
	FriSatOpen   int
	FriSatClose  int
	SunOpen      int
	SunClose     int
}

var roomSchedules = map[domain.Room]roomSchedule{
	domain.Room21: {
		WeekdayOpen: 6, WeekdayClose: 23,
		FriSatOpen: 6, FriSatClose: 25, // до 01:00
		SunOpen: 6, SunClose: 23,
	},
	domain.Room256: {
		WeekdayOpen: 6, WeekdayClose: 23,
		FriSatOpen: 6, FriSatClose: 25,
		SunOpen: 6, SunClose: 23,
	},
	domain.Room132: {
		WeekdayOpen: 6, WeekdayClose: 22,
		FriSatOpen: 6, FriSatClose: 23,
		SunOpen: 6, SunClose: 22,
	},
}

// validateRoomSchedule проверяет, что бронь целиком укладывается в разрешённые часы работы комнаты.
func validateRoomSchedule(b domain.Booking) error {
	sched, ok := roomSchedules[b.Room]
	if !ok {
		return domain.ErrInvalidRoom
	}

	loc := b.Start.Location()
	startLocal := b.Start.In(loc)
	endLocal := b.End.In(loc)

	dayStart := time.Date(startLocal.Year(), startLocal.Month(), startLocal.Day(), 0, 0, 0, 0, loc)
	var openHour, closeHour int

	switch startLocal.Weekday() {
	case time.Friday, time.Saturday:
		openHour = sched.FriSatOpen
		closeHour = sched.FriSatClose
	case time.Sunday:
		openHour = sched.SunOpen
		closeHour = sched.SunClose
	default:
		openHour = sched.WeekdayOpen
		closeHour = sched.WeekdayClose
	}

	openTime := dayStart.Add(time.Duration(openHour) * time.Hour)
	closeTime := dayStart.Add(time.Duration(closeHour) * time.Hour) // может быть > 24ч (до 01:00)

	if startLocal.Before(openTime) || endLocal.After(closeTime) {
		return domain.ErrInvalidTime
	}
	now := time.Now().In(loc)
	if startLocal.Before(now) {
		return domain.ErrInvalidTime
	}

	return nil
}

// "Частные посиделки" (ЧП)

// validatePrivateRules проверяет ночь, лимит ЧП в день и лимит вечерних ЧП.
func (s *Service) validatePrivateRules(ctx context.Context, b domain.Booking) error {
	loc := b.Start.Location()
	startLocal := b.Start.In(loc)
	endLocal := b.End.In(loc)

	// Нет ЧП в ночь с пятницы на субботу и с субботы на воскресенье в 23:00–06:00.
	if overlapsForbiddenPrivateNight(startLocal, endLocal) {
		return domain.ErrInvalidTime
	}

	// Не более 3 ЧП в день, не более одной ЧП после 18:00 по комнате.
	existing, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	dayY, dayM, dayD := startLocal.Date()
	privateCountDay := 0
	privateEveningCount := 0

	for _, e := range existing {
		if !e.IsPrivate || e.Room != b.Room {
			continue
		}

		eLocal := e.Start.In(loc)
		y, m, d := eLocal.Date()
		if y == dayY && m == dayM && d == dayD {
			privateCountDay++
			if eLocal.Hour() >= 18 {
				privateEveningCount++
			}
		}
	}

	if privateCountDay >= 3 {
		return domain.ErrPrivateDailyLimit
	}

	if startLocal.Hour() >= 18 && privateEveningCount >= 1 {
		return domain.ErrPrivateEveningLimit
	}

	return nil
}

// overlapsForbiddenPrivateNight проверяет, пересекает ли бронь ночные интервалы.
func overlapsForbiddenPrivateNight(start, end time.Time) bool {
	loc := start.Location()
	dayStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
	dayEnd := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, loc)

	for day := dayStart.Add(-24 * time.Hour); !day.After(dayEnd.Add(24 * time.Hour)); day = day.Add(24 * time.Hour) {
		wd := day.Weekday()
		if wd != time.Friday && wd != time.Saturday {
			continue
		}

		nightStart := time.Date(day.Year(), day.Month(), day.Day(), 23, 0, 0, 0, loc)
		nightEnd := nightStart.Add(7 * time.Hour) // до 06:00 следующего дня

		if timesOverlap(start, end, nightStart, nightEnd) {
			return true
		}
	}

	return false
}
