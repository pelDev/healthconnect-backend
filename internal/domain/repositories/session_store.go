package repositories

import (
	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
)

type SessionStorage interface {
	CreateSession(sessionID uuid.UUID, vid uuid.UUID, ref *string) (*models.Session, error)
	GetSession(sessionID uuid.UUID) (*models.Session, error)
	Append(sessionID uuid.UUID, msgs ...models.Message) error
}
