package postgres

import (
	"context"
	"errors"

	"Dormitory_Booking/internal/domain/booking"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingPostgresRepo struct{ pool *pgxpool.Pool }

func NewBookingPostgresRepo(pool *pgxpool.Pool) *BookingPostgresRepo {
	return &BookingPostgresRepo{pool: pool}
}

func (r *BookingPostgresRepo) List(ctx context.Context) ([]booking.Booking, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, start_at, end_at, room, title,
		       COALESCE(description,''), user_email, telegram_id, is_private, tg_msg_id
		FROM bookings ORDER BY start_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []booking.Booking
	for rows.Next() {
		var b booking.Booking
		if err := rows.Scan(&b.ID, &b.Start, &b.End, &b.Room, &b.Title,
			&b.Description, &b.UserEmail, &b.TelegramID, &b.IsPrivate, &b.TgMsgID); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BookingPostgresRepo) Get(ctx context.Context, id string) (booking.Booking, error) {
	var b booking.Booking
	err := r.pool.QueryRow(ctx, `
		SELECT id, start_at, end_at, room, title,
		       COALESCE(description,''), user_email, telegram_id, is_private, tg_msg_id
		FROM bookings WHERE id = $1`, id).
		Scan(&b.ID, &b.Start, &b.End, &b.Room, &b.Title,
			&b.Description, &b.UserEmail, &b.TelegramID, &b.IsPrivate, &b.TgMsgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return booking.Booking{}, booking.ErrNotFound
	}
	return b, err
}

func (r *BookingPostgresRepo) Create(ctx context.Context, b booking.Booking) (booking.Booking, error) {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO bookings (id, start_at, end_at, room, title, description, user_email, telegram_id, is_private)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		b.ID, b.Start, b.End, b.Room, b.Title,
		nullIfEmpty(b.Description), b.UserEmail, b.TelegramID, b.IsPrivate)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return booking.Booking{}, booking.ErrOverlap
		}
		return booking.Booking{}, err
	}
	return b, nil
}

func (r *BookingPostgresRepo) Update(ctx context.Context, b booking.Booking) (booking.Booking, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE bookings
		SET start_at=$1, end_at=$2, room=$3, title=$4, description=$5, telegram_id=$6, is_private=$7
		WHERE id=$8`,
		b.Start, b.End, b.Room, b.Title,
		nullIfEmpty(b.Description), b.TelegramID, b.IsPrivate, b.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23P01" {
			return booking.Booking{}, booking.ErrOverlap
		}
		return booking.Booking{}, err
	}
	if tag.RowsAffected() == 0 {
		return booking.Booking{}, booking.ErrNotFound
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

func (r *BookingPostgresRepo) SetTgMsgID(ctx context.Context, bookingID string, msgID int) error {
	_, err := r.pool.Exec(ctx, `UPDATE bookings SET tg_msg_id=$1 WHERE id=$2`, msgID, bookingID)
	return err
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
