package usecases

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/ports"
	"github.com/pelDev/health-connect/internal/domain/repositories"
)

type sendChatUseCase struct {
	aiPort       ports.AIPort
	sessionStore repositories.SessionStorage
	messageStore repositories.MessageStorage
}

func NewSendChatUseCase(aiPort ports.AIPort, sessionStore repositories.SessionStorage, messageStore repositories.MessageStorage) *sendChatUseCase {
	return &sendChatUseCase{
		aiPort:       aiPort,
		sessionStore: sessionStore,
		messageStore: messageStore,
	}
}

func (u *sendChatUseCase) Execute(ctx context.Context, sessionID uuid.UUID, message string, vid uuid.UUID) (*models.Message, error) {
	session, err := u.sessionStore.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("ai chat failed: %w", err)
	}

	if session == nil {
		session, err = u.sessionStore.CreateSession(sessionID, vid, nil)
		if err != nil {
			return nil, fmt.Errorf("ai chat failed: %w", err)
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
	err = u.sessionStore.Append(sessionID, userMsg, assistantMsg)
	if err != nil {
		return nil, err
	}

	// TODO: Handle this cases
	_ = aiResponse.HasEmergency
	_ = aiResponse.NeedsDoctor

	return &assistantMsg, nil
}
