package model

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID    uuid.UUID
	Email     string
	ExpiresAt time.Time
}

type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var sessionStore = &SessionStore{
	sessions: make(map[string]*Session),
}

func GetSessionStore() *SessionStore {
	return sessionStore
}

// Maybe create a session on db so dont have to rely on in-memory store
// later
func (s *SessionStore) Create(userID uuid.UUID, email string, duration time.Duration) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := uuid.New().String()
	s.sessions[sessionID] = &Session{
		UserID:    userID,
		Email:     email,
		ExpiresAt: time.Now().Add(duration),
	}
	return sessionID
}

func (s *SessionStore) Get(sessionID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	return session, true
}

func (s *SessionStore) Delete(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}
