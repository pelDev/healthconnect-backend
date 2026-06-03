package ports

import (
	"context"

	"github.com/pelDev/health-connect/internal/domain/models"
)

type AIResponse struct {
	Content      string   // The main response text
	Tokens       []string // Extracted tokens like EMERGENCY_FLAG, DOCTOR_PROMPT
	HasEmergency bool     // Convenience flag for EMERGENCY_FLAG presence
	NeedsDoctor  bool     // Convenience flag for DOCTOR_PROMPT presence
}

type AIPort interface {
	// Chat sends a conversation history to the AI and returns a response
	// The AI follows HealthConnect system prompt rules including emergency detection
	// and doctor handoff triggers
	Chat(ctx context.Context, history []models.Message) (*AIResponse, error)

	// StreamChat sends a conversation and streams the response token by token
	// Useful for mobile clients to show real-time responses
	StreamChat(ctx context.Context, history []models.Message) (<-chan string, <-chan error)

	// ExtractSymptoms attempts to parse symptom information from conversation history
	// Returns structured symptoms for doctor handoff
	ExtractSymptoms(ctx context.Context, history []models.Message) (*models.SymptomList, error)
}
