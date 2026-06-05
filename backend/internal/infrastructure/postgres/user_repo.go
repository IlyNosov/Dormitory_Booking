package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"Dormitory_Booking/internal/domain/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct{ pool *pgxpool.Pool }

func NewUserRepo(pool *pgxpool.Pool) *UserRepo { return &UserRepo{pool: pool} }

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (user.User, error) {
	var u user.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, COALESCE(telegram_id,''), created_at FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.Email, &u.TelegramID, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, errors.New("пользователь не найден")
	}
	return u, err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (user.User, error) {
	var u user.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, COALESCE(telegram_id,''), created_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.TelegramID, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, errors.New("пользователь не найден")
	}
	return u, err
}

func (r *UserRepo) Upsert(ctx context.Context, email string) (user.User, error) {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (email) VALUES ($1) ON CONFLICT (email) DO NOTHING`, email)
	if err != nil {
		return user.User{}, err
	}
	return r.GetByEmail(ctx, email)
}

// SessionRepo

type SessionRepo struct{ pool *pgxpool.Pool }

func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo { return &SessionRepo{pool: pool} }

func (r *SessionRepo) Create(ctx context.Context, token, userID string) error {
	expires := time.Now().Add(7 * 24 * time.Hour)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES ($1,$2,$3)`,
		token, userID, expires)
	return err
}

func (r *SessionRepo) Get(ctx context.Context, token string) (user.User, error) {
	var u user.User
	err := r.pool.QueryRow(ctx, `
		SELECT u.id, u.email, COALESCE(u.telegram_id,''), u.created_at
		FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > NOW()`, token).
		Scan(&u.ID, &u.Email, &u.TelegramID, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, errors.New("сессия не найдена или истекла")
	}
	return u, err
}

func (r *SessionRepo) Delete(ctx context.Context, token string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

// OTPRepo

type OTPRepo struct{ pool *pgxpool.Pool }

func NewOTPRepo(pool *pgxpool.Pool) *OTPRepo { return &OTPRepo{pool: pool} }

func (r *OTPRepo) Create(ctx context.Context, email, code string) error {
	expires := time.Now().Add(10 * time.Minute)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO otp_codes (email, code, expires_at) VALUES ($1,$2,$3)`,
		email, code, expires)
	return err
}

func (r *OTPRepo) Verify(ctx context.Context, email, code string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE otp_codes SET used = true
		WHERE email = $1 AND code = $2 AND expires_at > NOW() AND NOT used
		  AND id = (
		    SELECT id FROM otp_codes
		    WHERE email = $1 AND code = $2 AND expires_at > NOW() AND NOT used
		    ORDER BY created_at DESC LIMIT 1
		  )`, email, code)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *OTPRepo) CountRecent(ctx context.Context, email string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM otp_codes WHERE email = $1 AND created_at > NOW() - INTERVAL '10 minutes'`,
		email).Scan(&count)
	return count, err
}

// AdminRepo

type AdminRepo struct{ pool *pgxpool.Pool }

func NewAdminRepo(pool *pgxpool.Pool) *AdminRepo { return &AdminRepo{pool: pool} }

func (r *AdminRepo) List(ctx context.Context) ([]adminEntry, error) {
	rows, err := r.pool.Query(ctx, `SELECT email, added_by, created_at FROM admins ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []adminEntry
	for rows.Next() {
		var a adminEntry
		if err := rows.Scan(&a.Email, &a.AddedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AdminRepo) IsAdmin(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM admins WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}

func (r *AdminRepo) Add(ctx context.Context, email, addedBy string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO admins (email, added_by) VALUES ($1,$2) ON CONFLICT DO NOTHING`, email, addedBy)
	return err
}

func (r *AdminRepo) Remove(ctx context.Context, email string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM admins WHERE email = $1`, email)
	return err
}

type adminEntry struct {
	Email     string
	AddedBy   string
	CreatedAt time.Time
}

// Helpers

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GenerateOTPCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
