package ai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/ports"
)

type ClaudeAdapter struct {
	client    *anthropic.Client
	maxTokens int
	model     string
}

func NewClaudeAdapter(maxTokens int, model, apiKey string) ports.AIPort {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &ClaudeAdapter{
		client:    &client,
		maxTokens: maxTokens,
		model:     model,
	}
}

func (c *ClaudeAdapter) Chat(ctx context.Context, history []models.Message) (*ports.AIResponse, error) {
	// Convert ports.Message to anthropic messages
	msgs := make([]anthropic.MessageParam, len(history))
	for i, m := range history {
		if m.Role == "user" {
			msgs[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content))
		} else {
			msgs[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content))
		}
	}

	resp, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: int64(c.maxTokens),
		System:    []anthropic.TextBlockParam{{Text: SystemPrompt}},
		Messages:  msgs,
	})
	if err != nil {
		return nil, err
	}

	content := resp.Content[0].Text

	// Extract tokens
	tokens := c.extractTokens(content)

	return &ports.AIResponse{
		Content:      content,
		Tokens:       tokens,
		HasEmergency: c.containsToken(tokens, EmergencyFlag),
		NeedsDoctor:  c.containsToken(tokens, DoctorPrompt),
	}, nil
}

func (c *ClaudeAdapter) StreamChat(ctx context.Context, history []models.Message) (<-chan string, <-chan error) {
	// Implementation for streaming (optional)
	panic("not implemented yet")
}

func (c *ClaudeAdapter) ExtractSymptoms(ctx context.Context, history []models.Message) (*models.SymptomList, error) {
	// Parse DOCTOR_PROMPT Symptoms: [...] from last assistant message
	// Or use AI to extract from conversation
	// This is a simplified version
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "assistant" && strings.Contains(history[i].Content, DoctorPrompt) {
			symptoms := c.parseSymptomsFromResponse(history[i].Content)
			return symptoms, nil
		}
	}
	return nil, nil
}

func (c *ClaudeAdapter) ValidateResponse(response string) error {
	// Check disclaimer is present
	if !strings.Contains(response, "⚠️ This is general information, not a medical diagnosis.") {
		return fmt.Errorf("missing required disclaimer")
	}

	// Check no prescription/dosage mentions (simplified)
	prescriptionKeywords := []string{"take 2 pills", "prescribe", "dosage", "mg of"}
	for _, keyword := range prescriptionKeywords {
		if strings.Contains(strings.ToLower(response), keyword) {
			return fmt.Errorf("response contains prescription language: %s", keyword)
		}
	}

	return nil
}

// Helper methods
func (c *ClaudeAdapter) extractTokens(content string) []string {
	var tokens []string
	if strings.Contains(content, EmergencyFlag) {
		tokens = append(tokens, EmergencyFlag)
	}
	if strings.Contains(content, DoctorPrompt) {
		tokens = append(tokens, DoctorPrompt)
	}
	return tokens
}

func (c *ClaudeAdapter) containsToken(tokens []string, target string) bool {
	for _, t := range tokens {
		if t == target {
			return true
		}
	}
	return false
}

func (c *ClaudeAdapter) parseSymptomsFromResponse(content string) *models.SymptomList {
	// Extract after DOCTOR_PROMPT Symptoms: [...]
	// Implementation depends on your exact format
	// This is a placeholder
	log.Println(fmt.Sprintf("parseSymptomsFromResponse content = %s", content))
	return &models.SymptomList{}
}
