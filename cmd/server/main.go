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
	"github.com/mormm/boxing/internal/platform/cors"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/platform/middleware"
	"github.com/mormm/boxing/internal/platform/redis"
	"github.com/mormm/boxing/internal/store"
)

const optionsMethod = "OPTIONS"

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
	var boxerStore *store.BoxerStore
	if dbConn != nil {
		boxerStore = store.NewBoxerStore(dbConn.DB)
	}

	// Setup handlers
	boxerHandler := handler.NewBoxerHandler(boxerStore)
	authHandler := handler.NewAuthHandler(dbConn)
	dashboardHandler := handler.NewDashboardHandler()

	// Setup router
	router := mux.NewRouter()

	// Apply CORS middleware
	router.Use(cors.Middleware)

	// Add logging middleware for debugging
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Request: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Dashboard endpoint - protected with authentication middleware
	router.HandleFunc("/dashboard", dashboardHandler.GetDashboard).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Authentication endpoints
	router.HandleFunc("/auth/register", authHandler.RegisterUser).Methods("POST")
	router.HandleFunc("/auth/login", authHandler.LoginUser).Methods("POST")

	// Handle CORS preflight requests for auth endpoints
	router.HandleFunc("/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		authHandler.RegisterUser(w, r)
	}).Methods(optionsMethod, "POST")

	router.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		authHandler.LoginUser(w, r)
	}).Methods(optionsMethod, "POST")

	// Boxer endpoints - protected with authentication middleware
	boxerRouter := router.PathPrefix("/boxers").Subrouter()
	boxerRouter.Use(middleware.AuthMiddleware)
	boxerRouter.HandleFunc("", boxerHandler.CreateBoxer).Methods("POST")
	boxerRouter.HandleFunc("/{id}", boxerHandler.GetBoxer).Methods("GET")
	boxerRouter.HandleFunc("/{id}", boxerHandler.UpdateBoxer).Methods("PUT")

	userBoxersRouter := router.PathPrefix("/users/{id}/boxers").Subrouter()
	userBoxersRouter.Use(middleware.AuthMiddleware)
	userBoxersRouter.HandleFunc("", boxerHandler.GetBoxersByUserID).Methods("GET")

	// Serve static files for the UI (React app)
	// For development, we'll serve from dist/ directory if it exists
	// In production, this would be handled by a separate web server or proxy
	webDir := http.Dir("./dist/")
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
