package repositories

import (
	"context"

	"github.com/pelDev/health-connect/internal/domain/models"
)

type UserStorage interface {
	Storage[models.User]
	FindByEmail(ctx context.Context, email string) (*models.User, error)
}
