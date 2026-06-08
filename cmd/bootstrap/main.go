package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

type CreateAgentToolReq struct {
	Name             string            `json:"name,required" validate:"required,lowercase,regexp=^[a-z0-9_]+$"`
	Description      string            `json:"description,omitempty"`
	ParametersSchema map[string]any    `json:"parameters_schema,omitempty"`
	EndpointURL      string            `json:"endpoint_url,required" validate:"required,url,https"`
	Headers          map[string]string `json:"headers,omitempty"`
	ToolType         string            `json:"tool_type,omitempty"` // default: "function"
}

var CreateFuncRequests = []CreateAgentToolReq{
	{
		Name: "trigger_emergency_alert",
		Description: `Call this immediately when the user mentions any life-threatening condition:
chest pain, difficulty breathing, severe bleeding, stroke symptoms (face drooping,
arm weakness, slurred speech), loss of consciousness, or suicidal thoughts.
Do not ask follow-up questions first. Call this before saying anything else.`,
		ParametersSchema: map[string]any{},
		EndpointURL:      "https://097e-154-113-67-70.ngrok-free.app/v1/aethex/function/trigger_emergency_alert",
		Headers:          map[string]string{},
		ToolType:         "function",
	},
	{
		Name: "refer_to_doctor",
		Description: `Call this when the user has symptoms that may require medical attention or a prescription.

Before calling, gather the following through natural, conversational follow-up questions — ask one at a time,
never as a list. Only ask what hasn't already been mentioned. Once you have a reasonably complete picture
and the user confirms, call this function. Do not suggest any medication or dosage yourself.

Information to collect:

PATIENT BASICS
- Age and biological sex (affects dosing, risk profiles, and differential diagnoses)
- Weight (optional, but useful for dosing context)
- Pregnancy status (if applicable)

CHIEF COMPLAINT
- Primary symptom in the patient's own words
- Which part of the body is affected

SYMPTOM DETAILS
- Onset: when did it start? Was it sudden or gradual?
- Duration: how long has it been going on?
- Character: how would they describe it? (e.g. sharp, dull, burning, throbbing, constant, intermittent)
- Severity: on a scale of 1–10, or mild/moderate/severe
- Progression: getting better, worse, or staying the same?
- Location and radiation: does it stay in one place or spread elsewhere?
- Aggravating factors: what makes it worse? (movement, food, stress, time of day, etc.)
- Relieving factors: what makes it better? (rest, position, OTC meds, etc.)

ASSOCIATED SYMPTOMS
- Any fever, chills, nausea, vomiting, fatigue, or other symptoms alongside the main complaint

RELEVANT HISTORY
- Any known medical conditions (diabetes, hypertension, asthma, etc.)
- Current medications (prescription, OTC, or supplements)
- Known allergies, especially drug allergies
- Similar episodes in the past and how they were treated
- Recent travel, sick contacts, or unusual exposures (if relevant)
- Relevant family history (e.g. heart disease, cancer, if applicable)

CONTEXT
- Has the patient seen a doctor for this before? What was the outcome?
- Have they tried anything already? Did it help?`,
		ParametersSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"symptoms": map[string]string{
					"type":        "string",
					"description": "array of symptom strings collected from the conversation",
				},
				"summary": map[string]string{
					"type":        "string",
					"description": "a short plain-English description of the case",
				},
				"age": map[string]string{
					"type":        "string",
					"description": "Patient's age",
				},
				"sex": map[string]string{
					"type":        "string",
					"description": "Patient's biological sex",
				},
				"onset": map[string]string{
					"type":        "string",
					"description": "When and how symptoms started (e.g. 'sudden onset 2 days ago')",
				},
				"duration": map[string]string{
					"type":        "string",
					"description": "How long the symptoms have been present",
				},
				"severity": map[string]string{
					"type":        "string",
					"description": "Severity of the primary symptom (e.g. '7/10', 'moderate')",
				},
				"medical_history": map[string]string{
					"type":        "string",
					"description": "Known conditions, past episodes, and relevant family history",
				},
				"current_medications": map[string]string{
					"type":        "string",
					"description": "Medications, supplements, or OTC drugs the patient is currently taking",
				},
				"allergies": map[string]string{
					"type":        "string",
					"description": "Known allergies, especially to medications",
				},
			},
			"required": []string{"symptoms", "summary", "age", "sex", "onset"},
		},
		EndpointURL: "https://097e-154-113-67-70.ngrok-free.app/v1/aethex/function/refer_to_doctor",
		Headers:     map[string]string{},
		ToolType:    "function",
	},
}

