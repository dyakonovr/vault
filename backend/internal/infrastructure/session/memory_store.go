package session

import (
	"context"
	"sync"
	"time"
	"vault/internal/domain"
	"vault/pkg/crypto"
)

type MemoryStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

type Session struct {
	UserID    int64
	ExpiresAt time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]Session),
	}
}

func (s *MemoryStore) Create(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	id, err := crypto.GenerateToken()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = Session{
		UserID:    userID,
		ExpiresAt: time.Now().Add(ttl),
	}
	return id, nil
}

func (s *MemoryStore) Delete(ctx context.Context, sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

func (s *MemoryStore) GetUserID(ctx context.Context, sessionID string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok || time.Now().After(sess.ExpiresAt) {
		return 0, domain.ErrSessionNotFound // доменная ошибка
	}
	return sess.UserID, nil
}