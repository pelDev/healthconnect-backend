package repositories

import (
	"github.com/pelDev/health-connect/internal/domain/models"
)

type SessionStorage interface {
	Storage[models.Session]
	GetSessionByReference(reference string) (*models.Session, error)
}
