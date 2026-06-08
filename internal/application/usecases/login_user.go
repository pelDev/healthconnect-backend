package usecases

import (
	"context"

	application_dto "github.com/pelDev/health-connect/internal/application/dtos"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
	"github.com/pelDev/health-connect/internal/domain/models"
)

type loginUserUseCase struct {
	uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error)
}

func NewLoginUserUseCase(uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error)) *loginUserUseCase {
	return &loginUserUseCase{
		uowFactory: uowFactory,
	}
}

func (usecase *loginUserUseCase) Execute(ctx context.Context, req application_dto.LoginReq) (*application_dto.LoginRes, error) {
	uowInstance, err := usecase.uowFactory(ctx)
	if err != nil {
		return nil, err
	}
	defer uowInstance.Rollback(ctx)

	userRepo := uowInstance.UserRepo()
	authSessionRepo := uowInstance.AuthSessionRepo()

	user, err := userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain_errors.ErrDatabase(err)
	}

	if user == nil {
		return nil, domain_errors.ErrNotFound("user", req.Email)
	}

	isValid := user.CheckPassword(req.Password)

	if !isValid {
		return nil, domain_errors.ErrInvalidCredentials
	}

	session := models.NewAuthSession(user.ID)

	err = authSessionRepo.Save(ctx, &session)
	if err != nil {
		return nil, domain_errors.ErrDatabase(err)
	}

	err = uowInstance.Commit(ctx)
	if err != nil {
		return nil, domain_errors.ErrDatabase(err)
	}

	return &application_dto.LoginRes{
		SessionID: session.ID,
	}, nil
}
