package repositories

import (
	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
)

type MessageStorage interface {
	GetMessagesBySession(sessionID uuid.UUID) ([]models.Message, error)
	Append(sessionID uuid.UUID, msgs ...models.Message) error
}
