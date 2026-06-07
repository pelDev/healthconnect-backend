package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type sendChatUseCase struct {
	aiPort       ports.AIPort
	uowFactory   func(ctx context.Context) (application_ports.UnitOfWork, error)
	sessionStore repositories.SessionStorage
	messageStore repositories.MessageStorage
}

func NewSendChatUseCase(
	aiPort ports.AIPort,
	sessionStore repositories.SessionStorage,
	uowFactory func(ctx context.Context) (application_ports.UnitOfWork, error),
	messageStore repositories.MessageStorage,
) *sendChatUseCase {
	return &sendChatUseCase{
		aiPort:       aiPort,
		uowFactory:   uowFactory,
		sessionStore: sessionStore,
		messageStore: messageStore,
	}
}

func (u *sendChatUseCase) Execute(ctx context.Context, sessionID uuid.UUID, message string, vid uuid.UUID) (*models.Message, error) {
	session, err := u.sessionStore.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("ai chat failed: %w", err)
	}

	if session == nil {
		uowInstance, err := u.uowFactory(ctx)
		if err != nil {
			return nil, err
		}
		defer uowInstance.Rollback(ctx)

		sessionStore := uowInstance.SessionStore()

		session := &models.Session{
			ID:        uuid.New(),
			VID:       vid,
			CreatedAt: time.Now().UTC(),
			Reference: nil,
			EndedAt:   nil,
		}
		err = sessionStore.Save(ctx, session)
		if err != nil {
			return nil, domain_errors.ErrDatabase(err)
		}

		err = uowInstance.Commit(ctx)
		if err != nil {
			return nil, domain_errors.ErrDatabase(err)
		}
	}

	history, err := u.messageStore.GetMessagesBySession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("ai chat failed: %w", err)
	}

	if history == nil {
		history = []models.Message{}
	}

	userMsg := models.Message{
		ID:      uuid.New(),
		Role:    models.MessageRoleUser,
		Content: message,
	}

	history = append(history, userMsg)

	aiResponse, err := u.aiPort.Chat(ctx, history)
	if err != nil {
		return nil, fmt.Errorf("ai chat failed: %w", err)
	}

	assistantMsg := models.Message{
		ID:      uuid.New(),
		Role:    models.MessageRoleAssistance,
		Content: aiResponse.Content,
	}

	// Persist both turns
	err = u.messageStore.Append(sessionID, userMsg, assistantMsg)
	if err != nil {
		return nil, err
	}

	// TODO: Handle this cases
	_ = aiResponse.HasEmergency
	_ = aiResponse.NeedsDoctor

	return &assistantMsg, nil
}
