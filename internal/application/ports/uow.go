package application_ports

import (
	"context"

	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type UnitOfWork interface {
	UserRepo() repositories.UserStorage
	AuthSessionRepo() repositories.AuthSessionStorage

	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