func main() {
	cfg := config.LoadConfig()

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	// Create data directory if it doesn't exist
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Panicln(fmt.Errorf("failed to create data directory: %w", err))
	}

	// Check if agent.json already exists
	agentFilePath := filepath.Join(dataDir, "agent.json")
	if _, err := os.Stat(agentFilePath); os.IsNotExist(err) {
		// Call your function here
		setupAgent(client, &cfg, dataDir, agentFilePath)
	} else if err != nil {
		// Handle other errors (e.g., permission issues)
		log.Panicln(fmt.Errorf("failed to check agent.json existence: %w", err))
	}

	// Get AgentID
	agentID, err := getAgentID(dataDir)
	if err != nil {
		log.Panicln(fmt.Errorf("failed to get agent ID: %w", err))
	}

	// Check if functions.json exists and create/update as needed
	functionsFilePath := filepath.Join(dataDir, "functions.json")

	var existingFunctions []CreateAgentToolReq

	// Read existing functions if file exists
	if _, err := os.Stat(functionsFilePath); err == nil {
		data, err := os.ReadFile(functionsFilePath)
		if err != nil {
			log.Panicln(fmt.Errorf("failed to read functions.json: %w", err))
		}

		if err := json.Unmarshal(data, &existingFunctions); err != nil {
			log.Panicln(fmt.Errorf("failed to parse functions.json: %w", err))
		}
	} else if !os.IsNotExist(err) {
		log.Panicln(fmt.Errorf("failed to check functions.json existence: %w", err))
	}

	// Create a map of existing function names for easy lookup
	existingFuncMap := make(map[string]bool)
	for _, fn := range existingFunctions {
		existingFuncMap[fn.Name] = true
	}

	// Track which functions need to be created
	var functionsToCreate []CreateAgentToolReq
	for _, functionReq := range CreateFuncRequests {
		if !existingFuncMap[functionReq.Name] {
			functionsToCreate = append(functionsToCreate, functionReq)
		}
	}

	if len(functionsToCreate) > 0 {
		for _, functionReq := range CreateFuncRequests {
			// Call your API to create the function/tool for the agent
			if err := createAgentTool(client, &cfg, agentID, functionReq); err != nil {
				log.Panicln(fmt.Errorf("failed to create tool %s: %w", functionReq.Name, err))
			}
			log.Printf("Successfully created tool: %s", functionReq.Name)

			// Append to existing functions list
			existingFunctions = append(existingFunctions, functionReq)
		}

		// Save updated functions list back to functions.json
		updatedData, err := json.MarshalIndent(existingFunctions, "", "  ")
		if err != nil {
			log.Panicln(fmt.Errorf("failed to marshal functions: %w", err))
		}

		if err := os.WriteFile(functionsFilePath, updatedData, 0644); err != nil {
			log.Panicln(fmt.Errorf("failed to write functions.json: %w", err))
		}

		log.Printf("Successfully updated functions.json with %d new function(s)", len(functionsToCreate))
	} else {
		log.Println("All functions already exist, nothing to create")
	}
}

func setupAgent(client *http.Client, cfg *config.Config, dataDir, agentFilePath string) {
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

	voice, err := getLastMaleVoice(voices)
	if err != nil {
		log.Panicln(err)
	}

	createAgentUrl := fmt.Sprintf("%s/agents", cfg.AethexBaseUrl)

	createAgentPayload := map[string]interface{}{
		"name":                     "John",
		"system_prompt":            ai.SystemPromptVoice,
		"voice_id":                 voice.ID,
		"first_message":            "Hi, my name is John, your AI Health First Responder. How can I help you today?",
		"dialect_style":            "formal",
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

func getLastMaleVoice(voices VoicesResponse) (*Voice, error) {
	for i := len(voices) - 1; i >= 0; i-- {
		if voices[i].Gender == "male" {
			return &voices[i], nil
		}
	}
	return nil, fmt.Errorf("no female voices found")
}

// Helper function to get agent ID from agent.json
func getAgentID(dataDir string) (string, error) {
	agentFilePath := filepath.Join(dataDir, "agent.json")
	data, err := os.ReadFile(agentFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read agent.json: %w", err)
	}

	var agentData struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(data, &agentData); err != nil {
		return "", fmt.Errorf("failed to parse agent.json: %w", err)
	}

	return agentData.ID, nil
}

// Placeholder for createAgentTool function - implement based on your API
func createAgentTool(client *http.Client, cfg *config.Config, agentID string, toolReq CreateAgentToolReq) error {
	url := fmt.Sprintf("%s/agents/%s/tools", cfg.AethexBaseUrl, agentID)

	body, err := json.Marshal(toolReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	setAethexHeaders(req, cfg.AethexApiKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return errors.New(fmt.Sprintf("failed to create agent %d", resp.StatusCode))
	}

	return nil
}

func setAethexHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
}
