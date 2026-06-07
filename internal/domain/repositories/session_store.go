package repositories

import (
	"context"

	"github.com/pelDev/health-connect/internal/domain/models"
)

type SessionStorage interface {
	Storage[models.Session]
	GetSessionByReference(ctx context.Context, reference string) (*models.Session, error)
}
