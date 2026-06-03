package booking_test

import (
	"context"
	"errors"
	"testing"
	"time"

	app "Dormitory_Booking/internal/application/booking"
	domain "Dormitory_Booking/internal/domain/booking"
)

type fakeRepo struct {
	data map[string]domain.Booking
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		data: make(map[string]domain.Booking),
	}
}

func (r *fakeRepo) List(ctx context.Context) ([]domain.Booking, error) {
	res := make([]domain.Booking, 0, len(r.data))
	for _, b := range r.data {
		res = append(res, b)
	}
	return res, nil
}

func (r *fakeRepo) Get(ctx context.Context, id string) (domain.Booking, error) {
	b, ok := r.data[id]
	if !ok {
		return domain.Booking{}, domain.ErrNotFound
	}
	return b, nil
}

func (r *fakeRepo) Create(ctx context.Context, b domain.Booking) (domain.Booking, error) {
	if b.ID == "" {
		b.ID = "id-" + b.Start.Format("150405")
	}
	r.data[b.ID] = b
	return b, nil
}

func (r *fakeRepo) Delete(ctx context.Context, id string) error {
	if _, ok := r.data[id]; !ok {
		return domain.ErrNotFound
	}
	delete(r.data, id)
	return nil
}

func futureInterval() (time.Time, time.Time) {
	loc := time.Local
	now := time.Now().In(loc)

	day := now.Add(24 * time.Hour)
	start := time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, loc)

	if !start.After(now) {
		day = day.Add(24 * time.Hour)
		start = time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, loc)
	}

	end := start.Add(1 * time.Hour)
	return start, end
}

func TestService_CreateBooking_OK(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := app.NewService(repo)

	start, end := futureInterval()

	input := app.CreateBookingInput{
		Start:       start,
		End:         end,
		Room:        domain.Room21,
		Title:       "Лаба по матану",
		Description: "ЗВУУУУУУУУУУУУУУУУУУУУУУУК",
		TelegramID:  "123",
		IsPrivate:   false,
	}

	created, err := svc.CreateBooking(ctx, input)
	if err != nil {
		t.Fatalf("ожидали nil, получили %v", err)
	}

	if created.ID == "" {
		t.Fatalf("ожидали, что ID будет установлен")
	}
	if created.Description != input.Description {
		t.Fatalf("описание не сохранилось: got=%q want=%q", created.Description, input.Description)
	}
}

func TestService_CreateBooking_Overlap(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := app.NewService(repo)

	start, end := futureInterval()

	// уже есть бронь 10:00–11:00 в той же аудитории
	existing := domain.Booking{
		ID:         "1",
		Start:      start,
		End:        end,
		Room:       domain.Room21,
		Title:      "Семинар",
		TelegramID: "111",
	}
	repo.data[existing.ID] = existing
	input := app.CreateBookingInput{
		Start:      start.Add(30 * time.Minute),
		End:        end.Add(30 * time.Minute),
		Room:       domain.Room21,
		Title:      "Пересекающаяся бронь",
		TelegramID: "222",
	}

	_, err := svc.CreateBooking(ctx, input)
	if !errors.Is(err, domain.ErrOverlap) {
		t.Fatalf("ожидали ErrOverlap, получили %v", err)
	}
}

func TestService_ListBookings(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := app.NewService(repo)

	start, end := futureInterval()
	repo.data["1"] = domain.Booking{ID: "1", Start: start, End: end, Title: "A"}
	repo.data["2"] = domain.Booking{ID: "2", Start: start, End: end, Title: "B"}

	list, err := svc.ListBookings(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestService_GetBooking_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := app.NewService(repo)

	_, err := svc.GetBooking(ctx, "nope")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ожидали ErrNotFound, получили %v", err)
	}
}

func TestService_DeleteBooking_Forbidden(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := app.NewService(repo)

	start, end := futureInterval()
	repo.data["1"] = domain.Booking{
		ID:         "1",
		Start:      start,
		End:        end,
		Title:      "Чужая бронь",
		TelegramID: "owner",
	}

	err := svc.DeleteBooking(ctx, "1", "not-owner", false)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("ожидали ErrForbidden, получили %v", err)
	}
}

func TestService_DeleteBooking_AdminCanDelete(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := app.NewService(repo)

	start, end := futureInterval()
	repo.data["1"] = domain.Booking{
		ID:         "1",
		Start:      start,
		End:        end,
		Title:      "Бронь",
		TelegramID: "owner",
	}

	err := svc.DeleteBooking(ctx, "1", "some-admin", true)
	if err != nil {
		t.Fatalf("админ должен уметь удалять, err=%v", err)
	}
	if _, ok := repo.data["1"]; ok {
		t.Fatalf("бронь должна быть удалена из репозитория")
	}
}
