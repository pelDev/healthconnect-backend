package sessionstore

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/ports"
)

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID][]models.Message
	ttl      map[uuid.UUID]time.Time
}

func NewInMemSessionStore() ports.SessionStorage {
	return &sessionStore{
		sessions: make(map[uuid.UUID][]models.Message),
		ttl:      make(map[uuid.UUID]time.Time),
	}
}

const sessionTTL = 30 * time.Minute

func (s *sessionStore) GetOrCreateSession(sessionID uuid.UUID) []models.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Expire stale sessions
	if exp, ok := s.ttl[sessionID]; ok && time.Now().After(exp) {
		delete(s.sessions, sessionID)
		delete(s.ttl, sessionID)
	}

	s.ttl[sessionID] = time.Now().Add(sessionTTL)

	if history, ok := s.sessions[sessionID]; ok {
		return history
	}

	s.sessions[sessionID] = []models.Message{}
	return s.sessions[sessionID]
}

func (s *sessionStore) Append(sessionID uuid.UUID, msgs ...models.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = append(s.sessions[sessionID], msgs...)
	s.ttl[sessionID] = time.Now().Add(sessionTTL) // Refresh TTL on activity
	return nil
}
