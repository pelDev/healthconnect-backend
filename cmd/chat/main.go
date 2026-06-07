package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	config "github.com/pelDev/health-connect"
	application_ports "github.com/pelDev/health-connect/internal/application/ports"
	"github.com/pelDev/health-connect/internal/application/usecases"
	"github.com/pelDev/health-connect/internal/infrastructure/ai"
	db "github.com/pelDev/health-connect/internal/infrastructure/postgres"
	postgres_repos "github.com/pelDev/health-connect/internal/infrastructure/postgres/repositories"
	"github.com/pelDev/health-connect/internal/infrastructure/postgres/sqlc"
	sessionstore "github.com/pelDev/health-connect/internal/infrastructure/session_store"
)

func main() {
	cfg := config.LoadConfig()

	ctx := context.Background()

	// Create connection pool config
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		log.Fatal("Failed to parse DSN:", err)
		return
	}

	// Configure connection pool
	maxConns, minConns, maxLifetime, maxIdleTime := cfg.GetDBPoolConfig()
	poolConfig.MaxConns = maxConns
	poolConfig.MinConns = minConns
	poolConfig.MaxConnLifetime = maxLifetime
	poolConfig.MaxConnIdleTime = maxIdleTime

	// Create connection pool
	connection, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatal("Can't create database pool:", err)
		return
	}

	// Ping database
	if err := connection.Ping(ctx); err != nil {
		log.Fatal("Can't connect to database:", err)
		return
	}

	log.Println("Database connection established")

	uowFactory := func(ctx context.Context) (application_ports.UnitOfWork, error) {
		return db.NewPostgresUoW(ctx, connection)
	}

	queries := sqlc.New(connection)

	sessionStore := postgres_repos.NewSessionStore(queries)

	geminiAdapter := ai.NewGeminiAdapter(cfg.AiMaxTokens, cfg.GeminiModel, cfg.GeminiApiKey, ctx)

	inMemorySessionStore := sessionstore.NewInMemSessionStore()

	chatUseCase := usecases.NewSendChatUseCase(geminiAdapter, sessionStore, uowFactory, inMemorySessionStore)

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
