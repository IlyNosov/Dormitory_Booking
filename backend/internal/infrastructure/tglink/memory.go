package tglinkstore

import (
	"sync"

	"Dormitory_Booking/internal/domain/tglink"
)

type MemoryStore struct {
	mu    sync.RWMutex
	bySession map[string]tglink.Link
	byToken   map[string]string // token -> sessionID
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bySession: make(map[string]tglink.Link),
		byToken:   make(map[string]string),
	}
}

func (s *MemoryStore) Upsert(l tglink.Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if old, ok := s.bySession[l.SessionID]; ok && old.Token != "" {
		delete(s.byToken, old.Token)
	}
	s.bySession[l.SessionID] = l
	if l.Token != "" {
		s.byToken[l.Token] = l.SessionID
	}
	return nil
}

func (s *MemoryStore) GetBySession(sessionID string) (tglink.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.bySession[sessionID]
	if !ok {
		return tglink.Link{}, tglink.ErrNotFound
	}
	return l, nil
}

func (s *MemoryStore) GetByToken(token string) (tglink.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sid, ok := s.byToken[token]
	if !ok {
		return tglink.Link{}, tglink.ErrTokenExpired
	}
	l := s.bySession[sid]
	if l.Confirmed {
		return tglink.Link{}, tglink.ErrTokenExpired
	}
	return l, nil
}

func (s *MemoryStore) Confirm(token string, telegramID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sid, ok := s.byToken[token]
	if !ok {
		return tglink.ErrTokenExpired
	}
	l := s.bySession[sid]
	l.TelegramID = telegramID
	l.Confirmed = true
	s.bySession[sid] = l
	delete(s.byToken, token)
	return nil
}

func (s *MemoryStore) Unlink(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l, ok := s.bySession[sessionID]; ok {
		delete(s.byToken, l.Token)
	}
	delete(s.bySession, sessionID)
	return nil
}
