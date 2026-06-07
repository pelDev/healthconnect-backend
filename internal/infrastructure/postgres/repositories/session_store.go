package postgres_repos

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/repositories"
	"github.com/pelDev/health-connect/internal/infrastructure/postgres/sqlc"
)

type sessionStore struct {
	q *sqlc.Queries
}

func (s *sessionStore) Delete(ctx context.Context, id uuid.UUID) error {
	return s.q.DeleteSessionByID(ctx, id)
}

func (s *sessionStore) FindByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	row, err := s.q.GetSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	var session models.Session
	s.mapToDomain(row, &session)
	return &session, nil
}

func (s *sessionStore) GetSessionByReference(ctx context.Context, reference string) (*models.Session, error) {
	row, err := s.q.GetSessionByReference(ctx, toPGText(&reference))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	var session models.Session
	s.mapToDomain(row, &session)
	return &session, nil
}

func (s *sessionStore) Save(ctx context.Context, entity *models.Session) error {
	params := sqlc.CreateSessionParams{
		ID:        entity.ID,
		Vid:       entity.VID,
		Reference: toPGText(entity.Reference),
		EndedAt:   toPGTimestamp(entity.EndedAt),
		CreatedAt: toPGTimestamp(&entity.CreatedAt),
	}

	return s.q.CreateSession(ctx, params)
}

func (e *sessionStore) mapToDomain(db sqlc.Session, m *models.Session) {
	m.ID = db.ID
	m.VID = db.Vid
	m.Reference = fromPGText(db.Reference)
	m.EndedAt = fromPGTimestamp(db.EndedAt)
	m.CreatedAt = *fromPGTimestamp(db.CreatedAt) // TODO: db.CreatedAt should be NOT NULL
}

func NewSessionStore(q *sqlc.Queries) repositories.SessionStorage {
	return &sessionStore{
		q: q,
	}
}
