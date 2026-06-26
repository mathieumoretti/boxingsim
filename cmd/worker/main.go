package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/platform/redis"
	"github.com/mormm/boxing/internal/world"
)

func main() {
	// Load configuration
	cfg := config.Load()
	logger := logger.New("WORKER")

	logger.Info("Starting Boxing World Worker")

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Error("Failed to connect to database: " + err.Error())
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("Database connected successfully")

	// Initialize Redis
	r, err := redis.New(cfg)
	if err != nil {
		logger.Error("Failed to connect to Redis: " + err.Error())
		os.Exit(1)
	}
	defer r.Close()

	logger.Info("Redis connected successfully")

	// Create world ticker
	ticker := world.NewTicker(logger)

	// Start the world ticker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Shutting down world worker...")
		cancel()
	}()

	// Start the ticker
	ticker.Start(ctx)

	logger.Info("World worker shutdown complete")
}