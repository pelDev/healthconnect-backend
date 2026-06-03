package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/ports"
)

type sendChatUseCase struct {
	aiPort       ports.AIPort
	sessionStore ports.SessionStorage
}

func NewSendChatUseCase(aiPort ports.AIPort, sessionStore ports.SessionStorage) *sendChatUseCase {
	return &sendChatUseCase{
		aiPort:       aiPort,
		sessionStore: sessionStore,
	}
}

func (u *sendChatUseCase) Execute(ctx context.Context, sessionID uuid.UUID, message string) (*models.Message, error) {
	history := u.sessionStore.GetOrCreateSession(sessionID)

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
	err = u.sessionStore.Append(sessionID, userMsg, assistantMsg)
	if err != nil {
		return nil, err
	}

	// TODO: Handle this cases
	_ = aiResponse.HasEmergency
	_ = aiResponse.NeedsDoctor

	return &assistantMsg, nil
}
