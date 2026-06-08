package usecases

import (
	"context"
	"encoding/json"

	application_dto "github.com/pelDev/health-connect/internal/application/dtos"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type referToDoctorUseCase struct {
	eventBus     ports.EventBus
	sessionStore repositories.SessionStorage
	uowFactory   func(ctx context.Context) (application_ports.UnitOfWork, error)
}

func NewReferToDoctorUseCase(
	eventBus ports.EventBus,
	sessionStore repositories.SessionStorage,
	uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error),
) *referToDoctorUseCase {
	return &referToDoctorUseCase{
		eventBus:     eventBus,
		sessionStore: sessionStore,
		uowFactory:   uowFactory,
	}
}

func (usecase *referToDoctorUseCase) Execute(ctx context.Context, sessionReference string, data application_dto.ReferDoctorData) error {
	session, err := usecase.sessionStore.GetSessionByReference(ctx, sessionReference)
	if err != nil {
		return domain_errors.ErrDatabase(err)
	}

	if session == nil {
		return domain_errors.ErrNotFound("session", sessionReference)
	}

	metadata := map[string]any{
		"data": data,
	}

	metadataBytes, _ := json.Marshal(metadata)

	agentRequest := models.NewAgentRequest(session.ID, models.AgentRequestTypeReferToDoc, metadataBytes)

	uowInstance, err := usecase.uowFactory(ctx)
	if err != nil {
		return err
	}
	defer uowInstance.Rollback(ctx)

	err = uowInstance.AgentRequestRepo().Save(ctx, agentRequest)
	if err != nil {
		return domain_errors.ErrDatabase(err)
	}

	err = uowInstance.Commit(ctx)
	if err != nil {
		return domain_errors.ErrDatabase(err)
	}

	for _, e := range agentRequest.GetEvents() {
		usecase.eventBus.Publish(ctx, e)
	}

	return nil
}
