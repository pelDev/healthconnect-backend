package ai

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/pelDev/health-connect/internal/domain/models"
	"github.com/pelDev/health-connect/internal/domain/ports"
	"google.golang.org/genai"
)

type GeminiAdapter struct {
	client    *genai.Client
	maxTokens int
	model     string
}

func NewGeminiAdapter(maxTokens int, model, apiKey string, ctx context.Context) ports.AIPort {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})

	if err != nil {
		log.Panicln(err)
	}

	return &GeminiAdapter{
		client:    client,
		maxTokens: maxTokens,
		model:     model,
	}
}

func (g *GeminiAdapter) Chat(ctx context.Context, history []models.Message) (*ports.AIResponse, error) {
	chat, err := g.client.Chats.Create(
		ctx,
		g.model,
		&genai.GenerateContentConfig{
			MaxOutputTokens:   int32(g.maxTokens),
			SystemInstruction: genai.NewContentFromText(SystemPrompt, genai.RoleModel),
		},
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	var lastUserMessageIndex int = -1

	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == models.MessageRoleUser {
			lastUserMessageIndex = i
			break
		}
	}

	if lastUserMessageIndex == -1 {
		return nil, fmt.Errorf("no user message found in history")
	}

	// Replay prior turns so Gemini has full conversation context
	// TODO: Should be optimizable
	for i := 0; i < lastUserMessageIndex; i++ {
		msg := history[i]
		if msg.Role != models.MessageRoleUser && msg.Role != models.MessageRoleAssistance {
			continue // skip doctor messages or unknown roles
		}
		if _, err := chat.SendMessage(ctx, genai.Part{Text: msg.Content}); err != nil {
			return nil, fmt.Errorf("failed to replay history at index %d: %w", i, err)
		}
	}

	result, err := chat.SendMessage(ctx, genai.Part{Text: history[lastUserMessageIndex].Content})
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	var sb strings.Builder
	for _, part := range result.Candidates[0].Content.Parts {
		if part.Text != "" {
			sb.WriteString(part.Text)
		}
	}
	content := sb.String()

	return parseAIResponse(content), nil
}

func (g *GeminiAdapter) StreamChat(ctx context.Context, history []models.Message) (<-chan string, <-chan error) {
	tokenCh := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(tokenCh)
		defer close(errCh)

		chat, err := g.client.Chats.Create(
			ctx,
			g.model,
			&genai.GenerateContentConfig{
				MaxOutputTokens:   int32(g.maxTokens),
				SystemInstruction: genai.NewContentFromText(SystemPrompt, genai.RoleModel),
			},
			nil,
		)
		if err != nil {
			errCh <- fmt.Errorf("failed to create chat: %w", err)
			return
		}

		lastUserMessageIndex := -1
		for i := len(history) - 1; i >= 0; i-- {
			if history[i].Role == models.MessageRoleUser {
				lastUserMessageIndex = i
				break
			}
		}
		if lastUserMessageIndex == -1 {
			errCh <- fmt.Errorf("no user message found in history")
			return
		}

		for i := 0; i < lastUserMessageIndex; i++ {
			msg := history[i]
			if msg.Role != models.MessageRoleUser && msg.Role != models.MessageRoleAssistance {
				continue
			}
			if _, err := chat.SendMessage(ctx, genai.Part{Text: msg.Content}); err != nil {
				errCh <- fmt.Errorf("failed to replay history at index %d: %w", i, err)
				return
			}
		}

		iter := chat.SendMessageStream(ctx, genai.Part{Text: history[lastUserMessageIndex].Content})
		for result, err := range iter {
			if err != nil {
				errCh <- fmt.Errorf("stream error: %w", err)
				return
			}
			for _, part := range result.Candidates[0].Content.Parts {
				if part.Text != "" {
					select {
					case tokenCh <- part.Text:
					case <-ctx.Done():
						errCh <- ctx.Err()
						return
					}
				}
			}
		}
	}()

	return tokenCh, errCh
}

// ExtractSymptoms implements ports.AIPort.
func (g *GeminiAdapter) ExtractSymptoms(ctx context.Context, history []models.Message) (*models.SymptomList, error) {
	panic("unimplemented")
}

func parseAIResponse(content string) *ports.AIResponse {
	knownTokens := []string{EmergencyFlag, DoctorPrompt}
	var found []string

	for _, token := range knownTokens {
		if strings.Contains(content, token) {
			found = append(found, token)
		}
	}

	return &ports.AIResponse{
		Content:      content,
		Tokens:       found,
		HasEmergency: slices.Contains(found, EmergencyFlag),
		NeedsDoctor:  slices.Contains(found, DoctorPrompt),
	}
}
