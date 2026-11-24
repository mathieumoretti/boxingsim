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

	// 4. Worker processing: advance the clock past the fight event and run a worker pass
	t.Run("FightProcessedByWorker", func(t *testing.T) {
		fightEvents, err := env.eventStore.GetPendingByBoxerIDAndType(ctx, domain.BoxerID(env.boxerID), model.EventTypeFightSimulate)
		if err != nil {
			t.Fatalf("GetPendingByBoxerIDAndType(fight_simulate): %v", err)
		}
		if len(fightEvents) != 1 {
			t.Fatalf("expected exactly 1 pending fight_simulate event, got %d", len(fightEvents))
		}

		// Pre-fight snapshots (W/L deltas must be exact — Review Focus #2)
		beforeA, err := env.boxerStore.GetByID(ctx, domain.BoxerID(env.boxerID))
		if err != nil {
			t.Fatalf("GetByID boxer %d: %v", env.boxerID, err)
		}
		beforeB, err := env.boxerStore.GetByID(ctx, domain.BoxerID(env.opponentID))
		if err != nil {
			t.Fatalf("GetByID opponent %d: %v", env.opponentID, err)
		}

		advanceGameTime(t, env, fightEvents[0].EventTime.Add(time.Hour))
		if n := processDueEvents(t, env); n < 1 {
			t.Fatalf("processDueEvents processed %d events, want >= 1", n)
		}

		f, err := env.fightStore.GetByID(ctx, domain.FightID(env.fightID))
		if err != nil {
			t.Fatalf("fightStore.GetByID(%d): %v", env.fightID, err)
		}
		if f.Status != model.FightStatusCompleted {
			t.Fatalf("fight status = %q, want %q", f.Status, model.FightStatusCompleted)
		}
		if f.WinnerID == nil {
			t.Fatal("completed fight has no winner")
		}
		winnerID := *f.WinnerID
		if winnerID != env.boxerID && winnerID != env.opponentID {
			t.Fatalf("winner %d is not one of the booked boxers (%d, %d)", winnerID, env.boxerID, env.opponentID)
		}
		loserID := env.opponentID
		if winnerID == env.opponentID {
			loserID = env.boxerID
		}

		afterA, err := env.boxerStore.GetByID(ctx, domain.BoxerID(env.boxerID))
		if err != nil {
			t.Fatalf("GetByID boxer %d after fight: %v", env.boxerID, err)
		}
		afterB, err := env.boxerStore.GetByID(ctx, domain.BoxerID(env.opponentID))
		if err != nil {
			t.Fatalf("GetByID opponent %d after fight: %v", env.opponentID, err)
		}
		if afterA.Health >= 100 || afterB.Health >= 100 {
			t.Errorf("expected both boxers to take damage: boxer A health=%.2f, opponent B health=%.2f", afterA.Health, afterB.Health)
		}

		if winnerID == env.boxerID {
			if afterA.Wins != beforeA.Wins+1 {
				t.Errorf("winner wins = %d, want %d (exactly +1)", afterA.Wins, beforeA.Wins+1)
			}
			if afterB.Losses != beforeB.Losses+1 {
				t.Errorf("loser losses = %d, want %d (exactly +1)", afterB.Losses, beforeB.Losses+1)
			}
			if afterB.ForcedRecoveryUntil == nil {
				t.Errorf("loser %d has nil ForcedRecoveryUntil (worker-parity callback not fired)", loserID)
			}
		} else {
			if afterB.Wins != beforeB.Wins+1 {
				t.Errorf("winner wins = %d, want %d (exactly +1)", afterB.Wins, beforeB.Wins+1)
			}
			if afterA.Losses != beforeA.Losses+1 {
				t.Errorf("loser losses = %d, want %d (exactly +1)", afterA.Losses, beforeA.Losses+1)
			}
			if afterA.ForcedRecoveryUntil == nil {
				t.Errorf("loser %d has nil ForcedRecoveryUntil (worker-parity callback not fired)", loserID)
			}
		}
	})

	// 5. Rankings: live view over boxers; both fighters must appear (Review Focus #4: 0-0 and 1-0 win rates)
	t.Run("RankingsUpdated", func(t *testing.T) {
		ranked, err := env.rankingsStore.GetRankings(ctx, model.RankingByWinRate, 10)
		if err != nil {
			t.Fatalf("GetRankings(win_rate): %v", err)
		}
		found := map[int]bool{}
		for _, rb := range ranked {
			found[rb.ID] = true
		}
		if !found[env.boxerID] {
			t.Errorf("win-rate rankings missing player boxer %d (got %d rows)", env.boxerID, len(ranked))
		}
		if !found[env.opponentID] {
			t.Errorf("win-rate rankings missing opponent boxer %d (got %d rows)", env.opponentID, len(ranked))
		}
	})

	// 6. Recovery: a rest event restores the loser's health; NULL-data rest events still complete (Review Focus #1)
	t.Run("RecoveryRestoresHealth", func(t *testing.T) {
		f, err := env.fightStore.GetByID(ctx, domain.FightID(env.fightID))
		if err != nil {
			t.Fatalf("fightStore.GetByID(%d): %v", env.fightID, err)
		}
		if f.WinnerID == nil {
			t.Fatalf("no winner recorded for fight %d (FightProcessedByWorker should have set one)", env.fightID)
		}
		winnerID := *f.WinnerID
		loserID := env.opponentID
		if winnerID == env.opponentID {
			loserID = env.boxerID
		}

		loserBefore, err := env.boxerStore.GetByID(ctx, domain.BoxerID(loserID))
		if err != nil {
			t.Fatalf("GetByID loser %d: %v", loserID, err)
		}

		gameTime, err := worldClockModel.GetCurrentGameTime(ctx, env.testDB)
		if err != nil {
			t.Fatalf("GetCurrentGameTime: %v", err)
		}
		restID := injectRestEvent(t, env, loserID, 8, gameTime.Add(time.Hour))
		advanceGameTime(t, env, gameTime.Add(2*time.Hour))
		if n := processDueEvents(t, env); n < 1 {
			t.Fatalf("processDueEvents processed %d events, want >= 1", n)
		}

		ev, err := env.eventStore.GetByID(ctx, domain.ScheduledEventID(restID))
		if err != nil {
			t.Fatalf("eventStore.GetByID(%d): %v", restID, err)
		}
		if ev.Status != model.EventStatusCompleted {
			t.Fatalf("rest event %d status = %q, want %q (error: %v)", restID, ev.Status, model.EventStatusCompleted, ev.ErrorMessage)
		}
		if ev.ProcessedAt == nil {
			t.Errorf("rest event %d processed_at is nil, want NOT NULL", restID)
		}

		loserAfter, err := env.boxerStore.GetByID(ctx, domain.BoxerID(loserID))
		if err != nil {
			t.Fatalf("GetByID loser %d after rest: %v", loserID, err)
		}
		if loserAfter.Health <= loserBefore.Health {
			t.Errorf("loser %d health after rest = %.2f, want > %.2f (pre-rest)", loserID, loserAfter.Health, loserBefore.Health)
		}

		// Review Focus #1: a rest event with NULL event_data (foreign/legacy row)
		// must process with the default rest_hours=1 and end completed, not failed.
		gameTime2, err := worldClockModel.GetCurrentGameTime(ctx, env.testDB)
		if err != nil {
			t.Fatalf("GetCurrentGameTime: %v", err)
		}
		orphanID := injectRestEventNULL(t, env, winnerID, gameTime2.Add(time.Hour))
		advanceGameTime(t, env, gameTime2.Add(2*time.Hour))
		processDueEvents(t, env)
		orphan, err := env.eventStore.GetByID(ctx, domain.ScheduledEventID(orphanID))
		if err != nil {
			t.Fatalf("eventStore.GetByID(%d): %v", orphanID, err)
		}
		if orphan.Status != model.EventStatusCompleted {
			t.Errorf("NULL-data rest event %d status = %q, want %q (error: %v)", orphanID, orphan.Status, model.EventStatusCompleted, orphan.ErrorMessage)
		}
	})

	// 7. Cleanup invariants: no pending events, no failed events, no completed events missing processed_at
	t.Run("EventCleanup", func(t *testing.T) {
		checkCount := func(desc, query string) {
			var got int
			if err := env.testDB.QueryRow(query).Scan(&got); err != nil {
				t.Fatalf("EventCleanup (%s): %v", desc, err)
			}
			if got != 0 {
				t.Errorf("EventCleanup (%s): got %d rows, want 0", desc, got)
			}
		}
		checkCount("pending events", "SELECT count(*) FROM scheduled_events WHERE processed = FALSE")
		checkCount("failed events", "SELECT count(*) FROM scheduled_events WHERE status = 'failed'")
		checkCount("completed without processed_at", "SELECT count(*) FROM scheduled_events WHERE status = 'completed' AND processed_at IS NULL")
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
	// Production applies these simulation defaults in config.Load (internal\platform\config\config.go:166-184);
	// the bare test config has a zero-valued Fight.Simulation (MaxRounds=0 → 0-round degenerate draw).
	cfg.Fight.Simulation = config.FightSimulationConfig{
		Version:          "1.0.0",
		MaxRounds:        12,
		DamageMultiplier: 0.1,
		EvasionThreshold: 0.4,
		EnergyDrain:      10.0,
		ExperienceGain:   50.0,
	}

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

// processDueEvents runs one worker-style poll-and-process pass: it fetches all pending
// events due at the current game time and processes each through the worker-parity
// EventProcessor, exactly like one iteration of the worker's event loop.
// It returns the number of due events processed.
func processDueEvents(t *testing.T, env *gameplayEnv) int {
	t.Helper()
	ctx := context.Background()
	gt, err := worldClockModel.GetCurrentGameTime(ctx, env.testDB)
	if err != nil {
		t.Fatalf("processDueEvents: GetCurrentGameTime: %v", err)
	}
	due, err := env.eventStore.GetPendingEventsBeforeGameTime(ctx, gt, 50)
	if err != nil {
		t.Fatalf("processDueEvents: GetPendingEventsBeforeGameTime: %v", err)
	}
	for _, ev := range due {
		if err := env.processor.ProcessScheduledEvent(ctx, ev); err != nil {
			t.Fatalf("processDueEvents: event %d (%s) failed: %v", ev.ID, ev.EventType, err)
		}
	}
	return len(due)
}

// injectRestEvent inserts a pending rest event directly with the same event_data shape the
// training flow writes ({"rest_hours": N, "forced_rest": true, "training_id": 0}).
// Returns the new event's id.
func injectRestEvent(t *testing.T, env *gameplayEnv, boxerID int, restHours int, at time.Time) int {
	t.Helper()
	data, _ := json.Marshal(map[string]any{"rest_hours": restHours, "forced_rest": true, "training_id": 0})
	var id int
	err := env.testDB.QueryRow(
		`INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data, created_at)
		 VALUES ($1, $2, $3, FALSE, 'pending', $4::jsonb, CURRENT_TIMESTAMP) RETURNING id`,
		boxerID, model.EventTypeRest, at, data,
	).Scan(&id)
	if err != nil {
		t.Fatalf("injectRestEvent: %v", err)
	}
	return id
}

// injectRestEventNULL inserts a pending rest event with a NULL event_data payload
// (foreign/legacy row) to pin the processor's rest_hours=1 fallback (Review Focus #1).
func injectRestEventNULL(t *testing.T, env *gameplayEnv, boxerID int, at time.Time) int {
	t.Helper()
	var id int
	err := env.testDB.QueryRow(
		`INSERT INTO scheduled_events (boxer_id, event_type, event_time, processed, status, event_data, created_at)
		 VALUES ($1, $2, $3, FALSE, 'pending', NULL, CURRENT_TIMESTAMP) RETURNING id`,
		boxerID, model.EventTypeRest, at,
	).Scan(&id)
	if err != nil {
		t.Fatalf("injectRestEventNULL: %v", err)
	}
	return id
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
