package ports

import (
	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
)

type SessionStorage interface {
	GetOrCreateSession(sessionID uuid.UUID) []models.Message
	Append(sessionID uuid.UUID, msgs ...models.Message) error
}
