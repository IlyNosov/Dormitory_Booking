package apptglink

import (
	"crypto/rand"
	"encoding/base32"
	"strings"

	"Dormitory_Booking/internal/domain/tglink"
)

type Service struct {
	repo tglink.Repository
}

func NewService(repo tglink.Repository) *Service {
	return &Service{repo: repo}
}

// GetOrCreate возвращает текущую привязку для сессии.
// Если записи нет — создаёт пустую.
func (s *Service) GetOrCreate(sessionID string) (tglink.Link, error) {
	l, err := s.repo.GetBySession(sessionID)
	if err == nil {
		return l, nil
	}
	if err != tglink.ErrNotFound {
		return tglink.Link{}, err
	}
	empty := tglink.Link{SessionID: sessionID}
	if err := s.repo.Upsert(empty); err != nil {
		return tglink.Link{}, err
	}
	return empty, nil
}

// GenerateToken создаёт новый токен привязки для сессии и возвращает его.
func (s *Service) GenerateToken(sessionID string) (string, error) {
	token := newToken()
	l := tglink.Link{SessionID: sessionID, Token: token, Confirmed: false}
	if err := s.repo.Upsert(l); err != nil {
		return "", err
	}
	return token, nil
}

// ConfirmLink вызывается ботом: подтверждает токен и сохраняет telegramID.
func (s *Service) ConfirmLink(token string, telegramID string) error {
	return s.repo.Confirm(token, telegramID)
}

// GetLinkedTelegramID возвращает Telegram ID для сессии, если привязка подтверждена.
func (s *Service) GetLinkedTelegramID(sessionID string) (string, bool) {
	l, err := s.repo.GetBySession(sessionID)
	if err != nil || !l.Confirmed {
		return "", false
	}
	return l.TelegramID, true
}

// Unlink сбрасывает привязку для сессии.
func (s *Service) Unlink(sessionID string) error {
	return s.repo.Unlink(sessionID)
}

func newToken() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	// 8 символов base32 без padding, только буквы+цифры
	return strings.TrimRight(base32.StdEncoding.EncodeToString(b), "=")
}
