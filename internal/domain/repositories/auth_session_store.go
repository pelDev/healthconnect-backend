package repositories

import (
	"github.com/pelDev/health-connect/internal/domain/models"
)

type AuthSessionStorage interface {
	Storage[models.AuthSession]
}
