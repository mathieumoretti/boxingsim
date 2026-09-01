package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/handler"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/cors"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/platform/redis"
	"github.com/mormm/boxing/internal/service"
	"github.com/mormm/boxing/internal/store"
)

const optionsMethod = "OPTIONS"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger := logger.New("SERVER")
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	logger := logger.New("SERVER")

	logger.Info("Starting Boxing API Server")
	logger.Info("Configuration loaded",
		"dbHost", cfg.Database.Host,
		"dbPort", cfg.Database.Port,
		"dbName", cfg.Database.Name,
		"jwtSecretSet", len(cfg.JWT.Secret) > 0)

	// Initialize database
	dbConn, dbErr := database.NewPostgresDB(cfg)
	if dbErr != nil {
		logger.Error("Failed to connect to database - proceeding without database connection for UI serving", "error", dbErr)
		// Continue without database connection for web UI serving
	} else {
		logger.Info("Successfully connected to database")

		// Run migrations
		if err := db.MigrateDatabase(dbConn.DB); err != nil {
			logger.Error("Failed to run database migrations", "error", err)
			os.Exit(1)
		}

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
	var fightService *service.FightService
	if dbConn != nil {
		boxerStore = store.NewBoxerStore(dbConn.DB)
		fightService = service.NewFightService(&service.PostgresDBWrapper{Conn: dbConn.DB})
	}

	// Setup auth service for middleware
	authService := auth.NewAuthService(cfg)

	// Setup handlers
	boxerHandler := handler.NewBoxerHandler(boxerStore)
	fightHandler := handler.NewFightHandler(fightService)
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

	// Create protected subrouter for authenticated endpoints
	protectedRouter := router.NewRoute().Subrouter()
	protectedRouter.Use(authService.RequireAuth)

	// Dashboard endpoint - protected with authentication middleware
	protectedRouter.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		dashboardHandler.GetDashboard(w, r)
	}).Methods(optionsMethod, "GET")

	// Health check endpoint (public)
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
	protectedRouter.HandleFunc("/boxers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		boxerHandler.CreateBoxer(w, r)
	}).Methods(optionsMethod, "POST")

	protectedRouter.HandleFunc("/boxers/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		boxerHandler.GetBoxer(w, r)
	}).Methods(optionsMethod, "GET")

	protectedRouter.HandleFunc("/boxers/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		boxerHandler.UpdateBoxer(w, r)
	}).Methods(optionsMethod, "PUT")

	protectedRouter.HandleFunc("/users/{id}/boxers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		boxerHandler.GetBoxersByUserID(w, r)
	}).Methods(optionsMethod, "GET")

	// Fight endpoints - protected with authentication middleware
	protectedRouter.HandleFunc("/fights/book", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		fightHandler.BookFight(w, r)
	}).Methods(optionsMethod, "POST")

	protectedRouter.HandleFunc("/fights/active", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		fightHandler.GetActiveFights(w, r)
	}).Methods(optionsMethod, "GET")

	protectedRouter.HandleFunc("/fights/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		fightHandler.GetFightByID(w, r)
	}).Methods(optionsMethod, "GET")

	// Serve static files for the UI (React app)
	// For development, we'll serve from dist/ directory if it exists
	// In production, this would be handled by a separate web server or proxy
	webDir := http.Dir("./dist/")
	router.PathPrefix("/").Handler(http.FileServer(webDir)).Methods("GET")

	// Start server with configured port
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("Server starting", "address", serverAddr)
	server := &http.Server{
		Addr:         serverAddr,
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
