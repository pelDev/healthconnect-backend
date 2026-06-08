package usecases

import (
	"context"

	"github.com/google/uuid"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
)

type revokeAuthSessionUseCase struct {
	uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error)
}

func NewRevokeAuthSessionUseCase(uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error)) *revokeAuthSessionUseCase {
	return &revokeAuthSessionUseCase{
		uowFactory: uowFactory,
	}
}

func (usecase *revokeAuthSessionUseCase) Execute(ctx context.Context, sessionID uuid.UUID) error {
	uowInstance, err := usecase.uowFactory(ctx)
	if err != nil {
		return err
	}
	defer uowInstance.Rollback(ctx)

	authSessionRepo := uowInstance.AuthSessionRepo()

	session, err := authSessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return domain_errors.ErrDatabase(err)
	}

	if session == nil {
		return domain_errors.ErrNotFound("session", sessionID)
	}

	session.Revoke()

	err = authSessionRepo.Save(ctx, session)
	if err != nil {
		return domain_errors.ErrDatabase(err)
	}

	err = uowInstance.Commit(ctx)
	if err != nil {
		return domain_errors.ErrDatabase(err)
	}

	return nil
}
