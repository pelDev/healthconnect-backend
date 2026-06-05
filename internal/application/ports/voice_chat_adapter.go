package application_ports

import (
	"context"

	"github.com/google/uuid"
	application_dto "github.com/pelDev/health-connect/internal/application/dtos"
)

type VoiceChatAdapter interface {
	InitializeSession(ctx context.Context, vid uuid.UUID) (*string, *map[string]interface{}, error)
	SendOffer(ctx context.Context, sessionId string, req application_dto.SendVoiceChatOfferReq) (*application_dto.SendVoiceChatOfferRes, error)
}
