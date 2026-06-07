package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	config "github.com/pelDev/health-connect"
	"github.com/pelDev/health-connect/internal/domain/models"
	postgres_repos "github.com/pelDev/health-connect/internal/infrastructure/postgres/repositories"
	"github.com/pelDev/health-connect/internal/infrastructure/postgres/sqlc"
)

func main() {
	cfg := config.LoadConfig()

	// signal context for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	queries := sqlc.New(connection)

	userStorage := postgres_repos.NewUserStorage(queries)

	sampleDoc, err := models.NewUser("John", "Oleyele", "johnoleyele@healthconnect.com", "Test1234@")
	if err != nil {
		log.Panicln(err)
	}

	err = userStorage.Save(ctx, &sampleDoc)
	if err != nil {
		log.Fatal("Can't save user:", err)
		return
	}

	user, err := userStorage.FindByID(ctx, sampleDoc.ID)
	if err != nil {
		log.Fatal("Can't find user:", err)
		return
	}

	log.Println(fmt.Sprintf("User created = %s", user.ID))
}
