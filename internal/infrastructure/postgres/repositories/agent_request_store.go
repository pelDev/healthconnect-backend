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

type agentRequestStore struct {
	q *sqlc.Queries
}

func (a *agentRequestStore) Delete(ctx context.Context, id uuid.UUID) error {
	panic("unimplemented")
}

func (a *agentRequestStore) FindByID(ctx context.Context, id uuid.UUID) (*models.AgentRequest, error) {
	row, err := a.q.GetAgentRequestByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	var agentRequest models.AgentRequest
	a.mapToDomain(row, &agentRequest)
	return &agentRequest, nil
}

func (a *agentRequestStore) Save(ctx context.Context, entity *models.AgentRequest) error {
	params := sqlc.CreateAgentRequestParams{
		ID:          entity.ID,
		SessionID:   entity.SessionID,
		AcceptedAt:  toPGTimestamp(entity.AcceptedAt),
		AcceptedBy:  toPGUUID(entity.AcceptedBy),
		CreatedAt:   toPGTimestamp(&entity.CreatedAt),
		RequestType: sqlc.AgentRequestTypeEnum(entity.RequestType),
		Metadata:    entity.Metadata,
	}
	return a.q.CreateAgentRequest(ctx, params)
}

func (e *agentRequestStore) mapToDomain(db sqlc.AgentRequest, m *models.AgentRequest) {
	m.ID = db.ID
	m.RequestType = models.AgentRequestType(db.RequestType)
	m.SessionID = db.SessionID
	m.AcceptedBy = fromPGUUID(db.AcceptedBy)
	m.AcceptedAt = fromPGTimestamp(db.AcceptedAt)
	m.Metadata = db.Metadata
	m.CreatedAt = *fromPGTimestamp(db.CreatedAt) // TODO: db.CreatedAt should be NOT NULL
}

func NewAgentRequestStore(q *sqlc.Queries) repositories.AgentRequestStorage {
	return &agentRequestStore{
		q: q,
	}
}
