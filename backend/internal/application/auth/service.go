package auth

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"Dormitory_Booking/internal/domain/user"
	"Dormitory_Booking/internal/infrastructure/mailer"
	"Dormitory_Booking/internal/infrastructure/postgres"
)

var hseEmailRe = regexp.MustCompile(`(?i)^[a-zA-Z0-9._%+\-]+@(edu\.hse\.ru|hse\.ru)$`)

type Service struct {
	users    *postgres.UserRepo
	sessions *postgres.SessionRepo
	otps     *postgres.OTPRepo
	mailer   *mailer.Mailer
}

func NewService(
	users *postgres.UserRepo,
	sessions *postgres.SessionRepo,
	otps *postgres.OTPRepo,
	m *mailer.Mailer,
) *Service {
	return &Service{users: users, sessions: sessions, otps: otps, mailer: m}
}

func ValidateHSEEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if !hseEmailRe.MatchString(email) {
		return errors.New("принимаются только адреса @hse.ru и @edu.hse.ru")
	}
	return nil
}

// RequestOTP отправляет 6-значный код на почту, не более 5 раз за 10 мин
func (s *Service) RequestOTP(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if err := ValidateHSEEmail(email); err != nil {
		return err
	}

	count, err := s.otps.CountRecent(ctx, email)
	if err != nil {
		return err
	}
	if count >= 5 {
		return errors.New("слишком много попыток, подождите 10 минут")
	}

	code, err := postgres.GenerateOTPCode()
	if err != nil {
		return err
	}

	if err := s.otps.Create(ctx, email, code); err != nil {
		return err
	}

	return s.mailer.SendOTP(email, code)
}

// VerifyOTP проверяет код, создаёт сессию, возвращает токен
func (s *Service) VerifyOTP(ctx context.Context, email, code string) (string, user.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	ok, err := s.otps.Verify(ctx, email, code)
	if err != nil {
		return "", user.User{}, err
	}
	if !ok {
		return "", user.User{}, errors.New("неверный или истёкший код")
	}

	u, err := s.users.Upsert(ctx, email)
	if err != nil {
		return "", user.User{}, err
	}

	token, err := postgres.GenerateSessionToken()
	if err != nil {
		return "", user.User{}, err
	}

	if err := s.sessions.Create(ctx, token, u.ID); err != nil {
		return "", user.User{}, err
	}

	return token, u, nil
}

// VerifyOTPOnly проверяет код без создания веб-сессии, используется Telegram-ботом
func (s *Service) VerifyOTPOnly(ctx context.Context, email, code string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	ok, err := s.otps.Verify(ctx, email, code)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("неверный или истёкший код")
	}
	return nil
}

// GetSession возвращает пользователя по токену сессии
func (s *Service) GetSession(ctx context.Context, token string) (user.User, error) {
	return s.sessions.Get(ctx, token)
}

// Logout инвалидирует токен сессии
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.sessions.Delete(ctx, token)
}
