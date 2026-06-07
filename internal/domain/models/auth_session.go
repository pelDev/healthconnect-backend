package models

import (
	"time"

	"github.com/google/uuid"
)

const LoginSessionDurationMinutes = 60 * 24 // 1 day

type AuthSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
	IsRevoked bool
}

func (s *AuthSession) Revoke() {
	s.IsRevoked = true
}

func NewAuthSession(userId uuid.UUID) AuthSession {
	return AuthSession{
		ID:        uuid.New(),
		UserID:    userId,
		ExpiresAt: time.Now().UTC().Add(LoginSessionDurationMinutes * time.Minute),
	}
}
