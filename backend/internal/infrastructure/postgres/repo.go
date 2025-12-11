package postgres

// В этом файле лежит реализация репозитория бронирований через Postgres.
// По сути это адаптер между доменной моделью и таблицей bookings в БД.

import (
	"context"
	"errors"

	"Dormitory_Booking/internal/domain/booking"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingPostgresRepo struct {
	pool *pgxpool.Pool
}

// NewBookingPostgresRepo создаёт репозиторий поверх пула соединений pgx.
func NewBookingPostgresRepo(pool *pgxpool.Pool) *BookingPostgresRepo {
	return &BookingPostgresRepo{pool: pool}
}

func (r *BookingPostgresRepo) List(ctx context.Context) ([]booking.Booking, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, start_at, end_at, room, title, COALESCE(description, ''), telegram_id, is_private
		 FROM bookings
		 ORDER BY start_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []booking.Booking
	for rows.Next() {
		var b booking.Booking
		if err := rows.Scan(
			&b.ID,
			&b.Start,
			&b.End,
			&b.Room,
			&b.Title,
			&b.Description,
			&b.TelegramID,
			&b.IsPrivate,
		); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BookingPostgresRepo) Get(ctx context.Context, id string) (booking.Booking, error) {
	var b booking.Booking
	err := r.pool.QueryRow(ctx,
		`SELECT id, start_at, end_at, room, title, COALESCE(description, ''), telegram_id, is_private
		 FROM bookings
		 WHERE id = $1`,
		id,
	).Scan(
		&b.ID,
		&b.Start,
		&b.End,
		&b.Room,
		&b.Title,
		&b.Description,
		&b.TelegramID,
		&b.IsPrivate,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return booking.Booking{}, booking.ErrNotFound
		}
		return booking.Booking{}, err
	}
	return b, nil
}

func (r *BookingPostgresRepo) Create(ctx context.Context, b booking.Booking) (booking.Booking, error) {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO bookings (id, start_at, end_at, room, title, description, telegram_id, is_private)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		b.ID,
		b.Start,
		b.End,
		int(b.Room),
		b.Title,
		nullIfEmpty(b.Description),
		b.TelegramID,
		b.IsPrivate,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23P01" {
				return booking.Booking{}, booking.ErrOverlap
			}
		}
		return booking.Booking{}, err
	}

	return b, nil
}

func (r *BookingPostgresRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM bookings WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return booking.ErrNotFound
	}
	return nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
