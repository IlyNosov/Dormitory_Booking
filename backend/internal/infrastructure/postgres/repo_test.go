package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"Dormitory_Booking/internal/domain/booking"
	pgrepo "Dormitory_Booking/internal/infrastructure/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

func requireTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DB_URL")
	if url == "" {
		t.Skip("TEST_DB_URL не установлен, пропуск тестов Postgres репозитория")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Skipf("не удалось подключиться к тестовой БД (%s): %v", url, err)
	}
	if _, err := pool.Exec(ctx, `SELECT 1`); err != nil {
		t.Skipf("тестовая БД не готова, пропуск тестов Postgres репозитория: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM bookings`); err != nil {
		t.Skipf("не удалось очистить таблицу bookings, пропуск тестов Postgres репозитория: %v", err)
	}

	return pool
}

func TestPostgresRepo_CreateAndGet(t *testing.T) {
	pool := requireTestDB(t)
	repo := pgrepo.NewBookingPostgresRepo(pool)
	ctx := context.Background()

	b := booking.Booking{
		Start:      time.Now(),
		End:        time.Now().Add(time.Hour),
		Room:       booking.Room21,
		Title:      "PG Test",
		TelegramID: "111",
		IsPrivate:  false,
	}

	created, err := repo.Create(ctx, b)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if got.Title != b.Title {
		t.Fatalf("ожидалось %s, получено %s", b.Title, got.Title)
	}
}

func TestPostgresRepo_List(t *testing.T) {
	pool := requireTestDB(t)
	repo := pgrepo.NewBookingPostgresRepo(pool)
	ctx := context.Background()

	repo.Create(ctx, booking.Booking{
		Start:      time.Now(),
		End:        time.Now().Add(time.Hour),
		Room:       booking.Room21,
		Title:      "A",
		TelegramID: "123",
	})

	repo.Create(ctx, booking.Booking{
		Start:      time.Now().Add(2 * time.Hour),
		End:        time.Now().Add(3 * time.Hour),
		Room:       booking.Room21,
		Title:      "B",
		TelegramID: "123",
	})

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("ожидалось 2, получено %d", len(list))
	}
}

func TestPostgresRepo_Delete(t *testing.T) {
	pool := requireTestDB(t)
	repo := pgrepo.NewBookingPostgresRepo(pool)
	ctx := context.Background()

	created, _ := repo.Create(ctx, booking.Booking{
		Start:      time.Now(),
		End:        time.Now().Add(time.Hour),
		Room:       booking.Room21,
		Title:      "To delete",
		TelegramID: "222",
	})

	err := repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}

	_, err = repo.Get(ctx, created.ID)
	if err != booking.ErrNotFound {
		t.Fatalf("ожидалось ErrNotFound, получено %v", err)
	}
}

func TestPostgresRepo_OverlapConstraint(t *testing.T) {
	pool := requireTestDB(t)
	repo := pgrepo.NewBookingPostgresRepo(pool)
	ctx := context.Background()

	start := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	_, err := repo.Create(ctx, booking.Booking{
		Start:      start,
		End:        start.Add(1 * time.Hour),
		Room:       booking.Room21,
		Title:      "Base",
		TelegramID: "333",
	})
	if err != nil {
		t.Fatalf("неожиданная ошибка: %v", err)
	}
	_, err = repo.Create(ctx, booking.Booking{
		Start:      start.Add(30 * time.Minute),
		End:        start.Add(90 * time.Minute),
		Room:       booking.Room21,
		Title:      "Overlap",
		TelegramID: "333",
	})

	if !errors.Is(err, booking.ErrOverlap) {
		t.Fatalf("ожидалось ErrOverlap, got %v", err)
	}
}
