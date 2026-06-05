package authmemory

import (
	"context"
	"errors"
	"sync"
	"time"

	"Dormitory_Booking/internal/domain/auth"
)

// Store реализует UserRepository, SessionRepository и OTPRepository в памяти.
type Store struct {
	mu       sync.RWMutex
	users    map[string]auth.User    // email → User
	sessions map[string]auth.Session // token → Session
	otps     []auth.OTPCode
}

func NewStore() *Store {
	return &Store{
		users:    make(map[string]auth.User),
		sessions: make(map[string]auth.Session),
	}
}

// ─── UserRepository ───────────────────────────────────────────────────────────

func (s *Store) Upsert(_ context.Context, email, telegramID string, isAdmin, isSuperAdmin bool) (auth.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[email]
	if !ok {
		u = auth.User{ID: email, Email: email, CreatedAt: time.Now()}
	}
	if telegramID != "" {
		u.TelegramID = telegramID
	}
	u.IsAdmin = isAdmin
	u.IsSuperAdmin = isSuperAdmin
	s.users[email] = u
	return u, nil
}

func (s *Store) GetByEmail(_ context.Context, email string) (auth.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[email]
	if !ok {
		return auth.User{}, auth.ErrUserNotFound
	}
	return u, nil
}

func (s *Store) GetByID(_ context.Context, id string) (auth.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, u := range s.users {
		if u.ID == id {
			return u, nil
		}
	}
	return auth.User{}, auth.ErrUserNotFound
}

// ─── SessionRepository ────────────────────────────────────────────────────────

func (s *Store) CreateSession(_ context.Context, sess auth.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.Token] = sess
	return nil
}

func (s *Store) Get(_ context.Context, token string) (auth.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[token]
	if !ok {
		return auth.Session{}, auth.ErrSessionNotFound
	}
	return sess, nil
}

func (s *Store) Delete(_ context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
	return nil
}

// ─── OTPRepository ────────────────────────────────────────────────────────────

func (s *Store) CreateOTP(_ context.Context, o auth.OTPCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.otps = append(s.otps, o)
	return nil
}

func (s *Store) GetLatestUnused(_ context.Context, email string) (auth.OTPCode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := len(s.otps) - 1; i >= 0; i-- {
		if s.otps[i].Email == email && !s.otps[i].Used {
			return s.otps[i], nil
		}
	}
	return auth.OTPCode{}, errors.New("not found")
}

func (s *Store) CountSince(_ context.Context, email string, since time.Time) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := 0
	for _, o := range s.otps {
		if o.Email == email && o.CreatedAt.After(since) {
			n++
		}
	}
	return n, nil
}

func (s *Store) MarkUsed(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.otps {
		if s.otps[i].ID == id {
			s.otps[i].Used = true
			return nil
		}
	}
	return errors.New("not found")
}
