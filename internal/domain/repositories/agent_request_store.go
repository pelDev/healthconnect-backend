package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
)

type AgentRequestStorage interface {
	Storage[models.AgentRequest]
	ListDoctorRequests(ctx context.Context, docID uuid.UUID) ([]models.AgentRequest, error)
}
