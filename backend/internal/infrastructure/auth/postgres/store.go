package authpostgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"Dormitory_Booking/internal/domain/auth"
)

// Store реализует UserRepository, SessionRepository и OTPRepository поверх Postgres.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ─── UserRepository ───────────────────────────────────────────────────────────

func (s *Store) Upsert(ctx context.Context, email, telegramID string, isAdmin, isSuperAdmin bool) (auth.User, error) {
	row := s.pool.QueryRow(ctx,
		`INSERT INTO users (id, email, telegram_id, is_admin, is_super_admin)
		 VALUES (gen_random_uuid()::text, $1, $2, $3, $4)
		 ON CONFLICT (email) DO UPDATE
		   SET telegram_id    = CASE WHEN EXCLUDED.telegram_id != '' THEN EXCLUDED.telegram_id ELSE users.telegram_id END,
		       is_admin       = EXCLUDED.is_admin,
		       is_super_admin = EXCLUDED.is_super_admin
		 RETURNING id, email, telegram_id, is_admin, is_super_admin, created_at`,
		email, telegramID, isAdmin, isSuperAdmin,
	)
	var u auth.User
	if err := row.Scan(&u.ID, &u.Email, &u.TelegramID, &u.IsAdmin, &u.IsSuperAdmin, &u.CreatedAt); err != nil {
		return auth.User{}, err
	}
	return u, nil
}

func (s *Store) GetByEmail(ctx context.Context, email string) (auth.User, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, email, telegram_id, is_admin, is_super_admin, created_at FROM users WHERE email=$1`, email)
	var u auth.User
	if err := row.Scan(&u.ID, &u.Email, &u.TelegramID, &u.IsAdmin, &u.IsSuperAdmin, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, auth.ErrUserNotFound
		}
		return auth.User{}, err
	}
	return u, nil
}

func (s *Store) GetByID(ctx context.Context, id string) (auth.User, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, email, telegram_id, is_admin, is_super_admin, created_at FROM users WHERE id=$1`, id)
	var u auth.User
	if err := row.Scan(&u.ID, &u.Email, &u.TelegramID, &u.IsAdmin, &u.IsSuperAdmin, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.User{}, auth.ErrUserNotFound
		}
		return auth.User{}, err
	}
	return u, nil
}

// ─── SessionRepository ────────────────────────────────────────────────────────

func (s *Store) CreateSession(ctx context.Context, sess auth.Session) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		sess.Token, sess.UserID, sess.ExpiresAt,
	)
	return err
}

func (s *Store) Get(ctx context.Context, token string) (auth.Session, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT token, user_id, expires_at FROM sessions WHERE token=$1`, token)
	var sess auth.Session
	if err := row.Scan(&sess.Token, &sess.UserID, &sess.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.Session{}, auth.ErrSessionNotFound
		}
		return auth.Session{}, err
	}
	return sess, nil
}

func (s *Store) Delete(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token=$1`, token)
	return err
}

// ─── OTPRepository ────────────────────────────────────────────────────────────

func (s *Store) CreateOTP(ctx context.Context, o auth.OTPCode) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO otp_codes (id, email, code, expires_at) VALUES ($1, $2, $3, $4)`,
		o.ID, o.Email, o.Code, o.ExpiresAt,
	)
	return err
}

func (s *Store) GetLatestUnused(ctx context.Context, email string) (auth.OTPCode, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT id, email, code, created_at, expires_at, used
		 FROM otp_codes
		 WHERE email=$1 AND used=false
		 ORDER BY created_at DESC LIMIT 1`, email)
	var o auth.OTPCode
	if err := row.Scan(&o.ID, &o.Email, &o.Code, &o.CreatedAt, &o.ExpiresAt, &o.Used); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.OTPCode{}, auth.ErrInvalidOTP
		}
		return auth.OTPCode{}, err
	}
	return o, nil
}

func (s *Store) CountSince(ctx context.Context, email string, since time.Time) (int, error) {
	row := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM otp_codes WHERE email=$1 AND created_at > $2`, email, since)
	var n int
	return n, row.Scan(&n)
}

func (s *Store) MarkUsed(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE otp_codes SET used=true WHERE id=$1`, id)
	return err
}
