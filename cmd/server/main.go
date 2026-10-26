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
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/cors"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/platform/middleware"
	"github.com/mormm/boxing/internal/platform/redis"
	"github.com/mormm/boxing/internal/platform/spa"
	"github.com/mormm/boxing/internal/service"
	"github.com/mormm/boxing/internal/store"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger := logger.New("SERVER")
		logger.Error("Failed to load configuration", "error", err.Error())
		os.Exit(1)
	}
	logger := logger.New("SERVER")

	logger.Info("Starting Boxing API Server")
	logger.Info("Configuration loaded",
		"db_host", cfg.Database.Host,
		"db_port", cfg.Database.Port,
		"db_name", cfg.Database.Name,
		"jwt_secret_set", len(cfg.JWT.Secret) > 0)

	// Initialize database
	dbConn, dbErr := database.NewPostgresDB(cfg)
	if dbErr != nil {
		logger.Error("Failed to connect to database - proceeding without database connection for UI serving", "error", dbErr.Error())
		// Continue without database connection for web UI serving
	} else {
		logger.Info("Successfully connected to database")

		// Run migrations
		if err := db.MigrateDatabase(dbConn.DB); err != nil {
			logger.Error("Failed to run database migrations", "error", err.Error())
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
		logger.Error("Failed to connect to Redis", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		_ = redisClient.Close()
	}()

	// Setup repositories only if DB is connected
	var boxerStore *store.BoxerStore
	var trainingTypeStore *store.TrainingTypeStore
	var trainingSessionStore *store.TrainingSessionStore
	var scheduledEventStore *store.ScheduledEventStore
	var rankingsStore *store.RankingsStore
	var fightService *service.FightService
	var fatigueService *service.FatigueService
	var progressionService *service.ProgressionService
	var trainingService *service.TrainingService
	var rankingsService *service.RankingsService
	var worldClockModel *model.WorldClockModel
	var opponentDiscoveryService *service.OpponentDiscoveryService
	if dbConn != nil {
		boxerStore = store.NewBoxerStore(dbConn.DB)
		trainingTypeStore = store.NewTrainingTypeStore(dbConn.DB)
		trainingSessionStore = store.NewTrainingSessionStore(dbConn.DB)
		scheduledEventStore = store.NewScheduledEventStore(dbConn.DB)
		rankingsStore = store.NewRankingsStore(dbConn.DB)
		fightStore := store.NewFightStore(dbConn.DB)
		worldClockModel = model.NewWorldClockModel(logger)
		fightService = service.NewFightService(fightStore, boxerStore, scheduledEventStore, worldClockModel, cfg, dbConn.DB)
		fatigueService = service.NewFatigueService(boxerStore, logger)
		progressionService = service.NewProgressionService(logger)
		trainingService = service.NewTrainingService(boxerStore, trainingTypeStore, trainingSessionStore, scheduledEventStore, fatigueService, progressionService, worldClockModel, logger, dbConn.DB)
		rankingsService = service.NewRankingsService(rankingsStore, redisClient.Client, logger)
		opponentDiscoveryService = service.NewOpponentDiscoveryService(dbConn.DB)
	}

	// Setup auth service for middleware
	authService := auth.NewAuthService(cfg)

	// Setup handlers only if DB is connected
	var boxerHandler *handler.BoxerHandler
	var trainingHandler *handler.TrainingHandler
	var rankingsHandler *handler.RankingsHandler
	var fightHandler *handler.FightHandler
	var opponentHandler *handler.OpponentHandler
	var authHandler *handler.AuthHandler
	var dashboardHandler *handler.DashboardHandler
	var worldClockHandler *handler.WorldClockHandler

	if dbConn != nil {
		boxerHandler = handler.NewBoxerHandler(boxerStore, scheduledEventStore)
		trainingHandler = handler.NewTrainingHandler(
			boxerStore,
			trainingTypeStore,
			trainingSessionStore,
			scheduledEventStore,
			fatigueService,
			progressionService,
			trainingService,
		)
		rankingsHandler = handler.NewRankingsHandler(rankingsService)
		fightHandler = handler.NewFightHandler(fightService, worldClockModel, cfg, dbConn.DB)
		opponentHandler = handler.NewOpponentHandler(opponentDiscoveryService)
		authHandler = handler.NewAuthHandler(dbConn)
		dashboardHandler = handler.NewDashboardHandler()
		worldClockHandler = handler.NewWorldClockHandler(dbConn.DB)
	}

	// Setup router
	router := mux.NewRouter()

	// Apply CORS middleware
	router.Use(cors.Middleware)

	// Add request ID middleware (generates unique ID per request for tracing)
	router.Use(middleware.RequestIDMiddleware)

	// Add logging middleware with request ID correlation
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := middleware.RequestIDFromRequest(r)
			logger.Info("Request", "request_id", requestID, "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Create protected subrouter for authenticated endpoints
	protectedRouter := router.NewRoute().Subrouter()
	protectedRouter.Use(authService.RequireAuth)

	// Dashboard endpoint - protected with authentication middleware
	protectedRouter.HandleFunc("/dashboard", dashboardHandler.GetDashboard).Methods("GET")

	// Health check endpoint (public, outside API routes)
	router.HandleFunc("/health", healthCheck).Methods("GET")

	// Authentication endpoints (CORS preflight handled by middleware)
	// Note: webpack dev server proxies /api/* to backend and strips the /api prefix
	router.HandleFunc("/auth/register", authHandler.RegisterUser).Methods("POST")
	router.HandleFunc("/auth/login", authHandler.LoginUser).Methods("POST")

	// Boxer endpoints - protected with authentication middleware
	protectedRouter.HandleFunc("/boxers", boxerHandler.CreateBoxer).Methods("POST")

	// More specific boxer routes MUST come before /boxers/{id} (Gorilla Mux route ordering)
	protectedRouter.HandleFunc("/boxers/{id}/fights", fightHandler.BookFightForBoxer).Methods("POST")
	protectedRouter.HandleFunc("/boxers/{id}/opponents", opponentHandler.GetOpponents).Methods("GET")

	// Upcoming fight endpoint for a specific boxer (MAT-106)
	protectedRouter.HandleFunc("/boxers/{id}/upcoming-fight", fightHandler.GetUpcomingFightForBoxer).Methods("GET")

	// Fight history endpoint with opponent names (MAT-103)
	protectedRouter.HandleFunc("/boxers/{id}/fights-history", fightHandler.GetFightHistoryWithOpponents).Methods("GET")

	// Training history and progression endpoints (MAT-147) - must be before /boxers/{id} catch-all
	protectedRouter.HandleFunc("/boxers/{id}/training-history", trainingHandler.GetTrainingHistory).Methods("GET")
	protectedRouter.HandleFunc("/boxers/{id}/stat-progression", trainingHandler.GetStatProgression).Methods("GET")

	// Training endpoints (MAT-147) - must be before /boxers/{id} catch-all
	protectedRouter.HandleFunc("/boxers/{id}/train", trainingHandler.ScheduleTraining).Methods("POST")
	protectedRouter.HandleFunc("/boxers/{id}/training-sessions", trainingHandler.GetTrainingSessionsForBoxer).Methods("GET")

	// Now the catch-all boxer ID routes
	protectedRouter.HandleFunc("/boxers/{id}", boxerHandler.GetBoxer).Methods("GET")
	protectedRouter.HandleFunc("/boxers/{id}", boxerHandler.UpdateBoxer).Methods("PUT")

	protectedRouter.HandleFunc("/users/{id}/boxers", boxerHandler.GetBoxersByUserID).Methods("GET")

	// Fight endpoints - protected with authentication middleware
	protectedRouter.HandleFunc("/fights/book", fightHandler.BookFight).Methods("POST")
	protectedRouter.HandleFunc("/fights/active", fightHandler.GetActiveFights).Methods("GET")
	protectedRouter.HandleFunc("/fights/{id}", fightHandler.GetFightByID).Methods("GET")

	// Opponent discovery endpoints (best_match and validate)
	protectedRouter.HandleFunc("/opponents/best_match", opponentHandler.GetBestMatch).Methods("GET")
	protectedRouter.HandleFunc("/opponents/validate", opponentHandler.ValidateMatch).Methods("POST")

	// Training types endpoint - public (reference data)
	router.HandleFunc("/training-types", trainingHandler.GetAllTrainingTypes).Methods("GET")

	// World clock endpoint - public (game time is not user-specific)
	router.HandleFunc("/world/time", worldClockHandler.GetCurrentGameTime).Methods("GET")

	// Training completion endpoint - protected with authentication middleware
	protectedRouter.HandleFunc("/training/{id}/complete", trainingHandler.CompleteTraining).Methods("POST")

	// Bulk training completion endpoint (development feature) - protected with authentication middleware
	protectedRouter.HandleFunc("/training/bulk-complete", trainingHandler.BulkCompleteTraining).Methods("POST")

	// Rankings endpoints - protected with authentication middleware
	protectedRouter.HandleFunc("/rankings", rankingsHandler.GetRankings).Methods("GET")
	protectedRouter.HandleFunc("/rankings/{boxer_id}", rankingsHandler.GetRankingForBoxer).Methods("GET")
	protectedRouter.HandleFunc("/rankings/nearby/{boxer_id}", rankingsHandler.GetNearbyRankings).Methods("GET")
	protectedRouter.HandleFunc("/rankings/invalidate", rankingsHandler.InvalidateRankings).Methods("POST")

	// Serve static files for the UI (React app) with SPA fallback
	// The SPA handler serves actual static assets and falls back to index.html for React Router paths
	spaHandler := spa.New("./dist")
	router.PathPrefix("/").Handler(spaHandler).Methods("GET", "HEAD")

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
			logger.Error("Server failed to start", "error", err.Error())
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
		logger.Error("Server shutdown error", "error", err.Error())
	}

	logger.Info("Server stopped")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "OK"}`))
}
