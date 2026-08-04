package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/handler"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/cors"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/platform/redis"
	"github.com/mormm/boxing/internal/store"
)

const optionsMethod = "OPTIONS"

func main() {
	// Load configuration
	cfg := config.Load()
	logger := logger.New("SERVER")

	logger.Info("Starting Boxing API Server")
	logger.Info("Configuration loaded",
		"dbHost", cfg.DBHost,
		"dbPort", cfg.DBPort,
		"dbName", cfg.DBName,
		"jwtSecretSet", len(cfg.JWTSecret) > 0)

	// Check if migration command was passed
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			logger.Info("Running database migrations...")

			// Initialize database connection for migrations
			dbConn, err := database.NewPostgresDB(cfg)
			if err != nil {
				logger.Error("Failed to connect to database for migrations", "error", err)
				os.Exit(1)
			}
			defer func() {
				if dbConn != nil {
					_ = dbConn.Close()
				}
			}()

			// Run migrations
			if err := db.MigrateDatabase(dbConn.DB); err != nil {
				logger.Error("Failed to run database migrations", "error", err)
				os.Exit(1)
			}

			logger.Info("Migrations completed successfully")
			return
		case "create-migration":
			if len(os.Args) < 3 {
				logger.Error("Migration name required")
				os.Exit(1)
			}
			name := os.Args[2]
			if err := db.CreateMigration(name); err != nil {
				logger.Error("Failed to create migration", "error", err)
				os.Exit(1)
			}
			return
		case "reset":
			logger.Info("Resetting database...")

			// Initialize database connection for reset
			dbConn, err := database.NewPostgresDB(cfg)
			if err != nil {
				logger.Error("Failed to connect to database for reset", "error", err)
				os.Exit(1)
			}
			defer func() {
				if dbConn != nil {
					_ = dbConn.Close()
				}
			}()

			// Reset database
			if err := db.ResetDatabase(dbConn.DB); err != nil {
				logger.Error("Failed to reset database", "error", err)
				os.Exit(1)
			}

			logger.Info("Database reset completed successfully")
			return
		case "status":
			logger.Info("Checking database status...")

			// Initialize database connection for status check
			dbConn, err := database.NewPostgresDB(cfg)
			if err != nil {
				logger.Error("Failed to connect to database for status", "error", err)
				os.Exit(1)
			}
			defer func() {
				if dbConn != nil {
					_ = dbConn.Close()
				}
			}()

			// Check status
			if err := db.StatusDatabase(dbConn.DB); err != nil {
				logger.Error("Failed to check database status", "error", err)
				os.Exit(1)
			}

			return
		}
	}

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
	router.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		dashboardHandler.GetDashboard(w, r)
	}).Methods(optionsMethod, "GET")

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
	router.HandleFunc("/boxers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		boxerHandler.CreateBoxer(w, r)
	}).Methods(optionsMethod, "POST")

	router.HandleFunc("/boxers/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		boxerHandler.GetBoxer(w, r)
	}).Methods(optionsMethod, "GET")

	router.HandleFunc("/boxers/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		boxerHandler.UpdateBoxer(w, r)
	}).Methods(optionsMethod, "PUT")

	router.HandleFunc("/users/{id}/boxers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == optionsMethod {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}
		boxerHandler.GetBoxersByUserID(w, r)
	}).Methods(optionsMethod, "GET")

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
