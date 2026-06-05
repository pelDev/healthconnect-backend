package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	config "github.com/pelDev/health-connect"
	"github.com/pelDev/health-connect/internal/application/usecases"
	"github.com/pelDev/health-connect/internal/infrastructure/ai"
	sessionstore "github.com/pelDev/health-connect/internal/infrastructure/session_store"
)

func main() {
	cfg := config.LoadConfig()

	ctx := context.Background()

	geminiAdapter := ai.NewGeminiAdapter(cfg.AiMaxTokens, cfg.GeminiModel, cfg.GeminiApiKey, ctx)

	inMemorySessionStore := sessionstore.NewInMemSessionStore()

	chatUseCase := usecases.NewSendChatUseCase(geminiAdapter, inMemorySessionStore, inMemorySessionStore)

	sessionID := uuid.New()
	vid := uuid.New()

	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║     HealthConnect — Chat Test CLI    ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Printf("Session ID : %s\n", sessionID)
	fmt.Printf("Visitor ID : %s\n", vid)
	fmt.Println("Type your message and press Enter. Use Ctrl+C or type 'exit' to quit.")
	fmt.Println(strings.Repeat("─", 42))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break // EOF (e.g. piped input exhausted)
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if strings.EqualFold(input, "exit") || strings.EqualFold(input, "quit") {
			fmt.Println("Goodbye.")
			break
		}

		response, err := chatUseCase.Execute(ctx, sessionID, input, vid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}

		fmt.Printf("\nAssistant: %s\n", response.Content)

		// Surface any special flags the AI returned
		// These come from tokens like EMERGENCY_FLAG / DOCTOR_PROMPT in the response
		if strings.Contains(response.Content, ai.EmergencyFlag) {
			fmt.Println("\n⚠️  EMERGENCY FLAG detected — escalate immediately.")
		}
		if strings.Contains(response.Content, ai.DoctorPrompt) {
			fmt.Println("\n🩺  DOCTOR PROMPT detected — handoff to a doctor recommended.")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scanner error: %v", err)
	}

}
