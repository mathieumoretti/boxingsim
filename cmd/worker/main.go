package main

import (
	"context"
	"errors"
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
		lg := logger.New("WORKER")
		lg.Error("Failed to load configuration", "error", err.Error())
		os.Exit(1)
	}
	lg := logger.New("WORKER")

	lg.Info("Starting Boxing World Worker")
	lg.Info("Configuration loaded",
		"poll_interval_ms", cfg.Worker.PollIntervalMS,
		"startup_jitter_max_ms", cfg.Worker.StartupJitterMaxMS,
		"game_speed_factor", cfg.Worker.GameSpeedFactor)

	// Apply startup jitter to prevent thundering herd (random 0-max delay)
	if cfg.Worker.StartupJitterMaxMS > 0 {
		jitterMs := rand.Intn(cfg.Worker.StartupJitterMaxMS)
		lg.Info("Applying startup jitter", "jitter_ms", jitterMs)
		time.Sleep(time.Duration(jitterMs) * time.Millisecond)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		lg.Error("Failed to connect to database", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		_ = db.Close()
	}()

	lg.Info("Database connected successfully")

	// Initialize Redis
	r, err := redis.New(cfg)
	if err != nil {
		lg.Error("Failed to connect to Redis", "error", err.Error())
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
		lg.Error("Failed to acquire worker authority", "error", err.Error())
		os.Exit(1)
	}
	lg.Info("Acquired worker authority lock")

	// Set up graceful shutdown - release lock before exit
	defer func() {
		if err := lockAcquirer.ReleaseWorkerAuthority(ctx); err != nil {
			lg.Error("Failed to release worker authority", "error", err.Error())
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
			lg.Info("Current world clock speed factor", "speed_factor", currentSpeed)
			if currentSpeed != cfg.Worker.GameSpeedFactor {
				if err := worldClock.SetSpeedFactor(ctx, db.DB, cfg.Worker.GameSpeedFactor); err != nil {
					lg.Error("Failed to set game speed factor", "error", err.Error())
				} else {
					lg.Info("Set game speed factor", "speed_factor", cfg.Worker.GameSpeedFactor)
				}
			}
		} else {
			lg.Warn("Could not read current world clock speed", "error", err.Error())
		}
	}

	// Initialize stores
	eventStore := store.NewScheduledEventStore(db.DB)
	boxerStore := store.NewBoxerStore(db.DB)
	trainingTypeStore := store.NewTrainingTypeStore(db.DB)
	trainingSessionStore := store.NewTrainingSessionStore(db.DB)
	fightStore := store.NewFightStore(db.DB)

	// Initialize fatigue service
	fatigueService := service.NewFatigueService(boxerStore, lg)

	// Initialize progression service (MAT-22)
	progressionService := service.NewProgressionService(lg)

	// Initialize boxer service (needed for fight service and training service)
	boxerSvc := boxer.NewBoxerService(boxerStore)

	// Initialize training service (for MAT-74)
	trainingService := service.NewTrainingService(boxerStore, trainingTypeStore, trainingSessionStore, eventStore, fatigueService, progressionService, worldClock, lg, db.DB)

	// Initialize fight service (for MAT-99)
	fightSvc := fight.NewFightService(fightStore, boxerStore, cfg, boxerSvc, eventStore)

	// Initialize event processor (for scheduled events)
	eventProcessor := service.NewEventProcessor(eventStore, boxerStore, fightSvc, fatigueService, *lg)

	// MAT-143: Initialize cleanup service for old failed events
	cleanupService := service.NewEventCleanupService(eventStore, fightStore, cfg.Events.Cleanup, lg)

	lg.Info("Starting worker event loop")

	// Start the automatic cleanup goroutine (MAT-143)
	startCleanupGoroutine(ctx, cleanupService, db, cfg.Events.Cleanup.RunIntervalMinutes, lg)

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
	// MAT-113: Crash recovery - check for events stuck in 'processing' state from previous crashes
	recoverProcessingEvents(ctx, eventStore, lg)

	// Check for failed events and log them for visibility
	logFailedEvents(ctx, eventStore, lg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	lg.Info("Worker loop started", "poll_interval", pollInterval)

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
				lg.Error("Failed to get current game time", "error", err.Error())
				// Sleep and retry on next iteration
				time.Sleep(pollInterval)
				continue
			}

			// Query pending scheduled events due at or before current game time
			events, err := eventStore.GetPendingEventsBeforeGameTime(ctx, gameTime, 100)
			if err != nil {
				lg.Error("Failed to get pending events", "error", err.Error())
			} else if len(events) > 0 {
				lg.Debug("Found pending events", "game_time", gameTime.Format("2006-01-02 15:04:05"), "event_count", len(events))
				for _, event := range events {
					// MAT-117: Log event processing with event_id correlation
					lg.Debug("Processing event", "event_id", event.ID, "event_type", event.EventType, "boxer_id", event.BoxerID)
					if err := eventProcessor.ProcessScheduledEvent(ctx, event); err != nil {
						lg.Error("Failed to process event", "event_id", event.ID, "error", err.Error())
					} else {
						lg.Debug("Event processed successfully", "event_id", event.ID)
					}
				}
			}

			// Process training sessions (MAT-74)
			if trainingService != nil {
				completed, failed, err := trainingService.CompleteAllDueTrainingSessions(ctx, db.DB)
				if err != nil {
					lg.Error("Failed to process training sessions", "error", err.Error())
				} else if completed > 0 {
					lg.Info("Training sessions completed", "completed", completed, "failed", failed)
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

// recoverProcessingEvents handles crash recovery by resetting events stuck in 'processing' state
// back to 'pending' so they can be reprocessed. This should be called on worker startup.
func recoverProcessingEvents(ctx context.Context, eventStore *store.ScheduledEventStore, lg *logger.Logger) {
	processingEvents, err := eventStore.GetProcessingEvents(ctx)
	if err != nil {
		lg.Error("Failed to query processing events for crash recovery", "error", err.Error())
		return
	}

	if len(processingEvents) > 0 {
		lg.Warn("Found events stuck in processing state", "count", len(processingEvents))
		for _, event := range processingEvents {
			if err := eventStore.ResetToPending(ctx, event.ID); err != nil {
				if !errors.Is(err, store.ErrEventNotFound) {
					lg.Error("Failed to reset event to pending", "event_id", event.ID, "error", err.Error())
				}
			} else {
				lg.Info("Recovered event from processing state", "event_id", event.ID, "event_type", event.EventType)
			}
		}
		lg.Info("Crash recovery complete", "events_reset", len(processingEvents))
	}
}

// logFailedEvents logs all failed events for visibility and monitoring purposes.
// This should be called on worker startup and periodically during runtime.
func logFailedEvents(ctx context.Context, eventStore *store.ScheduledEventStore, lg *logger.Logger) {
	failedEvents, err := eventStore.GetFailedEvents(ctx, 100) // Limit to last 100
	if err != nil {
		lg.Error("Failed to query failed events", "error", err.Error())
		return
	}

	if len(failedEvents) > 0 {
		lg.Warn("Found failed events in system", "count", len(failedEvents))
		for _, event := range failedEvents {
			errorMsg := "unknown error"
			if event.ErrorMessage != nil {
				errorMsg = *event.ErrorMessage
			}
			lg.Warn("Failed event details", "event_id", event.ID, "event_type", event.EventType, "boxer_id", event.BoxerID, "error", errorMsg)
		}
	}
}

// startCleanupGoroutine starts a background goroutine that runs the automatic cleanup job (MAT-143).
// It checks for old failed events and resolves them according to the configured interval.
func startCleanupGoroutine(
	ctx context.Context,
	cleanupService *service.EventCleanupService,
	db *database.PostgresDB,
	intervalMinutes int,
	lg *logger.Logger,
) {
	interval := time.Duration(intervalMinutes) * time.Minute

	// Run initial cleanup immediately (to handle existing zombie events)
	go func() {
		summary, err := cleanupService.Run(ctx, db.DB)
		if err != nil {
			lg.Error("Initial cleanup job failed", "error", err.Error())
		} else if summary.TotalProcessed > 0 {
			lg.Info("Initial cleanup completed",
				"processed", summary.TotalProcessed,
				"fights_cancelled", summary.FightsCancelled,
				"events_completed", summary.EventsMarkedCompleted)
		}
	}()

	// Start periodic cleanup at configured interval
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		lg.Info("Cleanup goroutine started", "interval_minutes", intervalMinutes)

		for {
			select {
			case <-ctx.Done():
				lg.Info("Cleanup goroutine shutting down")
				return
			case <-ticker.C:
				summary, err := cleanupService.Run(ctx, db.DB)
				if err != nil {
					lg.Error("Cleanup job failed", "error", err.Error())
					continue
				}

				if summary.TotalProcessed > 0 {
					lg.Info("Periodic cleanup completed",
						"processed", summary.TotalProcessed,
						"fights_cancelled", summary.FightsCancelled,
						"events_completed", summary.EventsMarkedCompleted,
						"events_skipped", summary.EventsSkipped)
				} else {
					lg.Debug("No old failed events to clean up")
				}
			}
		}
	}()
}
