package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	config "github.com/pelDev/health-connect"
	"github.com/pelDev/health-connect/internal/infrastructure/ai"
)

type Voice struct {
	Country              string      `json:"country"`
	Description          interface{} `json:"description"` // Can be nil or string
	Gender               string      `json:"gender"`
	ID                   string      `json:"id"`
	IsCloned             bool        `json:"is_cloned"`
	Language             string      `json:"language"`
	Name                 string      `json:"name"`
	PreviewURL           string      `json:"preview_url"`
	SupportsDialectStyle bool        `json:"supports_dialect_style"`
	Tags                 []string    `json:"tags"`
}

type VoicesResponse []Voice

type Agent struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	SystemPrompt         string                 `json:"system_prompt"`
	FirstMessage         string                 `json:"first_message"`
	VoiceID              string                 `json:"voice_id"`
	Language             string                 `json:"language"`
	Temperature          float64                `json:"temperature"`
	MaxTokens            *int                   `json:"max_tokens"` // Pointer to handle null
	MaxDurationSeconds   int                    `json:"max_duration_seconds"`
	TransferPhoneNumber  *string                `json:"transfer_phone_number"` // Pointer to handle null
	RecordingEnabled     bool                   `json:"recording_enabled"`
	TranscriptionEnabled bool                   `json:"transcription_enabled"`
	Metadata             map[string]interface{} `json:"metadata"`
	Status               string                 `json:"status"` // "active", "inactive", etc.
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

func main() {
	cfg := config.LoadConfig()

	client := http.Client{
		Timeout: 20 * time.Second,
	}

	voicesUrl := fmt.Sprintf("%s/voices?language=english&country=NG&limit=5", cfg.AethexBaseUrl)

	req, err := http.NewRequest("GET", voicesUrl, nil)
	if err != nil {
		log.Println(fmt.Errorf("failed to create request: %w", err))
		return
	}

	setAethexHeaders(req, cfg.AethexApiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Fatalf("remote email provider returned status %d", resp.StatusCode)
	}

	var voices VoicesResponse
	json.NewDecoder(resp.Body).Decode(&voices)

	voice, err := getLastFemaleVoice(voices)
	if err != nil {
		log.Panicln(err)
	}

	createAgentUrl := fmt.Sprintf("%s/agents", cfg.AethexBaseUrl)

	createAgentPayload := map[string]interface{}{
		"name":                     "Ase",
		"system_prompt":            ai.SystemPromptVoice,
		"voice_id":                 voice.ID,
		"first_message":            "Hi, my name is Ase, your AI Health First Responser. How can I help you today?",
		"dialect_style":            "local",
		"max_tokens":               cfg.AiMaxTokens,
		"silence_timeout_seconds":  60,
		"idle_check_in_after_secs": 15,
		"max_idle_attempts":        4,
	}

	createAgentPayloadBytes, err := json.Marshal(&createAgentPayload)
	if err != nil {
		fmt.Errorf("failed to marshal payload: %w", err)
	}

	createAgentReq, err := http.NewRequest("POST", createAgentUrl, bytes.NewReader(createAgentPayloadBytes))
	if err != nil {
		log.Println(fmt.Errorf("failed to create request: %w", err))
		return
	}

	setAethexHeaders(createAgentReq, cfg.AethexApiKey)

	createAgentResp, err := client.Do(createAgentReq)
	if err != nil {
		log.Panicln(err)
	}
	defer createAgentResp.Body.Close()

	if createAgentResp.StatusCode != 201 {
		log.Fatalf("failed to create agent %d", resp.StatusCode)
	}

	var agent Agent
	if err := json.NewDecoder(createAgentResp.Body).Decode(&agent); err != nil {
		log.Panicln(fmt.Errorf("failed to decode agent response: %w", err))
	}

	// Create data directory if it doesn't exist
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Panicln(fmt.Errorf("failed to create data directory: %w", err))
	}

	// Check if agent.json already exists
	agentFilePath := filepath.Join(dataDir, "agent.json")
	if _, err := os.Stat(agentFilePath); err == nil {
		log.Panicln(fmt.Errorf("file already exists: %s. Refusing to overwrite", agentFilePath))
	} else if !os.IsNotExist(err) {
		log.Panicln(fmt.Errorf("failed to check file existence: %w", err))
	}

	// Save agent data to JSON file with pretty formatting
	agentJSON, err := json.MarshalIndent(agent, "", "  ")
	if err != nil {
		log.Panicln(fmt.Errorf("failed to marshal agent data: %w", err))
	}

	if err := os.WriteFile(agentFilePath, agentJSON, 0644); err != nil {
		log.Panicln(fmt.Errorf("failed to write agent file: %w", err))
	}

	log.Printf("Successfully created agent '%s' (ID: %s) and saved to %s\n", agent.Name, agent.ID, agentFilePath)
}

func setAethexHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
}

func getLastFemaleVoice(voices VoicesResponse) (*Voice, error) {
	for i := len(voices) - 1; i >= 0; i-- {
		if voices[i].Gender == "female" {
			return &voices[i], nil
		}
	}
	return nil, fmt.Errorf("no female voices found")
}
