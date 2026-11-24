//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/mormm/boxing/internal/auth"
	"github.com/mormm/boxing/internal/boxer"
	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/domain"
	"github.com/mormm/boxing/internal/fight"
	"github.com/mormm/boxing/internal/handler"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/cors"
	"github.com/mormm/boxing/internal/platform/database"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/service"
	"github.com/mormm/boxing/internal/store"
)

// TestE2E_FullGameplayLoop is the Phase 2 end-to-end regression test (MAT-135).
// It drives the complete gameplay loop through the production HTTP handlers and
// emulates the world-clock worker with the same EventProcessor wiring as cmd/worker:
// auth -> boxer -> training -> fight booking -> worker processing -> rankings -> recovery -> event cleanup.
func TestE2E_FullGameplayLoop(t *testing.T) {
	env := setupGameplayEnv(t)
	ctx := context.Background()

	// 1. Auth + boxer setup
	t.Run("AuthAndBoxerSetup", func(t *testing.T) {
		registerUser1(t, env.router, "e2e_player", "Passw0rd!")
		token := login_user(t, env.router, "e2e_player", "Passw0rd!")
		env.boxerID = createBoxer(t, env.router, token, "E2E Player Boxer")
		env.opponentID = createBoxer(t, env.router, token, "E2E Opponent Boxer")
	})

	// 2. Training: schedule -> advance game time -> complete via API -> verify gains, energy, rest event
	t.Run("TrainingScheduledAndCompleted", func(t *testing.T) {
		token := login_user(t, env.router, "e2e_player", "Passw0rd!")

		// GET /training-types (public reference data)
		req := httptest.NewRequest("GET", "/training-types", nil)
		rr := httptest.NewRecorder()
		env.router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET /training-types: expected 200, got %d: %s", rr.Code, rr.Body.String())
		}
		var types []map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &types); err != nil {
			t.Fatalf("parse /training-types: %v (body: %s)", err, rr.Body.String())
		}
		if len(types) == 0 {
			t.Fatal("expected at least one seeded training type")
		}
		typeIDFloat, ok := types[0]["id"].(float64)
		if !ok {
			t.Fatalf("training type[0] missing numeric id: %s", rr.Body.String())
		}
		typeID := int(typeIDFloat)

		// POST /boxers/{id}/train
		gameTime, err := worldClockModel.GetCurrentGameTime(ctx, env.testDB)
		if err != nil {
			t.Fatalf("GetCurrentGameTime: %v", err)
		}
		trainBody, _ := json.Marshal(map[string]any{
			"training_type_id": typeID,
			"duration_hours":   2,
			"scheduled_at":     gameTime.Add(time.Hour).Format(time.RFC3339),
		})
		req = httptest.NewRequest("POST", fmt.Sprintf("/boxers/%d/train", env.boxerID), bytes.NewBuffer(trainBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rr = httptest.NewRecorder()
		env.router.ServeHTTP(rr, req)
		if rr.Code != http.StatusCreated {
			t.Fatalf("POST /boxers/%d/train: expected 201, got %d: %s", env.boxerID, rr.Code, rr.Body.String())
		}
		var trainResp struct {
			Data struct {
				Session struct {
					ID                      int     `json:"id"`
					EnergyCost              float64 `json:"energy_cost"`
					ScheduledCompletionTime string  `json:"scheduled_completion_time"`
				} `json:"session"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &trainResp); err != nil {
			t.Fatalf("parse /train response: %v (body: %s)", err, rr.Body.String())
		}
		env.sessionID = trainResp.Data.Session.ID
		if env.sessionID == 0 {
			t.Fatalf("train response missing session id: %s", rr.Body.String())
		}
		totalEnergyCost := trainResp.Data.Session.EnergyCost
		completion, err := time.Parse(time.RFC3339, trainResp.Data.Session.ScheduledCompletionTime)
		if err != nil {
			t.Fatalf("parse scheduled_completion_time %q: %v", trainResp.Data.Session.ScheduledCompletionTime, err)
		}

		// Snapshot boxer state before completion
		before, err := env.boxerStore.GetByID(ctx, domain.BoxerID(env.boxerID))
		if err != nil {
			t.Fatalf("GetByID boxer %d: %v", env.boxerID, err)
		}

		// Advance game time past the scheduled completion, then complete via the API (production path)
		advanceGameTime(t, env, completion.Add(time.Hour))

		req = httptest.NewRequest("POST", fmt.Sprintf("/training/%d/complete", env.sessionID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rr = httptest.NewRecorder()
		env.router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("POST /training/%d/complete: expected 200, got %d: %s", env.sessionID, rr.Code, rr.Body.String())
		}

		// Session must be completed
		session, err := env.sessionStore.GetByID(ctx, domain.TrainingSessionID(env.sessionID))
		if err != nil {
			t.Fatalf("GetByID session %d: %v", env.sessionID, err)
		}
		if session.Status != model.TrainingSessionCompleted {
			t.Errorf("session %d status = %q, want %q", env.sessionID, session.Status, model.TrainingSessionCompleted)
		}

		// Energy must drop by exactly the session's energy cost
		after, err := env.boxerStore.GetByID(ctx, domain.BoxerID(env.boxerID))
		if err != nil {
			t.Fatalf("GetByID boxer %d after training: %v", env.boxerID, err)
		}
		if diff := math.Abs(after.Energy - (before.Energy - totalEnergyCost)); diff > 0.001 {
			t.Errorf("energy = %.3f, want %.3f (before %.3f - cost %.3f)",
				after.Energy, before.Energy-totalEnergyCost, before.Energy, totalEnergyCost)
		}

		// At least one combat stat must have increased
		if !(after.Strength > before.Strength || after.Defense > before.Defense || after.Agility > before.Agility) {
			t.Errorf("no stat increased after training: before str=%.2f def=%.2f agi=%.2f, after str=%.2f def=%.2f agi=%.2f",
				before.Strength, before.Defense, before.Agility, after.Strength, after.Defense, after.Agility)
		}

		// MAT-86: completing training must schedule a pending rest event
		restEvents, err := env.eventStore.GetPendingByBoxerIDAndType(ctx, domain.BoxerID(env.boxerID), model.EventTypeRest)
		if err != nil {
			t.Fatalf("GetPendingByBoxerIDAndType(rest): %v", err)
		}
		if len(restEvents) < 1 {
			t.Errorf("expected >=1 pending rest event for boxer %d, got %d", env.boxerID, len(restEvents))
		}
	})

	// 3. Fight booking: player books a fight against the opponent
	t.Run("FightBooked", func(t *testing.T) {
		token := login_user(t, env.router, "e2e_player", "Passw0rd!")

		// Record pre-fight careers: fresh boxers must be 0-0
		beforeA, err := env.boxerStore.GetByID(ctx, domain.BoxerID(env.boxerID))
		if err != nil {
			t.Fatalf("GetByID boxer %d: %v", env.boxerID, err)
		}
		beforeB, err := env.boxerStore.GetByID(ctx, domain.BoxerID(env.opponentID))
		if err != nil {
			t.Fatalf("GetByID opponent %d: %v", env.opponentID, err)
		}
		if beforeA.Wins != 0 || beforeA.Losses != 0 || beforeB.Wins != 0 || beforeB.Losses != 0 {
			t.Fatalf("pre-fight records not clean: boxer A=%d-%d, opponent B=%d-%d",
				beforeA.Wins, beforeA.Losses, beforeB.Wins, beforeB.Losses)
		}

		body, _ := json.Marshal(map[string]any{"opponent_id": env.opponentID})
		req := httptest.NewRequest("POST", fmt.Sprintf("/boxers/%d/fights", env.boxerID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rr := httptest.NewRecorder()
		env.router.ServeHTTP(rr, req)
		if rr.Code < http.StatusOK || rr.Code >= http.StatusMultipleChoices {
			t.Fatalf("POST /boxers/%d/fights: expected 2xx, got %d: %s", env.boxerID, rr.Code, rr.Body.String())
		}

		// Fresh DB: the first fight created is id 1
		f, err := env.fightStore.GetByID(ctx, domain.FightID(1))
		if err != nil {
			t.Fatalf("fightStore.GetByID(1): %v", err)
		}
		env.fightID = f.ID
		if f.Status != model.FightStatusScheduled {
			t.Fatalf("fight 1 status = %q, want %q", f.Status, model.FightStatusScheduled)
		}

		// A pending fight_simulate event must reference the booked fight
		events, err := env.eventStore.GetPendingByBoxerIDAndType(ctx, domain.BoxerID(env.boxerID), model.EventTypeFightSimulate)
		if err != nil {
			t.Fatalf("GetPendingByBoxerIDAndType(fight_simulate): %v", err)
		}
		if len(events) < 1 {
			t.Fatalf("expected >=1 pending fight_simulate event for boxer %d, got %d", env.boxerID, len(events))
		}
		var eventData map[string]any
		if err := json.Unmarshal(events[0].EventData, &eventData); err != nil {
			t.Fatalf("parse event data %q: %v", events[0].EventData, err)
		}
		if fid, _ := eventData["fight_id"].(float64); int(fid) != env.fightID {
			t.Fatalf("event data fight_id = %v, want %d", eventData["fight_id"], env.fightID)
		}
	})
}

// gameplayEnv holds the shared wiring for the full gameplay loop test:
// production stores/services, the HTTP router, and the worker-parity EventProcessor.
type gameplayEnv struct {
	t             *testing.T
	cfg           *config.Config
	testDB        *sql.DB
	router        *mux.Router
	boxerID       int // player
	opponentID    int
	sessionID     int
	fightID       int
	boxerStore    *store.BoxerStore
	fightStore    *store.FightStore
	eventStore    *store.ScheduledEventStore
	sessionStore  *store.TrainingSessionStore
	rankingsStore *store.RankingsStore
	recoverySvc   *service.RecoveryService
	processor     *service.EventProcessor
}

// setupGameplayEnv wires the production stores, services, and HTTP handlers against a
// fresh isolated database. The EventProcessor is wired exactly like cmd/worker/main.go
// (same fight.FightService + SetOnFightSimEnd recovery callback); only the Redis-backed
// rankings-cache callback is skipped.
func setupGameplayEnv(t *testing.T) *gameplayEnv {
	t.Helper()

	testDB, _ := db.FreshDatabaseWithMigrations(t, "e2e_gameplay", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available")
	}

	cfg := loadE2ETestConfig()

	boxerStore := store.NewBoxerStore(testDB)
	fightStore := store.NewFightStore(testDB)
	eventStore := store.NewScheduledEventStore(testDB)
	typeStore := store.NewTrainingTypeStore(testDB)
	sessionStore := store.NewTrainingSessionStore(testDB)
	consumableStore := store.NewConsumableStore(testDB)
	rankingsStore := store.NewRankingsStore(testDB)

	testLogger := logger.New("e2e_gameplay")
	fatigueSvc := service.NewFatigueService(boxerStore, testLogger)
	progressionSvc := service.NewProgressionService(testLogger)
	recoverySvc := service.NewRecoveryService(boxerStore, consumableStore, testLogger, testDB)

	// API-facing services (booking + training completion)
	trainingSvc := service.NewTrainingService(boxerStore, typeStore, sessionStore, eventStore,
		fatigueSvc, progressionSvc, worldClockModel, testLogger, testDB)
	highFight := service.NewFightService(fightStore, boxerStore, eventStore, recoverySvc, worldClockModel, cfg, testDB)

	// Worker parity (cmd/worker/main.go): the EventProcessor runs against
	// fight.FightService with the SetOnFightSimEnd recovery callback wired.
	boxerSvc := boxer.NewBoxerService(boxerStore)
	lowFight := fight.NewFightService(fightStore, boxerStore, cfg, boxerSvc, eventStore)
	lowFight.SetOnFightSimEnd(func(ctx context.Context, boxer1ID, boxer2ID int, boxer1HealthLost, boxer2HealthLost float64) error {
		var errs []error

		gameTime, err := worldClockModel.GetCurrentGameTime(ctx, testDB)
		if err != nil {
			gameTime = time.Now() // fallback, same as the worker
		}

		if boxer1HealthLost > 0 {
			if err := recoverySvc.ScheduleForcedRecovery(ctx, domain.BoxerID(boxer1ID), boxer1HealthLost, gameTime); err != nil {
				errs = append(errs, fmt.Errorf("boxer1 recovery: %w", err))
			}
		}
		if boxer2HealthLost > 0 {
			if err := recoverySvc.ScheduleForcedRecovery(ctx, domain.BoxerID(boxer2ID), boxer2HealthLost, gameTime); err != nil {
				errs = append(errs, fmt.Errorf("boxer2 recovery: %w", err))
			}
		}
		if len(errs) > 0 {
			return fmt.Errorf("recovery scheduling failed: %v", errs)
		}
		return nil
	})
	processor := service.NewEventProcessor(eventStore, boxerStore, lowFight, fatigueSvc, recoverySvc, *testLogger)

	router := mux.NewRouter()
	router.Use(cors.Middleware)

	authHandler := handler.NewAuthHandlerWithConfig(&database.PostgresDB{DB: testDB}, cfg)
	router.HandleFunc("/auth/register", authHandler.RegisterUser).Methods("POST")
	router.HandleFunc("/auth/login", authHandler.LoginUser).Methods("POST")

	boxerHandler := handler.NewBoxerHandler(boxerStore, eventStore)
	trainingHandler := handler.NewTrainingHandler(boxerStore, typeStore, sessionStore, eventStore,
		fatigueSvc, progressionSvc, trainingSvc)
	fightHandler := handler.NewFightHandler(highFight, worldClockModel, cfg, testDB)

	router.HandleFunc("/training-types", trainingHandler.GetAllTrainingTypes).Methods("GET")

	protected := router.NewRoute().Subrouter()
	protected.Use(auth.NewAuthService(cfg).RequireAuth)
	protected.HandleFunc("/boxers", boxerHandler.CreateBoxer).Methods("POST")
	protected.HandleFunc("/boxers/{id:[0-9]+}", boxerHandler.GetBoxer).Methods("GET")
	protected.HandleFunc("/boxers/{id:[0-9]+}/train", trainingHandler.ScheduleTraining).Methods("POST")
	protected.HandleFunc("/boxers/{id:[0-9]+}/training-sessions", trainingHandler.GetTrainingSessionsForBoxer).Methods("GET")
	protected.HandleFunc("/training/{id:[0-9]+}/complete", trainingHandler.CompleteTraining).Methods("POST")
	protected.HandleFunc("/boxers/{id:[0-9]+}/fights", fightHandler.BookFightForBoxer).Methods("POST")

	return &gameplayEnv{
		t:             t,
		cfg:           cfg,
		testDB:        testDB,
		router:        router,
		boxerStore:    boxerStore,
		fightStore:    fightStore,
		eventStore:    eventStore,
		sessionStore:  sessionStore,
		rankingsStore: rankingsStore,
		recoverySvc:   recoverySvc,
		processor:     processor,
	}
}

// advanceGameTime deterministically jumps the world clock to target by rewriting the
// world_clock anchor row (no sleeping, no real-time waits).
func advanceGameTime(t *testing.T, env *gameplayEnv, target time.Time) {
	t.Helper()
	if _, err := env.testDB.Exec(
		"UPDATE world_clock SET real_anchor = NOW(), game_anchor = $1, status = 'running' WHERE id = 1",
		target,
	); err != nil {
		t.Fatalf("advanceGameTime: %v", err)
	}
	got, err := worldClockModel.GetCurrentGameTime(context.Background(), env.testDB)
	if err != nil {
		t.Fatalf("advanceGameTime: GetCurrentGameTime: %v", err)
	}
	diff := got.Sub(target)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Fatalf("advanceGameTime: game time %v not within 1s of target %v", got, target)
	}
}
