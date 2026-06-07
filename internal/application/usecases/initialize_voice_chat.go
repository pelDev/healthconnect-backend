package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	application_dto "github.com/pelDev/health-connect/internal/application/dtos"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
	"github.com/pelDev/health-connect/internal/domain/models"
)

type initializeVoiceChatUseCase struct {
	voiceAdapter application_ports.VoiceChatAdapter
	uowFactory   func(ctx context.Context) (application_ports.UnitOfWork, error)
}

func NewInitializeVoiceChatUseCase(voiceAdapter application_ports.VoiceChatAdapter, uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error)) *initializeVoiceChatUseCase {
	return &initializeVoiceChatUseCase{
		voiceAdapter: voiceAdapter,
		uowFactory:   uowFactory,
	}
}

func (usecase *initializeVoiceChatUseCase) Execute(ctx context.Context, vid uuid.UUID) (*application_dto.InitializeSessionRes, error) {
	uowInstance, err := usecase.uowFactory(ctx)
	if err != nil {
		return nil, err
	}
	defer uowInstance.Rollback(ctx)

	reference, iceConfig, err := usecase.voiceAdapter.InitializeSession(ctx, vid)

	if err != nil {
		return nil, err
	}

	sessionStore := uowInstance.SessionStore()

	session := &models.Session{
		ID:        uuid.New(),
		VID:       vid,
		Reference: reference,
		CreatedAt: time.Now().UTC(),
	}
	err = sessionStore.Save(ctx, session)

	if err != nil {
		return nil, domain_errors.ErrDatabase(err)
	}

	err = uowInstance.Commit(ctx)
	if err != nil {
		return nil, domain_errors.ErrDatabase(err)
	}

	return &application_dto.InitializeSessionRes{
		SessionID: session.ID.String(),
		ICEConfig: *iceConfig,
	}, nil
}
