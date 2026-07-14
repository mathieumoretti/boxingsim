package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/mormm/boxing/internal/handler"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/platform/redis"
	"github.com/mormm/boxing/internal/store"
)

func main() {
	// Load configuration
	cfg := config.Load()
	logger := logger.New("SERVER")

	logger.Info("Starting Boxing API Server")

	// Initialize database
	dbConn, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Error("Failed to connect to database - proceeding without database connection for UI serving", "error", err)
		// Continue without database connection for web UI serving
	} else {
		defer func() {
			if dbConn != nil {
				_ = dbConn.Close()
			}
		}()
	}

	// Initialize Redis
	redisClient, err := redis.New(cfg)
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	// Setup repositories only if DB is connected
	if dbConn != nil {
		store.NewBoxerStore(dbConn.DB)
	}

	// Setup handlers
	boxerHandler := handler.NewBoxerHandler()
	authHandler := handler.NewAuthHandler()

	// Setup router
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Authentication endpoints
	router.HandleFunc("/auth/register", authHandler.RegisterUser).Methods("POST")
	router.HandleFunc("/auth/login", authHandler.LoginUser).Methods("POST")

	// Boxer endpoints (stubbed)
	router.HandleFunc("/boxers", boxerHandler.CreateBoxer).Methods("POST")
	router.HandleFunc("/boxers/{id}", boxerHandler.GetBoxer).Methods("GET")
	router.HandleFunc("/boxers/{id}", boxerHandler.UpdateBoxer).Methods("PUT")

	// Serve static files for the UI
	webDir := http.Dir("./web/")
	router.PathPrefix("/").Handler(http.FileServer(webDir)).Methods("GET")

	// Start server
	logger.Info("Server starting on port 8080")
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", "error", err)
	}

	logger.Info("Server stopped")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "OK"}`))
}
