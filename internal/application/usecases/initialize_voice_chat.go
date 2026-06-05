package usecases

import (
	"context"

	"github.com/google/uuid"
	application_dto "github.com/pelDev/health-connect/internal/application/dtos"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type initializeVoiceChatUseCase struct {
	voiceAdapter application_ports.VoiceChatAdapter
	sessionStore repositories.SessionStorage
}

func NewInitializeVoiceChatUseCase(voiceAdapter application_ports.VoiceChatAdapter, sessionStore repositories.SessionStorage) *initializeVoiceChatUseCase {
	return &initializeVoiceChatUseCase{
		voiceAdapter: voiceAdapter,
		sessionStore: sessionStore,
	}
}

func (usecase *initializeVoiceChatUseCase) Execute(ctx context.Context, vid uuid.UUID) (*application_dto.InitializeSessionRes, error) {
	reference, iceConfig, err := usecase.voiceAdapter.InitializeSession(ctx, vid)

	if err != nil {
		return nil, err
	}

	session, err := usecase.sessionStore.CreateSession(uuid.New(), vid, reference)

	return &application_dto.InitializeSessionRes{
		SessionID: session.ID.String(),
		ICEConfig: *iceConfig,
	}, nil
}
