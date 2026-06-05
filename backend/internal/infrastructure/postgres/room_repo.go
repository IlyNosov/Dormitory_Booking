package postgres

import (
	"context"
	"errors"

	"Dormitory_Booking/internal/domain/room"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepo struct{ pool *pgxpool.Pool }

func NewRoomRepo(pool *pgxpool.Pool) *RoomRepo { return &RoomRepo{pool: pool} }

func (r *RoomRepo) List(ctx context.Context) ([]room.Room, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT number, name, is_active,
		       weekday_open, weekday_close, frisat_open, frisat_close, sun_open, sun_close, created_at
		FROM rooms ORDER BY number`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []room.Room
	for rows.Next() {
		var rm room.Room
		if err := rows.Scan(&rm.Number, &rm.Name, &rm.IsActive,
			&rm.WeekdayOpen, &rm.WeekdayClose, &rm.FriSatOpen, &rm.FriSatClose,
			&rm.SunOpen, &rm.SunClose, &rm.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rm)
	}
	return out, rows.Err()
}

func (r *RoomRepo) Get(ctx context.Context, number int) (room.Room, error) {
	var rm room.Room
	err := r.pool.QueryRow(ctx, `
		SELECT number, name, is_active,
		       weekday_open, weekday_close, frisat_open, frisat_close, sun_open, sun_close, created_at
		FROM rooms WHERE number = $1`, number).
		Scan(&rm.Number, &rm.Name, &rm.IsActive,
			&rm.WeekdayOpen, &rm.WeekdayClose, &rm.FriSatOpen, &rm.FriSatClose,
			&rm.SunOpen, &rm.SunClose, &rm.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return room.Room{}, errors.New("комната не найдена")
	}
	return rm, err
}

func (r *RoomRepo) Create(ctx context.Context, rm room.Room) (room.Room, error) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO rooms (number, name, is_active,
		    weekday_open, weekday_close, frisat_open, frisat_close, sun_open, sun_close)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		rm.Number, rm.Name, rm.IsActive,
		rm.WeekdayOpen, rm.WeekdayClose, rm.FriSatOpen, rm.FriSatClose,
		rm.SunOpen, rm.SunClose)
	if err != nil {
		return room.Room{}, err
	}
	return r.Get(ctx, rm.Number)
}

func (r *RoomRepo) Delete(ctx context.Context, number int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM rooms WHERE number = $1`, number)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("комната не найдена")
	}
	return nil
}
