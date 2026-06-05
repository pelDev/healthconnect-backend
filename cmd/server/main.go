package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	config "github.com/pelDev/health-connect"
	sessionstore "github.com/pelDev/health-connect/internal/infrastructure/session_store"
	"github.com/pelDev/health-connect/internal/infrastructure/voice"
	http_interface "github.com/pelDev/health-connect/internal/interfaces/http"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.AethexAgentId == nil {
		log.Fatal("AethexAgentId is required but was nil")
	}

	// signal context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := &http.Client{
		Timeout: 20 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		}}

	voiceAdapter := voice.NewAethexAdapter(cfg.AethexApiKey, *cfg.AethexAgentId, cfg.AethexBaseUrl, client)

	inMemorySessionStore := sessionstore.NewInMemSessionStore()

	// Http Router
	router := http_interface.NewRouter(
		inMemorySessionStore,
		inMemorySessionStore,
		voiceAdapter,
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("🚀 Server listening on :%v", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// Wait for shutdown signal OR server error
	select {
	case <-ctx.Done():
		log.Println("🛑 Shutdown signal received...")
	case err := <-serverErr:
		if err != nil {
			log.Printf("❌ Server error: %v", err)
			stop() // Trigger shutdown
		}
	}

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	log.Println("⏳ Stopping HTTP server gracefully...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️ Server shutdown failed: %v", err)
	}

	// Check for any server errors that occurred during shutdown
	select {
	case err := <-serverErr:
		if err != nil {
			log.Printf("⚠️ Server error during shutdown: %v", err)
		}
	default:
	}

	log.Println("✅ Shutdown complete. Goodbye!")
}
