package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	application_dto "github.com/pelDev/health-connect/internal/application/dtos"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
)

type sessionResponse struct {
	SessionID string                 `json:"session_id"`
	ICEConfig map[string]interface{} `json:"ice_config"`
}

type aethexAdapter struct {
	client  *http.Client
	agentID string
	apiKey  string
	baseUrl string
}

func NewAethexAdapter(apiKey, agentID, baseUrl string, client *http.Client) application_ports.VoiceChatAdapter {
	return &aethexAdapter{
		client:  client,
		agentID: agentID,
		apiKey:  apiKey,
		baseUrl: baseUrl,
	}
}

func (a *aethexAdapter) InitializeSession(ctx context.Context, vid uuid.UUID) (*string, *map[string]interface{}, error) {
	url := fmt.Sprintf("%s/conversation/connect", a.baseUrl)

	payload := map[string]string{
		"agent_id": a.agentID,
	}

	payloadBytes, err := json.Marshal(&payload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	a.buildHeader(req)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed http call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, nil, fmt.Errorf("remote email provider returned status %d", resp.StatusCode)
	}

	var result sessionResponse
	json.NewDecoder(resp.Body).Decode(&result)

	return &result.SessionID, &result.ICEConfig, nil
}

func (a *aethexAdapter) SendOffer(ctx context.Context, sessionId string, input application_dto.SendVoiceChatOfferReq) (*application_dto.SendVoiceChatOfferRes, error) {
	url := fmt.Sprintf("%s/conversation/%s/offer", a.baseUrl, sessionId)

	payloadBytes, err := json.Marshal(&input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	a.buildHeader(req)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed http call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return nil, fmt.Errorf("remote email provider returned status %d", resp.StatusCode)
	}

	var result application_dto.SendVoiceChatOfferRes
	json.NewDecoder(resp.Body).Decode(&result)

	return &result, nil
}

func (a *aethexAdapter) buildHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", a.apiKey)
}
