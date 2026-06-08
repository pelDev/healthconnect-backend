package application_ports

import (
	"context"

	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type UnitOfWork interface {
	UserRepo() repositories.UserStorage
	SessionStore() repositories.SessionStorage
	AuthSessionRepo() repositories.AuthSessionStorage
	AgentRequestRepo() repositories.AgentRequestStorage

	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
