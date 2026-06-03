package tglinkstore

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"Dormitory_Booking/internal/domain/tglink"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Upsert(l tglink.Link) error {
	_, err := s.pool.Exec(context.Background(),
		`INSERT INTO telegram_links (session_id, telegram_id, token, confirmed)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (session_id) DO UPDATE
		   SET telegram_id = EXCLUDED.telegram_id,
		       token       = EXCLUDED.token,
		       confirmed   = EXCLUDED.confirmed`,
		l.SessionID, l.TelegramID, l.Token, l.Confirmed,
	)
	return err
}

func (s *PostgresStore) GetBySession(sessionID string) (tglink.Link, error) {
	row := s.pool.QueryRow(context.Background(),
		`SELECT session_id, telegram_id, token, confirmed FROM telegram_links WHERE session_id=$1`, sessionID)
	var l tglink.Link
	if err := row.Scan(&l.SessionID, &l.TelegramID, &l.Token, &l.Confirmed); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tglink.Link{}, tglink.ErrNotFound
		}
		return tglink.Link{}, err
	}
	return l, nil
}

func (s *PostgresStore) GetByToken(token string) (tglink.Link, error) {
	row := s.pool.QueryRow(context.Background(),
		`SELECT session_id, telegram_id, token, confirmed FROM telegram_links WHERE token=$1 AND confirmed=false`, token)
	var l tglink.Link
	if err := row.Scan(&l.SessionID, &l.TelegramID, &l.Token, &l.Confirmed); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tglink.Link{}, tglink.ErrTokenExpired
		}
		return tglink.Link{}, err
	}
	return l, nil
}

func (s *PostgresStore) Confirm(token string, telegramID string) error {
	ct, err := s.pool.Exec(context.Background(),
		`UPDATE telegram_links SET telegram_id=$1, confirmed=true, token='' WHERE token=$2 AND confirmed=false`,
		telegramID, token,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return tglink.ErrTokenExpired
	}
	return nil
}

func (s *PostgresStore) Unlink(sessionID string) error {
	_, err := s.pool.Exec(context.Background(),
		`DELETE FROM telegram_links WHERE session_id=$1`, sessionID)
	return err
}
