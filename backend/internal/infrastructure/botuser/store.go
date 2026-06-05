package botuser

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store saves and loads the telegram_chat_id → email mapping for bot-authenticated users.
type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Save(ctx context.Context, chatID int64, email string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO bot_users (telegram_chat_id, email)
		 VALUES ($1, $2)
		 ON CONFLICT (telegram_chat_id) DO UPDATE SET email = EXCLUDED.email`,
		chatID, email)
	return err
}

func (s *Store) LoadAll(ctx context.Context) (map[int64]string, error) {
	rows, err := s.pool.Query(ctx, `SELECT telegram_chat_id, email FROM bot_users`)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return map[int64]string{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	m := make(map[int64]string)
	for rows.Next() {
		var id int64
		var email string
		if err := rows.Scan(&id, &email); err == nil {
			m[id] = email
		}
	}
	return m, rows.Err()
}
