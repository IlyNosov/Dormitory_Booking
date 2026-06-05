package booking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	app "Dormitory_Booking/internal/application/booking"
	approom "Dormitory_Booking/internal/application/room"
	domain "Dormitory_Booking/internal/domain/booking"
	roomdomain "Dormitory_Booking/internal/domain/room"
)

// fakeRepo is a simple in-memory booking repository for testing.
type fakeRepo struct {
	data map[string]domain.Booking
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{data: make(map[string]domain.Booking)}
}

func (r *fakeRepo) List(_ context.Context) ([]domain.Booking, error) {
	res := make([]domain.Booking, 0, len(r.data))
	for _, b := range r.data {
		res = append(res, b)
	}
	return res, nil
}
func (r *fakeRepo) Get(_ context.Context, id string) (domain.Booking, error) {
	b, ok := r.data[id]
	if !ok {
		return domain.Booking{}, domain.ErrNotFound
	}
	return b, nil
}
func (r *fakeRepo) Create(_ context.Context, b domain.Booking) (domain.Booking, error) {
	if b.ID == "" {
		b.ID = "id-" + b.Start.Format("150405")
	}
	for _, e := range r.data {
		if e.Room == b.Room && e.Start.Before(b.End) && e.End.After(b.Start) {
			return domain.Booking{}, domain.ErrOverlap
		}
	}
	r.data[b.ID] = b
	return b, nil
}
func (r *fakeRepo) Update(_ context.Context, b domain.Booking) (domain.Booking, error) {
	if _, ok := r.data[b.ID]; !ok {
		return domain.Booking{}, domain.ErrNotFound
	}
	r.data[b.ID] = b
	return b, nil
}
func (r *fakeRepo) Delete(_ context.Context, id string) error {
	if _, ok := r.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.data, id)
	return nil
}

// fakeRoomRepo always reports rooms as active with a generous schedule.
type fakeRoomRepo struct{}

func (r *fakeRoomRepo) List(_ context.Context) ([]roomdomain.Room, error) {
	return []roomdomain.Room{
		{Number: 21, Name: "Комната 21", IsActive: true, WeekdayOpen: 0, WeekdayClose: 24, FriSatOpen: 0, FriSatClose: 24, SunOpen: 0, SunClose: 24},
		{Number: 132, Name: "Комната 132", IsActive: true, WeekdayOpen: 0, WeekdayClose: 24, FriSatOpen: 0, FriSatClose: 24, SunOpen: 0, SunClose: 24},
		{Number: 256, Name: "Комната 256", IsActive: true, WeekdayOpen: 0, WeekdayClose: 24, FriSatOpen: 0, FriSatClose: 24, SunOpen: 0, SunClose: 24},
	}, nil
}
func (r *fakeRoomRepo) Get(_ context.Context, number int) (roomdomain.Room, error) {
	return roomdomain.Room{
		Number: number, Name: "Room", IsActive: true,
		WeekdayOpen: 0, WeekdayClose: 24, FriSatOpen: 0, FriSatClose: 24, SunOpen: 0, SunClose: 24,
	}, nil
}
func (r *fakeRoomRepo) Create(_ context.Context, rm roomdomain.Room) (roomdomain.Room, error) {
	return rm, nil
}
func (r *fakeRoomRepo) Delete(_ context.Context, _ int) error { return nil }

func newTestService() *app.Service {
	return app.NewService(newFakeRepo(), approom.NewService(&fakeRoomRepo{}))
}

func futureInterval() (time.Time, time.Time) {
	now := time.Now()
	day := now.Add(24 * time.Hour)
	start := time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, time.Local)
	return start, start.Add(time.Hour)
}

func TestService_CreateBooking_OK(t *testing.T) {
	svc := newTestService()
	start, end := futureInterval()

	created, err := svc.CreateBooking(context.Background(), app.CreateBookingInput{
		Start:      start,
		End:        end,
		Room:       21,
		Title:      "Лаба по матану",
		TelegramID: "123",
	})
	if err != nil {
		t.Fatalf("ожидали nil, получили %v", err)
	}
	if created.ID == "" {
		t.Fatal("ID должен быть установлен")
	}
}

func TestService_CreateBooking_Overlap(t *testing.T) {
	repo := newFakeRepo()
	svc := app.NewService(repo, approom.NewService(&fakeRoomRepo{}))
	ctx := context.Background()

	start, end := futureInterval()
	repo.data["1"] = domain.Booking{ID: "1", Start: start, End: end, Room: 21, TelegramID: "111"}

	_, err := svc.CreateBooking(ctx, app.CreateBookingInput{
		Start: start.Add(30 * time.Minute), End: end.Add(30 * time.Minute),
		Room: 21, Title: "Overlap", TelegramID: "222",
	})
	if !errors.Is(err, domain.ErrOverlap) {
		t.Fatalf("ожидали ErrOverlap, получили %v", err)
	}
}

