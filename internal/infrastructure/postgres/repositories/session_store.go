package postgres_repos

import (
	"context"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	"github.com/pelDev/health-connect/internal/infrastructure/postgres/sqlc"
)

type sessionStore struct {
	q *sqlc.Queries
}

// Delete implements repositories.SessionStorage.
func (s *sessionStore) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

// FindByID implements repositories.SessionStorage.
func (s *sessionStore) FindByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	panic("unimplemented")
}

// GetSessionByReference implements repositories.SessionStorage.
func (s *sessionStore) GetSessionByReference(reference string) (*models.Session, error) {
	panic("unimplemented")
}

// Save implements repositories.SessionStorage.
func (s *sessionStore) Save(ctx context.Context, entity *models.Session) error {
	panic("unimplemented")
}

func NewSessionStore(q *sqlc.Queries) repositories.SessionStorage {
	return &sessionStore{
		q: q,
	}
}
