package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mormm/boxing/internal/boxer"
	"github.com/mormm/boxing/internal/fight"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/platform/redis"
	"github.com/mormm/boxing/internal/service"
	"github.com/mormm/boxing/internal/store"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger := logger.New("WORKER")
		logger.Error("Failed to load configuration: " + err.Error())
		os.Exit(1)
	}
	lg := logger.New("WORKER")

	lg.Info("Starting Boxing World Worker")
	lg.Info("Configuration: PollInterval=%dms, StartupJitterMax=%dms, GameSpeedFactor=%.1f",
		cfg.Worker.PollIntervalMS, cfg.Worker.StartupJitterMaxMS, cfg.Worker.GameSpeedFactor)

	// Apply startup jitter to prevent thundering herd (random 0-max delay)
	if cfg.Worker.StartupJitterMaxMS > 0 {
		jitterMs := rand.Intn(cfg.Worker.StartupJitterMaxMS)
		lg.Info("Applying startup jitter: %dms", jitterMs)
		time.Sleep(time.Duration(jitterMs) * time.Millisecond)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		lg.Error("Failed to connect to database: " + err.Error())
		os.Exit(1)
	}
	defer func() {
		_ = db.Close()
	}()

	lg.Info("Database connected successfully")

	// Initialize Redis
	r, err := redis.New(cfg)
	if err != nil {
		lg.Error("Failed to connect to Redis: " + err.Error())
		os.Exit(1)
	}
	defer func() {
		_ = r.Close()
	}()

	lg.Info("Redis connected successfully")

	// Acquire worker authority lock (ensures only one worker runs at a time)
	lockAcquirer := database.NewLockAcquirer(db.DB)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := lockAcquirer.AcquireWorkerAuthority(ctx); err != nil {
		lg.Error("Failed to acquire worker authority: " + err.Error())
		os.Exit(1)
	}
	lg.Info("Acquired worker authority lock")

	// Set up graceful shutdown - release lock before exit
	defer func() {
		if err := lockAcquirer.ReleaseWorkerAuthority(ctx); err != nil {
			lg.Error("Failed to release worker authority: " + err.Error())
		} else {
			lg.Info("Released worker authority lock")
		}
	}()

	// Create world clock model for time-derived game calculations
	worldClock := model.NewWorldClockModel(lg)

	// Apply configurable game speed factor from config (for faster testing)
	if cfg.Worker.GameSpeedFactor > 0 {
		currentSpeed, err := getCurrentWorldClockSpeed(ctx, db)
		if err == nil {
			lg.Info("Current world clock speed factor: %.1f", currentSpeed)
			if currentSpeed != cfg.Worker.GameSpeedFactor {
				if err := worldClock.SetSpeedFactor(ctx, db.DB, cfg.Worker.GameSpeedFactor); err != nil {
					lg.Error("Failed to set game speed factor: %v", err)
				} else {
					lg.Info("Set game speed factor to %.1f", cfg.Worker.GameSpeedFactor)
				}
			}
		} else {
			lg.Warn("Could not read current world clock speed: %v", err)
		}
	}

	// Initialize stores
	eventStore := store.NewScheduledEventStore(db.DB)
	boxerStore := store.NewBoxerStore(db.DB)
	trainingTypeStore := store.NewTrainingTypeStore(db.DB)
	trainingSessionStore := store.NewTrainingSessionStore(db.DB)

	// Initialize fatigue service
	fatigueService := service.NewFatigueService(boxerStore, lg)

	// Initialize progression service (MAT-22)
	progressionService := service.NewProgressionService(lg)

	// Initialize boxer service (needed for fight service and training service)
	boxerSvc := boxer.NewBoxerService(boxerStore)

	// Initialize training service (for MAT-74)
	trainingService := service.NewTrainingService(boxerStore, trainingTypeStore, trainingSessionStore, eventStore, fatigueService, progressionService, worldClock, lg, db.DB)

	// Initialize fight service (for MAT-99)
	fightSvc := fight.NewFightService(db.DB, cfg, boxerSvc, eventStore)

	// Initialize event processor (for scheduled events)
	eventProcessor := service.NewEventProcessor(eventStore, boxerStore, fightSvc, fatigueService, *lg)

	lg.Info("Starting worker event loop...")

	// Start the worker loop with configurable poll interval
	pollInterval := time.Duration(cfg.Worker.PollIntervalMS) * time.Millisecond
	startWorkerLoop(ctx, db, worldClock, eventStore, eventProcessor, trainingService, lg, pollInterval)

	lg.Info("World worker shutdown complete")
}

// getCurrentWorldClockSpeed retrieves the current speed factor from the world_clock table
func getCurrentWorldClockSpeed(ctx context.Context, db *database.PostgresDB) (float64, error) {
	var speedFactor float64
	err := db.DB.QueryRowContext(ctx, "SELECT speed_factor FROM world_clock WHERE id = 1").Scan(&speedFactor)
	return speedFactor, err
}

// startWorkerLoop implements the core simulation loop with configurable poll interval
func startWorkerLoop(
	ctx context.Context,
	db *database.PostgresDB,
	worldClock *model.WorldClockModel,
	eventStore *store.ScheduledEventStore,
	eventProcessor *service.EventProcessor,
	trainingService *service.TrainingService,
	lg *logger.Logger,
	pollInterval time.Duration,
) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ctx.Done():
			return
		case <-quit:
			lg.Info("Received shutdown signal")
			return
		default:
			// Calculate current game time from database anchors
			gameTime, err := worldClock.GetCurrentGameTime(ctx, db.DB)
			if err != nil {
				lg.Error("Failed to get current game time: " + err.Error())
			} else {
				lg.Debug("Current game time: " + gameTime.Format("2006-01-02 15:04:05"))
			}

			// Query pending scheduled events due at or before current game time
			events, err := eventStore.GetPendingEventsBeforeGameTime(ctx, gameTime, 100)
			if err != nil {
				lg.Error("Failed to get pending events: " + err.Error())
			} else if len(events) > 0 {
				lg.Info("Found %d pending events to process", len(events))
				for _, event := range events {
					if err := eventProcessor.ProcessScheduledEvent(ctx, event); err != nil {
						lg.Error("Failed to process event ID=%d Type=%s BoxerID=%d: %v",
							event.ID, event.EventType, event.BoxerID, err)
					} else {
						lg.Info("Processed event ID=%d Type=%s BoxerID=%d",
							event.ID, event.EventType, event.BoxerID)
					}
				}
			}

			// Process training sessions (MAT-74)
			if trainingService != nil {
				completed, failed, err := trainingService.CompleteAllDueTrainingSessions(ctx, db.DB)
				if err != nil {
					lg.Error("Failed to process training sessions: " + err.Error())
				} else if completed > 0 {
					lg.Info("Completed %d training sessions, %d failed", completed, failed)
				}
			}

			// Sleep before next iteration (configurable poll interval)
			select {
			case <-time.After(pollInterval):
			case <-ctx.Done():
				return
			case <-quit:
				return
			}
		}
	}
}