func TestService_ListBookings(t *testing.T) {
	repo := newFakeRepo()
	svc := app.NewService(repo, approom.NewService(&fakeRoomRepo{}))

	start, end := futureInterval()
	repo.data["1"] = domain.Booking{ID: "1", Start: start, End: end, Title: "A"}
	repo.data["2"] = domain.Booking{ID: "2", Start: start, End: end, Title: "B"}

	list, err := svc.ListBookings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestService_GetBooking_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetBooking(context.Background(), "nope")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ожидали ErrNotFound, получили %v", err)
	}
}

func TestService_DeleteBooking_Forbidden(t *testing.T) {
	repo := newFakeRepo()
	svc := app.NewService(repo, approom.NewService(&fakeRoomRepo{}))

	start, end := futureInterval()
	repo.data["1"] = domain.Booking{ID: "1", Start: start, End: end, Title: "Чужая", TelegramID: "owner"}

	err := svc.DeleteBooking(context.Background(), "1", "", "not-owner", false)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ожидали ErrForbidden, получили %v", err)
	}
}

func TestService_DeleteBooking_AdminCanDelete(t *testing.T) {
	repo := newFakeRepo()
	svc := app.NewService(repo, approom.NewService(&fakeRoomRepo{}))

	start, end := futureInterval()
	repo.data["1"] = domain.Booking{ID: "1", Start: start, End: end, Title: "Бронь", TelegramID: "owner"}

	if err := svc.DeleteBooking(context.Background(), "1", "admin@edu.hse.ru", "", true); err != nil {
		t.Fatalf("админ должен уметь удалять: %v", err)
	}
	if _, ok := repo.data["1"]; ok {
		t.Fatal("бронь должна быть удалена")
	}
}

func TestService_CreateBooking_TooLong(t *testing.T) {
	svc := newTestService()
	start, _ := futureInterval()

	_, err := svc.CreateBooking(context.Background(), app.CreateBookingInput{
		Start:      start,
		End:        start.Add(4*time.Hour + time.Minute),
		Room:       21,
		Title:      "Слишком длинная",
		TelegramID: "123",
	})
	if !errors.Is(err, domain.ErrTooLongDuration) {
		t.Fatalf("ожидали ErrTooLongDuration, получили %v", err)
	}
}

func TestService_CreateBooking_MaxDurationAllowed(t *testing.T) {
	svc := newTestService()
	start, _ := futureInterval()

	_, err := svc.CreateBooking(context.Background(), app.CreateBookingInput{
		Start:      start,
		End:        start.Add(4 * time.Hour),
		Room:       21,
		Title:      "Ровно 4 часа",
		TelegramID: "123",
	})
	if err != nil {
		t.Fatalf("4 часа должны быть разрешены, получили %v", err)
	}
}

func TestService_CreateBooking_DailyUserLimit(t *testing.T) {
	repo := newFakeRepo()
	svc := app.NewService(repo, approom.NewService(&fakeRoomRepo{}))
	ctx := context.Background()

	start, end := futureInterval()
	repo.data["existing"] = domain.Booking{
		ID:         "existing",
		Start:      start,
		End:        end,
		Room:       21,
		UserEmail:  "user@edu.hse.ru",
		TelegramID: "tg-user",
	}

	_, err := svc.CreateBooking(ctx, app.CreateBookingInput{
		Start:     start.Add(2 * time.Hour),
		End:       start.Add(3 * time.Hour),
		Room:      132,
		Title:     "Вторая бронь",
		UserEmail: "user@edu.hse.ru",
	})
	if !errors.Is(err, domain.ErrDailyUserLimit) {
		t.Fatalf("ожидали ErrDailyUserLimit, получили %v", err)
	}
}

func TestService_CreateBooking_DailyUserLimit_DifferentDays(t *testing.T) {
	repo := newFakeRepo()
	svc := app.NewService(repo, approom.NewService(&fakeRoomRepo{}))
	ctx := context.Background()

	start, end := futureInterval()
	repo.data["existing"] = domain.Booking{
		ID:        "existing",
		Start:     start,
		End:       end,
		Room:      21,
		UserEmail: "user@edu.hse.ru",
	}

	nextDay := start.Add(24 * time.Hour)
	_, err := svc.CreateBooking(ctx, app.CreateBookingInput{
		Start:     nextDay,
		End:       nextDay.Add(time.Hour),
		Room:      132,
		Title:     "Другой день",
		UserEmail: "user@edu.hse.ru",
	})
	if err != nil {
		t.Fatalf("другой день должен быть разрешён, получили %v", err)
	}
}
