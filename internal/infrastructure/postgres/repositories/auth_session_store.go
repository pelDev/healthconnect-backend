package postgres_repos

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	"github.com/pelDev/health-connect/internal/infrastructure/postgres/sqlc"
)

type authSessionStorage struct {
	q *sqlc.Queries
}

func (a *authSessionStorage) Delete(ctx context.Context, id uuid.UUID) error {
	return a.q.DeleteAuthSessionByID(ctx, id)
}

func (a *authSessionStorage) FindByID(ctx context.Context, id uuid.UUID) (*models.AuthSession, error) {
	row, err := a.q.GetAuthSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	var authSession models.AuthSession
	a.mapToDomain(row, &authSession)
	return &authSession, nil
}

func (a *authSessionStorage) Save(ctx context.Context, entity *models.AuthSession) error {
	params := sqlc.UpsertAuthSessionParams{
		ID:        entity.ID,
		UserID:    entity.UserID,
		CreatedAt: entity.CreatedAt,
		ExpiresAt: entity.ExpiresAt,
		IsRevoked: pgtype.Bool{
			Bool:  entity.IsRevoked,
			Valid: true,
		},
	}

	return a.q.UpsertAuthSession(ctx, params)
}

func (e *authSessionStorage) mapToDomain(db sqlc.AuthSession, m *models.AuthSession) {
	m.ID = db.ID
	m.UserID = db.UserID
	m.IsRevoked = db.IsRevoked.Bool
	m.CreatedAt = db.CreatedAt
	m.ExpiresAt = db.ExpiresAt
}

func NewAuthSessionStorage(q *sqlc.Queries) repositories.AuthSessionStorage {
	return &authSessionStorage{
		q: q,
	}
}
