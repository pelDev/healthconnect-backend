package sessionstore

import (
	"sync"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type sessionStore struct {
	mu              sync.RWMutex
	sessionMessages map[uuid.UUID][]models.Message
}

func NewInMemSessionStore() repositories.MessageStorage {
	return &sessionStore{
		sessionMessages: make(map[uuid.UUID][]models.Message),
	}
}

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
	return nil
}
