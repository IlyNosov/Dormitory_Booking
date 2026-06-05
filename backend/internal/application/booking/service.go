package booking

import (
	"context"
	"time"

	domain "Dormitory_Booking/internal/domain/booking"
	roomsvc "Dormitory_Booking/internal/application/room"
)

// UTC+3, FixedZone вместо LoadLocation чтобы не тащить tzdata в контейнер
var mskLoc = time.FixedZone("MSK", 3*60*60)


type Service struct {
	repo     domain.Repository
	roomsSvc *roomsvc.Service
	notifier Notifier
}

func NewService(repo domain.Repository, rooms *roomsvc.Service) *Service {
	return &Service{repo: repo, roomsSvc: rooms}
}

func NewServiceWithNotifier(repo domain.Repository, rooms *roomsvc.Service, n Notifier) *Service {
	return &Service{repo: repo, roomsSvc: rooms, notifier: n}
}

type CreateBookingInput struct {
	Start       time.Time
	End         time.Time
	Room        int
	Title       string
	Description string
	UserEmail   string
	TelegramID  string
	IsPrivate   bool
	Force       bool // только для админа, обходит бизнес-правила
}

func (s *Service) ListBookings(ctx context.Context) ([]domain.Booking, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetBooking(ctx context.Context, id string) (domain.Booking, error) {
	return s.repo.Get(ctx, id)
}

type UpdateBookingInput struct {
	ID          string
	Start       time.Time
	End         time.Time
	Room        int
	Title       string
	Description string
	TelegramID  string
	IsPrivate   bool
	Force       bool
}

func (s *Service) UpdateBooking(ctx context.Context, in UpdateBookingInput, requesterEmail string, isAdmin bool) (domain.Booking, error) {
	existing, err := s.repo.Get(ctx, in.ID)
	if err != nil {
		return domain.Booking{}, err
	}
	if !isAdmin && existing.UserEmail != requesterEmail {
		return domain.Booking{}, domain.ErrForbidden
	}

	updated := domain.Booking{
		ID:          in.ID,
		Start:       in.Start,
		End:         in.End,
		Room:        in.Room,
		Title:       sanitize(in.Title),
		Description: sanitize(in.Description),
		UserEmail:   existing.UserEmail,
		TelegramID:  in.TelegramID,
		IsPrivate:   in.IsPrivate,
	}

	active, err := s.roomsSvc.IsActive(ctx, updated.Room)
	if err != nil || !active {
		return domain.Booking{}, domain.ErrInvalidRoom
	}

	if in.Force || isAdmin {
		if err := updated.ValidateTimeOrder(); err != nil {
			return domain.Booking{}, err
		}
	} else {
		if err := updated.ValidateBasic(); err != nil {
			return domain.Booking{}, err
		}
		if err := s.validateDuration(updated); err != nil {
			return domain.Booking{}, err
		}
		if err := s.validateRoomSchedule(ctx, updated); err != nil {
			return domain.Booking{}, err
		}
		if err := s.validateUserDailyLimit(ctx, updated, in.ID); err != nil {
			return domain.Booking{}, err
		}
	}

	return s.repo.Update(ctx, updated)
}

func (s *Service) DeleteBooking(ctx context.Context, id, requesterEmail, requesterTG string, isAdmin bool) error {
	b, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if !isAdmin {
		ownsViaEmail := requesterEmail != "" && b.UserEmail == requesterEmail
		ownsViaTG := requesterTG != "" && b.TelegramID == requesterTG
		if !ownsViaEmail && !ownsViaTG {
			return domain.ErrForbidden
		}
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if s.notifier != nil {
		_ = s.notifier.NotifyDeletedBooking(ctx, b)
	}
	return nil
}

func (s *Service) CreateBooking(ctx context.Context, in CreateBookingInput) (domain.Booking, error) {
	b := domain.Booking{
		Start:       in.Start,
		End:         in.End,
		Room:        in.Room,
		Title:       sanitize(in.Title),
		Description: sanitize(in.Description),
		UserEmail:   in.UserEmail,
		TelegramID:  in.TelegramID,
		IsPrivate:   in.IsPrivate,
	}

	// проверяем комнату всегда, даже при форс-пуше
	active, err := s.roomsSvc.IsActive(ctx, b.Room)
	if err != nil || !active {
		return domain.Booking{}, domain.ErrInvalidRoom
	}

	if in.Force {
		// форс-пуш, только порядок времён
		if err := b.ValidateTimeOrder(); err != nil {
			return domain.Booking{}, err
		}
	} else {
		if err := b.ValidateBasic(); err != nil {
			return domain.Booking{}, err
		}
		if err := s.validateDuration(b); err != nil {
			return domain.Booking{}, err
		}
		if err := s.validateRoomSchedule(ctx, b); err != nil {
			return domain.Booking{}, err
		}
		if err := s.validateUserDailyLimit(ctx, b, ""); err != nil {
			return domain.Booking{}, err
		}
		if b.IsPrivate {
			if err := s.validatePrivateRules(ctx, b); err != nil {
				return domain.Booking{}, err
			}
		}
	}

	created, err := s.repo.Create(ctx, b)
	if err != nil {
		return domain.Booking{}, err
	}
	if s.notifier != nil {
		_ = s.notifier.NotifyNewBooking(ctx, created)
	}
	return created, nil
}

const maxBookingDuration = 4 * time.Hour

func (s *Service) validateDuration(b domain.Booking) error {
	dur := b.End.Sub(b.Start)
	if dur <= 0 {
		return domain.ErrInvalidPeriod
	}
	if dur > maxBookingDuration {
		return domain.ErrTooLongDuration
	}
	return nil
}

func (s *Service) validateRoomSchedule(ctx context.Context, b domain.Booking) error {
	rm, err := s.roomsSvc.GetSchedule(ctx, b.Room)
	if err != nil {
		return domain.ErrInvalidRoom
	}

	// расписание всегда проверяем по московскому времени
	loc := mskLoc
	startLocal := b.Start.In(loc)
	endLocal := b.End.In(loc)
	dayStart := time.Date(startLocal.Year(), startLocal.Month(), startLocal.Day(), 0, 0, 0, 0, loc)

	var openHour, closeHour int
	switch startLocal.Weekday() {
	case time.Friday, time.Saturday:
		openHour, closeHour = rm.FriSatOpen, rm.FriSatClose
	case time.Sunday:
		openHour, closeHour = rm.SunOpen, rm.SunClose
	default:
		openHour, closeHour = rm.WeekdayOpen, rm.WeekdayClose
	}

	openTime := dayStart.Add(time.Duration(openHour) * time.Hour)
	closeTime := dayStart.Add(time.Duration(closeHour) * time.Hour)

	if startLocal.Before(openTime) || endLocal.After(closeTime) {
		return domain.ErrInvalidTime
	}
	if startLocal.Before(time.Now().In(mskLoc)) {
		return domain.ErrInPast
	}
	return nil
}

func (s *Service) validatePrivateRules(ctx context.Context, b domain.Booking) error {
	loc := mskLoc
	startLocal := b.Start.In(loc)

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

func (s *Service) validateUserDailyLimit(ctx context.Context, b domain.Booking, excludeID string) error {
	existing, err := s.repo.List(ctx)
	if err != nil {
		return err
	}

	loc := mskLoc
	dayY, dayM, dayD := b.Start.In(loc).Date()

	for _, e := range existing {
		if e.ID == excludeID {
			continue
		}
		ey, em, ed := e.Start.In(loc).Date()
		if ey != dayY || em != dayM || ed != dayD {
			continue
		}
		sameUser := (b.UserEmail != "" && e.UserEmail == b.UserEmail) ||
			(b.TelegramID != "" && e.TelegramID == b.TelegramID)
		if sameUser {
			return domain.ErrDailyUserLimit
		}
	}
	return nil
}

// вырезает HTML-теги для защиты от XSS
func sanitize(s string) string {
	var result []rune
	inTag := false
	for _, c := range s {
		switch {
		case c == '<':
			inTag = true
		case c == '>':
			inTag = false
		case !inTag:
			result = append(result, c)
		}
	}
	r := []rune(string(result))
	if len(r) > 512 {
		r = r[:512]
	}
	return string(r)
}
