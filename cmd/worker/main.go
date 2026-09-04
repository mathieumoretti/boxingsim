package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Initialize stores and event processor
	eventStore := store.NewScheduledEventStore(db.DB)
	boxerStore := store.NewBoxerStore(db.DB)
	eventProcessor := service.NewEventProcessor(eventStore, boxerStore, *lg)

	// Start the worker loop with actual event processing
	startWorkerLoop(ctx, db, worldClock, eventStore, eventProcessor, lg)

	lg.Info("World worker shutdown complete")
}

// startWorkerLoop implements the core simulation loop
func startWorkerLoop(
	ctx context.Context,
	db *database.PostgresDB,
	worldClock *model.WorldClockModel,
	eventStore *store.ScheduledEventStore,
	eventProcessor *service.EventProcessor,
	lg *logger.Logger,
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
						lg.Error("Failed to process event ID=%d: %v", event.ID, err)
					}
				}
			}

			// Sleep before next iteration (50ms for rapid processing)
			select {
			case <-time.After(50 * time.Millisecond):
			case <-ctx.Done():
				return
			case <-quit:
				return
			}
		}
	}
}
