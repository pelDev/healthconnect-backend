package sessionstore

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type sessionStore struct {
	mu              sync.RWMutex
	sessions        []models.Session
	sessionMessages map[uuid.UUID][]models.Message
	ttl             map[uuid.UUID]time.Time
}

func NewInMemSessionStore() repositories.ChatStorage {
	return &sessionStore{
		sessions:        make([]models.Session, 0),
		sessionMessages: make(map[uuid.UUID][]models.Message),
		ttl:             make(map[uuid.UUID]time.Time),
	}
}

const sessionTTL = 30 * time.Minute

func (s *sessionStore) CreateSession(sessionID uuid.UUID, vid uuid.UUID, ref *string) (*models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Expire stale sessions
	if exp, ok := s.ttl[sessionID]; ok && time.Now().After(exp) {
		delete(s.sessionMessages, sessionID)
		delete(s.ttl, sessionID)
		// Remove from sessions slice
		for i, session := range s.sessions {
			if session.ID == sessionID {
				s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
				break
			}
		}
	}

	s.ttl[sessionID] = time.Now().Add(sessionTTL)

	// Check if session exists in messages map
	if _, ok := s.sessionMessages[sessionID]; ok {
		// Find and return existing session
		for i := range s.sessions {
			if s.sessions[i].ID == sessionID {
				return &s.sessions[i], nil
			}
		}
	}

	// Create new session
	newSession := models.Session{
		ID:        sessionID,
		VID:       vid,
		Reference: ref,
		EndedAt:   nil,
		CreatedAt: time.Now().UTC(),
	}
	s.sessions = append(s.sessions, newSession)
	s.sessionMessages[sessionID] = []models.Message{}

	return &newSession, nil
}

func (s *sessionStore) GetSession(sessionID uuid.UUID) (*models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ttl[sessionID] = time.Now().Add(sessionTTL)

	// Check if session exists in messages map
	if _, ok := s.sessionMessages[sessionID]; ok {
		// Find and return existing session
		for i := range s.sessions {
			if s.sessions[i].ID == sessionID {
				return &s.sessions[i], nil
			}
		}
	}

	return nil, nil
}

// GetMessagesBySession implements repositories.ChatStorage.
func (s *sessionStore) GetMessagesBySession(sessionID uuid.UUID) ([]models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages, ok := s.sessionMessages[sessionID]
	if !ok {
		return []models.Message{}, nil
	}
	return messages, nil
}

func (s *sessionStore) Append(sessionID uuid.UUID, msgs ...models.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize session messages if it doesn't exist
	if _, ok := s.sessionMessages[sessionID]; !ok {
		s.sessionMessages[sessionID] = []models.Message{}
	}

	s.sessionMessages[sessionID] = append(s.sessionMessages[sessionID], msgs...)
	s.ttl[sessionID] = time.Now().Add(sessionTTL) // Refresh TTL on activity
	return nil
}
