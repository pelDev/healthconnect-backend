package repositories

import "github.com/pelDev/health-connect/internal/domain/models"

type AgentRequestStorage interface {
	Storage[models.AgentRequest]
}
