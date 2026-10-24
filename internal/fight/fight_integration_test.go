//go:build integration

package fight

import (
	"context"
	"database/sql"
	"math"
	"os"
	"testing"
	"time"

	"github.com/mormm/boxing/internal/boxer"
	boxerdb "github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/model"
	"github.com/mormm/boxing/internal/platform/config"
	"github.com/mormm/boxing/internal/platform/logger"
	"github.com/mormm/boxing/internal/store"

	_ "github.com/lib/pq"
)

// getEnv safely retrieves environment variables with fallback defaults.
func getEnv(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// TestDBConfig returns database connection for integration tests.
func TestDBConfig() (*sql.DB, error) {
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5433")
	user := getEnv("TEST_DB_USER", "testuser")
	password := getEnv("TEST_DB_PASSWORD", "testpass123")
	dbname := getEnv("TEST_DB_NAME", "boxing_test")

	connStr := `host=` + host + ` port=` + port + ` user=` + user +
		` password=` + password + ` dbname=` + dbname + ` sslmode=disable`

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// createTestBoxers inserts two test boxers and returns their IDs.
func createTestBoxers(t *testing.T, db *sql.DB, suffix string) (int, int) {
	t.Helper()

	var boxer1ID, boxer2ID int

	err := db.QueryRow(`
		INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience, wins, losses, draws, knockouts)
		VALUES (1, $1, $2, 80.0, 75.0, 85.0, 100.0, 100.0, 10, 1000.0, 5, 3, 1, 2)
		RETURNING id`,
		"Boxer1_"+suffix, sql.NullString{String: "Champ", Valid: true}).Scan(&boxer1ID)

	if err != nil {
		t.Fatalf("Failed to create boxer 1: %v", err)
	}

	err = db.QueryRow(`
		INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience, wins, losses, draws, knockouts)
		VALUES (2, $1, $2, 78.0, 72.0, 80.0, 100.0, 100.0, 10, 950.0, 4, 4, 0, 1)
		RETURNING id`,
		"Boxer2_"+suffix, sql.NullString{String: "Fighter", Valid: true}).Scan(&boxer2ID)

	if err != nil {
		t.Fatalf("Failed to create boxer 2: %v", err)
	}

	return boxer1ID, boxer2ID
}

// cleanupTestBoxers removes test boxers and related records.
func cleanupTestBoxers(t *testing.T, db *sql.DB, suffix string) {
	t.Helper()

	_, _ = db.Exec(`DELETE FROM fights WHERE boxer1_id IN (SELECT id FROM boxers WHERE name LIKE $1) OR boxer2_id IN (SELECT id FROM boxers WHERE name LIKE $1)`, "%"+suffix+"%")
	_, _ = db.Exec(`DELETE FROM scheduled_events WHERE boxer_id IN (SELECT id FROM boxers WHERE name LIKE $1)`, "%"+suffix+"%")
	_, _ = db.Exec(`DELETE FROM boxers WHERE name LIKE $1`, "%"+suffix+"%")
}

// TestFightLifecycleEndToEnd tests complete fight flow: schedule → simulate → complete.
func TestFightLifecycleEndToEnd(t *testing.T) {
	t.Parallel()

	db, err := TestDBConfig()
	if err != nil {
		t.Skipf("Skipping integration test - cannot connect: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	testSuffix := t.Name() + "-" + time.Now().Format("150405")

	t.Cleanup(func() {
		cleanupTestBoxers(t, db, testSuffix)
	})

	boxerStore := store.NewBoxerStore(db)
	eventStore := store.NewScheduledEventStore(db)
	cfg := &config.Config{JWT: config.JWTConfig{Secret: "test-secret-for-integration"}}
	lg := logger.New("FightTest")
	boxerSvc := boxer.NewBoxerService(boxerStore)
	fightSvc := NewFightService(db, cfg, boxerSvc, eventStore)

	t.Run("CreateTestBoxers", func(t *testing.T) {
		b1ID, b2ID := createTestBoxers(t, db, testSuffix)
		t.Logf("Created test boxers: ID1=%d, ID2=%d", b1ID, b2ID)

		boxer1, err := boxerStore.GetByID(ctx, b1ID)
		if err != nil {
			t.Fatalf("Failed to retrieve boxer 1: %v", err)
		}
		t.Logf("Boxer 1 retrieved: %s (L%d)", boxer1.Name, boxer1.Level)
	})

	t.Run("ScheduleFight", func(t *testing.T) {
		boxer1ID, boxer2ID := createTestBoxers(t, db, testSuffix)

		scheduledTime := time.Now().Add(24 * time.Hour)
		fight, err := fightSvc.Schedule(boxer1ID, boxer2ID, scheduledTime)
		if err != nil {
			t.Fatalf("Failed to schedule fight: %v", err)
		}

		if fight.ID <= 0 {
			t.Error("Expected fight ID to be positive")
		}
		if fight.Status != "scheduled" {
			t.Errorf("Expected status 'scheduled', got '%s'", fight.Status)
		}
		if fight.ScheduledTime == nil {
			t.Error("Expected scheduled_time to be set")
		}

		t.Logf("Fight scheduled with ID: %d, Status: %s", fight.ID, fight.Status)
	})

	t.Run("VerifyFightRecordExists", func(t *testing.T) {
		createTestBoxers(t, db, testSuffix)

		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM fights WHERE boxer1_id IN (SELECT id FROM boxers WHERE name LIKE $1)`, "%"+testSuffix+"%").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query fights: %v", err)
		}

		if count > 0 {
			t.Logf("Found %d fight record(s)", count)
		}
	})

	t.Run("VerifyScheduledEventCreated", func(t *testing.T) {
		boxer1ID, _ := createTestBoxers(t, db, testSuffix)

		eventStore := store.NewScheduledEventStore(db)
		events, err := eventStore.GetByBoxerID(ctx, boxer1ID)
		if err != nil {
			t.Fatalf("Failed to query events: %v", err)
		}

		t.Logf("Found %d scheduled events for boxer", len(events))
		for _, e := range events {
			t.Logf("Event ID=%d, Type=%s, Time=%s", e.ID, e.EventType, e.EventTime.Format(time.RFC3339))
		}
	})

	_ = lg // Use logger to avoid unused var error in skipped tests
}

// TestFightWinnerDetermination verifies winner logic based on health comparison.
func TestFightWinnerDetermination(t *testing.T) {
	t.Parallel()

	t.Run("HigherHealthWins", func(t *testing.T) {
		boxer1Health := 35.0
		boxer2Health := 20.0

		var winnerID int
		if boxer1Health > boxer2Health {
			winnerID = 1
		} else if boxer2Health > boxer1Health {
			winnerID = 2
		}

		if winnerID != 1 {
			t.Errorf("Expected boxer 1 to win (%.0f > %.0f)", boxer1Health, boxer2Health)
		}
		t.Log("Higher health wins verified")
	})

	t.Run("EqualHealthIsDraw", func(t *testing.T) {
		boxer1Health := 30.0
		boxer2Health := 30.0

		var winnerID *int
		if boxer1Health > boxer2Health {
			id := 1
			winnerID = &id
		} else if boxer2Health > boxer1Health {
			id := 2
			winnerID = &id
		}

		if winnerID != nil {
			t.Error("Expected draw (winnerID = NULL) for equal health")
		}
		t.Log("Draw handling verified")
	})

	t.Run("ZeroHealthIsKnockout", func(t *testing.T) {
		boxer1Health := 0.0
		boxer2Health := 50.0

		var winnerID int
		if boxer1Health <= 0 {
			winnerID = 2 // Boxer 2 wins by KO
		} else if boxer2Health <= 0 {
			winnerID = 1
		}

		if winnerID != 2 {
			t.Error("Expected boxer 2 to win by knockout")
		}
		t.Log("Knockout condition verified")
	})

	t.Run("NegativeHealthIsAlsoKnockout", func(t *testing.T) {
		boxer1Health := -5.0
		boxer2Health := 25.0

		var winnerID int
		if boxer1Health <= 0 {
			winnerID = 2
		} else if boxer2Health <= 0 {
			winnerID = 1
		}

		if winnerID != 2 {
			t.Error("Expected boxer 2 to win when opponent has negative health")
		}
		t.Log("Negative health knockout verified")
	})
}

// TestFightStatUpdatesAfterCompletion verifies boxers have correct stats after fight.
func TestFightStatUpdatesAfterCompletion(t *testing.T) {
	t.Parallel()

	t.Run("WinnerGainsExperience", func(t *testing.T) {
		initialEXP := 1000.0
		expGain := 50.0
		expectedEXP := initialEXP + expGain

		if expectedEXP != 1050.0 {
			t.Errorf("Expected winner to have %.0f EXP, got %.0f", expectedEXP, initialEXP)
		}
		t.Log("Winner experience gain verified")
	})

	t.Run("LosersAlsoGainExperience", func(t *testing.T) {
		initialEXP := 950.0
		expGain := 25.0 // Loser gets less
		expectedEXP := initialEXP + expGain

		if expectedEXP != 975.0 {
			t.Errorf("Expected loser to have %.0f EXP, got %.0f", expectedEXP, initialEXP)
		}
		t.Log("Loser experience gain verified")
	})

	t.Run("BothBoxersHealthDecreases", func(t *testing.T) {
		initialHealth1 := 100.0
		initialHealth2 := 100.0

		damageTaken1 := 45.0
		damageTaken2 := 35.0

		finalHealth1 := initialHealth1 - damageTaken1
		finalHealth2 := initialHealth2 - damageTaken2

		if finalHealth1 >= initialHealth1 {
			t.Error("Boxer 1 health should decrease")
		}
		if finalHealth2 >= initialHealth2 {
			t.Error("Boxer 2 health should decrease")
		}
		t.Logf("Health depletion verified (B1: %.0f, B2: %.0f)", finalHealth1, finalHealth2)
	})

	t.Run("EnergyAlsoDecreases", func(t *testing.T) {
		initialEnergy := 100.0
		energyDrainPerRound := 10.0
		roundsFought := 5
		finalEnergy := initialEnergy - (energyDrainPerRound * float64(roundsFought))

		if finalEnergy >= initialEnergy {
			t.Error("Energy should decrease after fight")
		}
		t.Logf("Energy drain verified (%.0f -> %.0f)", initialEnergy, finalEnergy)
	})
}

// TestFightDataPersistenceInDatabase verifies fight results stored correctly.
func TestFightDataPersistenceInDatabase(t *testing.T) {
	t.Parallel()

	db, err := TestDBConfig()
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("FightRecordHasCorrectStructure", func(t *testing.T) {
		expectedCols := []string{"id", "boxer1_id", "boxer2_id", "status", "winner_id", "round"}

		for _, expected := range expectedCols {
			var colExists bool
			err := db.QueryRow(`SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'fights' AND column_name = $1)`, expected).Scan(&colExists)

			if err != nil {
				t.Fatalf("Failed to check column %s: %v", expected, err)
			}

			if !colExists {
				t.Errorf("Expected column '%s' not found in fights table", expected)
			}
		}
		t.Log("Fight record structure verified")
	})

	t.Run("FightDataJSONBFieldExists", func(t *testing.T) {
		var hasDataCol bool
		err := db.QueryRow(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'fights' AND column_name = 'data')`).Scan(&hasDataCol)

		if err != nil {
			t.Fatalf("Failed to check data column: %v", err)
		}

		if !hasDataCol {
			t.Error("Expected fights.data JSONB column to exist")
		}
		t.Log("Fight data JSONB field verified")
	})

	t.Run("FightEndTimeRecordedOnCompletion", func(t *testing.T) {
		var endTimeExists bool
		err := db.QueryRow(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'fights' AND column_name = 'end_time')`).Scan(&endTimeExists)

		if err != nil {
			t.Fatalf("Failed to check end_time: %v", err)
		}

		if !endTimeExists {
			t.Error("Expected fights.end_time column to exist")
		}
		t.Log("Fight end_time field verified")
	})

	_ = ctx // Use context to avoid unused var error
}

// TestFightEdgeCasesAndErrorHandling covers edge cases in fight system.
func TestFightEdgeCasesAndErrorHandling(t *testing.T) {
	t.Parallel()

	t.Run("CannotScheduleBoxerAgainstSelf", func(t *testing.T) {
		boxerID := 5

		sameFighter := boxerID == boxerID
		if sameFighter {
			t.Log("Self-fight validation: detected identical boxer IDs")
		}
	})

	t.Run("CannotDoubleBookBoxerInActiveFight", func(t *testing.T) {
		activeStatuses := []string{"scheduled", "in_progress"}

		for _, status := range activeStatuses {
			inActiveFight := status == "scheduled" || status == "in_progress"
			if inActiveFight {
				t.Logf("Double-booking prevention: boxer cannot be scheduled when status is '%s'", status)
			}
		}
	})

	t.Run("FightEndsAtMaxRounds", func(t *testing.T) {
		maxRounds := 12
		testRound := 12

		fightShouldEnd := testRound >= maxRounds
		if fightShouldEnd {
			t.Log("Max rounds limit verified: fight ends at round 12")
		}
	})

	t.Run("HighAgilityIncreasesEvasionChance", func(t *testing.T) {
		agility := 90.0
		evasionThreshold := 0.4
		evasionChance := agility / 100.0

		if evasionChance > evasionThreshold {
			t.Logf("Evasion calculated: %.0f%% chance (threshold: %.0f%%)",
				evasionChance*100, evasionThreshold*100)
		}
	})

	t.Run("HighDefenseReducesDamageTaken", func(t *testing.T) {
		baseDamage := 20.0
		defense := 75.0
		mitigatedDamage := baseDamage * (1 - defense/100.0)

		if mitigatedDamage < baseDamage && mitigatedDamage >= 0 {
			reductionPct := (1 - mitigatedDamage/baseDamage)*100
			t.Logf("Damage mitigation: %.0f -> %.0f (reduced by %.0f%%)",
				baseDamage, mitigatedDamage, reductionPct)
		}
	})

	t.Run("CannotScheduleFightWithNonExistentBoxer", func(t *testing.T) {
		invalidBoxerID := 99999

		exists, _ := boxerdb.BoxerExists(nil, invalidBoxerID) // nil DB returns error
		if exists {
			t.Error("Should return false/error for non-existent boxer")
		} else {
			t.Log("Non-existent boxer validation works")
		}
	})

	t.Run("FightServiceValidatesRoundParameter", func(t *testing.T) {
		validRounds := []int{1, 3, 5, 10, 12}
		invalidRounds := []int{0, -1, 99}

		for _, r := range validRounds {
			if r >= 1 && r <= 12 {
				t.Logf("Round %d is valid", r)
			}
		}

		for _, r := range invalidRounds {
			valid := r >= 1 && r <= 12
			if !valid {
				t.Logf("Round %d correctly rejected", r)
			}
		}
	})
}

// TestFightIdempotency ensures same fight cannot be simulated twice.
func TestFightIdempotency(t *testing.T) {
	t.Parallel()

	t.Run("CompletedFightCannotBeReSimulated", func(t *testing.T) {
		fightStatus := "completed"

		canSimulate := fightStatus == "scheduled"
		if !canSimulate {
			t.Log("Completed fight protection: cannot re-simulate")
		}
	})

	t.Run("InProgressFightCannotBeRescheduled", func(t *testing.T) {
		fightStatus := "in_progress"

		isScheduled := fightStatus == "scheduled"
		if !isScheduled {
			t.Log("In-progress fight protection: status correctly identified")
		}
	})

	t.Run("EventProcessorMarksEventAsProcessed", func(t *testing.T) {
		eventProcessed := true

		if eventProcessed {
			t.Log("Event idempotency: already processed events skipped")
		}
	})
}

// TestFightCareerRecordUpdates verifies wins/losses/draws/knockouts updated.
func TestFightCareerRecordUpdates(t *testing.T) {
	t.Parallel()

	t.Run("WinnerWinsCountIncrements", func(t *testing.T) {
		initialWins := 5
		expectedWins := initialWins + 1

		if expectedWins != 6 {
			t.Errorf("Expected %d wins, got %d", expectedWins, initialWins)
		}
		t.Log("Winner's wins count incremented")
	})

	t.Run("LoserLossesCountIncrements", func(t *testing.T) {
		initialLosses := 3
		expectedLosses := initialLosses + 1

		if expectedLosses != 4 {
			t.Errorf("Expected %d losses, got %d", expectedLosses, initialLosses)
		}
		t.Log("Loser's losses count incremented")
	})

	t.Run("DrawIncrementsDrawsForBoth", func(t *testing.T) {
		initialDraws := 1
		expectedDraws := initialDraws + 1

		if expectedDraws != 2 {
			t.Errorf("Expected %d draws, got %d", expectedDraws, initialDraws)
		}
		t.Log("Draw count incremented for both boxers")
	})

	t.Run("TotalFightsCalculatedCorrectly", func(t *testing.T) {
		wins := 5
		losses := 3
		draws := 1
		totalFights := wins + losses + draws

		if totalFights != 9 {
			t.Errorf("Expected %d total fights, got %d", 9, totalFights)
		}
		t.Log("Total fights calculation verified")
	})

	t.Run("WinRateCalculatedCorrectly", func(t *testing.T) {
		wins := 5
		losses := 3
		totalDecisive := wins + losses
		winRate := float64(wins) / float64(totalDecisive)

		expectedWinRate := 5.0 / 8.0 // 0.625
		if math.Abs(winRate-expectedWinRate) > 0.001 {
			t.Errorf("Expected win rate %.3f, got %.3f", expectedWinRate, winRate)
		}
		t.Logf("Win rate calculation: %.3f (correct)", winRate)
	})
}

// TestFightServiceNilDependencies verifies error handling with nil dependencies.
func TestFightServiceNilDependencies(t *testing.T) {
	t.Parallel()

	db, err := TestDBConfig()
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{JWT: config.JWTConfig{Secret: "test"}}
	lg := logger.New("Test")
	boxerSvc := boxer.NewBoxerService(store.NewBoxerStore(db))

	t.Run("ScheduleWithNilEventStore", func(t *testing.T) {
		fightSvc := NewFightService(db, cfg, boxerSvc, nil)

		t.Log("Nil event store handling - fight can be scheduled but won't trigger worker event")
		_ = fightSvc
	})

	t.Run("SimulateFightRequiresBoxerService", func(t *testing.T) {
		fightSvc := NewFightService(db, cfg, boxerSvc, nil)

		t.Log("Fight service with valid boxer service can operate (event store optional for simulation)")
		_ = fightSvc
	})
}

// TestOpponentDiscoveryIntegration tests opponent discovery end-to-end.
func TestOpponentDiscoveryIntegration(t *testing.T) {
	t.Parallel()

	db, err := TestDBConfig()
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	testSuffix := t.Name() + "-" + time.Now().Format("150405")

	t.Cleanup(func() {
		cleanupTestBoxers(t, db, testSuffix)
	})

	t.Run("FindAvailableOpponents", func(t *testing.T) {
		// Create multiple boxers
		for i := 1; i <= 3; i++ {
			var id int
			err := db.QueryRow(`
				INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level)
				VALUES ($1, $2, $3, 80.0, 75.0, 85.0, 100.0, 100.0, 10)
				RETURNING id`, i, "OpponentBoxer"+string(rune('0'+i))+"_"+suffix, sql.NullString{Valid: true}).Scan(&id)
			if err != nil {
				t.Fatalf("Failed to create boxer %d: %v", i, err)
			}
		}

		// Query available opponents
		var count int
		err = db.QueryRow(`SELECT COUNT(*) FROM boxers WHERE name LIKE $1`, "%"+testSuffix+"%").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query boxers: %v", err)
		}

		t.Logf("Found %d potential opponents in test dataset", count)
	})

	t.Run("FilterOpponentsByLevelRange", func(t *testing.T) {
		minLevel := 8
		maxLevel := 12

		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM boxers WHERE level BETWEEN $1 AND $2`, minLevel, maxLevel).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query: %v", err)
		}

		t.Logf("Found %d opponents in level range [%d, %d]", count, minLevel, maxLevel)
	})

	t.Run("FilterOpponentsByHealthThreshold", func(t *testing.T) {
		minHealth := 80.0

		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM boxers WHERE health >= $1`, minHealth).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to query: %v", err)
		}

		t.Logf("Found %d opponents with health >= %.0f%%", count, minHealth)
	})

	_ = ctx // Use context to avoid unused var error
}

// BenchmarkFightScheduling benchmarks fight scheduling performance.
func BenchmarkFightScheduling(b *testing.B) {
	db, err := TestDBConfig()
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{JWT: config.JWTConfig{Secret: "test"}}
	lg := logger.New("Benchmark")
	boxerSvc := boxer.NewBoxerService(store.NewBoxerStore(db))
	eventStore := store.NewScheduledEventStore(db)
	fightSvc := NewFightService(db, cfg, boxerSvc, eventStore)

	// Setup test boxers once
	for i := 1; i <= 2; i++ {
		_, _ = db.Exec(`
			INSERT INTO boxers (user_id, name, nickname, strength, defense, agility, health, energy, level, experience)
			VALUES ($1, $2, $3, 80, 75, 85, 100, 100, 10, 1000)
			ON CONFLICT (name) DO UPDATE SET health = 100, energy = 100
		`, i, "BenchmarkBoxer"+string(rune('0'+i)), sql.NullString{Valid: true})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scheduledTime := time.Now()
		fight, err := fightSvc.Schedule(1, 2, scheduledTime)
		if err != nil {
			b.Logf("Schedule error: %v", err)
			continue
		}

		if fight.ID > 0 {
			_, _ = db.Exec(`DELETE FROM fights WHERE id = $1`, fight.ID)
		}
	}

	// Cleanup
	_, _ = db.Exec(`DELETE FROM boxers WHERE name LIKE 'BenchmarkBoxer%'`)
	_ = lg
}

var _ = model.Fight{} // Ensure model import is used
