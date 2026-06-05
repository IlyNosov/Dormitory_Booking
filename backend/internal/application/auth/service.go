package appauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"

	"Dormitory_Booking/internal/domain/auth"
)

// Mailer отправляет OTP-коды пользователям.
type Mailer interface {
	SendOTP(email, code string) error
}

type Service struct {
	users    auth.UserRepository
	sessions auth.SessionRepository
	otps     auth.OTPRepository
	mailer   Mailer

	emailDomain     string
	adminEmails     map[string]bool
	superAdminEmail string
}

func NewService(
	users auth.UserRepository,
	sessions auth.SessionRepository,
	otps auth.OTPRepository,
	mailer Mailer,
) *Service {
	domain := os.Getenv("EMAIL_DOMAIN")
	if domain == "" {
		domain = "edu.hse.ru"
	}

	adminEmails := make(map[string]bool)
	for _, e := range strings.Split(os.Getenv("ADMIN_EMAILS"), ",") {
		e = strings.TrimSpace(strings.ToLower(e))
		if e != "" {
			adminEmails[e] = true
		}
	}

	return &Service{
		users:           users,
		sessions:        sessions,
		otps:            otps,
		mailer:          mailer,
		emailDomain:     domain,
		adminEmails:     adminEmails,
		superAdminEmail: strings.ToLower(strings.TrimSpace(os.Getenv("SUPER_ADMIN_EMAIL"))),
	}
}

var emailRe = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func (s *Service) validateEmail(email string) bool {
	if !emailRe.MatchString(email) {
		return false
	}
	at := strings.LastIndex(email, "@")
	return strings.EqualFold(email[at+1:], s.emailDomain)
}

// RequestOTP валидирует email, проверяет rate-limit и отправляет OTP.
func (s *Service) RequestOTP(ctx context.Context, email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !s.validateEmail(email) {
		return auth.ErrInvalidEmail
	}

	count, err := s.otps.CountSince(ctx, email, time.Now().Add(-1*time.Hour))
	if err != nil {
		return err
	}
	if count >= 5 {
		return auth.ErrTooManyRequests
	}

	code := generateOTP()
	otp := auth.OTPCode{
		ID:        generateToken(),
		Email:     email,
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := s.otps.CreateOTP(ctx, otp); err != nil {
		return err
	}
	return s.mailer.SendOTP(email, code)
}

// VerifyOTP проверяет код. telegramID опционален (для бота).
// Возвращает Session и User; если telegramID задан — сессия не нужна (бот-режим),
// но Session всё равно создаётся и может быть использована.
func (s *Service) VerifyOTP(ctx context.Context, email, code, telegramID string) (auth.Session, auth.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	otp, err := s.otps.GetLatestUnused(ctx, email)
	if err != nil {
		return auth.Session{}, auth.User{}, auth.ErrInvalidOTP
	}
	if time.Now().After(otp.ExpiresAt) || strings.TrimSpace(code) != otp.Code {
		return auth.Session{}, auth.User{}, auth.ErrInvalidOTP
	}
	if err := s.otps.MarkUsed(ctx, otp.ID); err != nil {
		return auth.Session{}, auth.User{}, err
	}

	isAdmin := s.adminEmails[email]
	isSuperAdmin := email != "" && email == s.superAdminEmail
	if isSuperAdmin {
		isAdmin = true
	}

	user, err := s.users.Upsert(ctx, email, telegramID, isAdmin, isSuperAdmin)
	if err != nil {
		return auth.Session{}, auth.User{}, err
	}

	sess := auth.Session{
		Token:     generateToken(),
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.sessions.CreateSession(ctx, sess); err != nil {
		return auth.Session{}, auth.User{}, err
	}

	return sess, user, nil
}

// GetSession возвращает сессию и пользователя по токену.
func (s *Service) GetSession(ctx context.Context, token string) (auth.Session, auth.User, error) {
	sess, err := s.sessions.Get(ctx, token)
	if err != nil {
		return auth.Session{}, auth.User{}, auth.ErrSessionNotFound
	}
	if time.Now().After(sess.ExpiresAt) {
		_ = s.sessions.Delete(ctx, token)
		return auth.Session{}, auth.User{}, auth.ErrSessionNotFound
	}
	user, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return auth.Session{}, auth.User{}, err
	}
	return sess, user, nil
}

// Logout удаляет сессию.
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.sessions.Delete(ctx, token)
}

func generateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1_000_000))
	return fmt.Sprintf("%06d", n.Int64())
}

func generateToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
