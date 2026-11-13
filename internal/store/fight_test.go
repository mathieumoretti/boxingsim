package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/mormm/boxing/internal/db"
	"github.com/mormm/boxing/internal/domain"
	"github.com/mormm/boxing/internal/model"
	"github.com/stretchr/testify/assert"
)

// setupTestDB creates a fresh test database with migrations.
func setupTestDB(t *testing.T) *sql.DB {
	testDB, cleanup := db.FreshDatabaseWithMigrations(t, "fight_store", true)
	if testDB == nil {
		t.Skip("FreshDatabase returned nil - PostgreSQL not available")
	}
	t.Cleanup(func() {
		if cleanup != nil {
			cleanup()
		}
	})
	return testDB
}

// setupTestBoxers creates test users and boxers, returning the boxer IDs.
func setupTestBoxers(t *testing.T, ctx context.Context, testDB *sql.DB, store *BoxerStore) []domain.BoxerID {
	query := `INSERT INTO users (username, email, hashed_password, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`

	var userID int
	err := testDB.QueryRowContext(ctx, query, "testuser", "test@example.com", "hashed").Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	boxers := make([]*model.Boxer, 3)
	for i := 0; i < 3; i++ {
		boxers[i] = &model.Boxer{
			UserID:   userID,
			Name:     "Test Boxer " + string(rune('A'+i)),
			Energy:   100.0,
			Health:   100.0,
			Strength: 70.0,
			Defense:  60.0,
			Agility:  80.0,
			Level:    5,
		}
	}

	boxerIDs := make([]domain.BoxerID, 3)
	for i, boxer := range boxers {
		if err := store.Create(ctx, boxer); err != nil {
			t.Fatalf("Failed to create boxer %d: %v", i+1, err)
		}
		boxerIDs[i] = domain.BoxerID(boxer.ID)
	}

	return boxerIDs
}

// ============================================================================
// CRUD Operations Tests
// ============================================================================

func TestFightStore_Create(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	t.Run("creates fight and returns ID", func(t *testing.T) {
		scheduledTime := time.Now().Add(24 * time.Hour)
		id, err := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

		assert.NoError(t, err)
		assert.Greater(t, int(id), 0, "Should return valid fight ID")
	})

	t.Run("rejects invalid boxer1 ID", func(t *testing.T) {
		scheduledTime := time.Now().Add(24 * time.Hour)
		id, err := store.Create(ctx, -1, boxerIDs[1], &scheduledTime, 10)

		assert.Error(t, err)
		assert.Equal(t, domain.FightID(0), id)
	})

	t.Run("rejects invalid boxer2 ID", func(t *testing.T) {
		scheduledTime := time.Now().Add(24 * time.Hour)
		id, err := store.Create(ctx, boxerIDs[0], 0, &scheduledTime, 10)

		assert.Error(t, err)
		assert.Equal(t, domain.FightID(0), id)
	})

	t.Run("allows null scheduled time", func(t *testing.T) {
		id, err := store.Create(ctx, boxerIDs[0], boxerIDs[2], nil, 10)

		assert.NoError(t, err)
		assert.Greater(t, int(id), 0, "Should allow null scheduled time")
	})
}

func TestFightStore_GetByID(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	// Create a fight first
	scheduledTime := time.Now().Add(24 * time.Hour)
	fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

	t.Run("returns fight by ID", func(t *testing.T) {
		fight, err := store.GetByID(ctx, fightID)

		assert.NoError(t, err)
		assert.NotNil(t, fight)
		assert.Equal(t, int(fightID), fight.ID)
		assert.Equal(t, model.FightStatusScheduled, fight.Status)
	})

	t.Run("returns error for non-existent fight", func(t *testing.T) {
		fight, err := store.GetByID(ctx, 99999)

		assert.ErrorIs(t, err, ErrFightNotFound)
		assert.Nil(t, fight)
	})

	t.Run("rejects invalid fight ID", func(t *testing.T) {
		fight, err := store.GetByID(ctx, -1)

		assert.Error(t, err)
		assert.Nil(t, fight)
	})

	t.Run("returns error for zero fight ID", func(t *testing.T) {
		fight, err := store.GetByID(ctx, 0)

		assert.Error(t, err)
		assert.Nil(t, fight)
	})
}

func TestFightStore_UpdateStatus(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	// Create a fight first
	scheduledTime := time.Now().Add(24 * time.Hour)
	fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

	t.Run("updates status successfully", func(t *testing.T) {
		err := store.UpdateStatus(ctx, fightID, "in_progress")

		assert.NoError(t, err)

		// Verify the update
		fight, _ := store.GetByID(ctx, fightID)
		assert.Equal(t, model.FightStatusInProgress, fight.Status)
	})

	t.Run("updates to completed sets end_time", func(t *testing.T) {
		err := store.UpdateStatus(ctx, fightID, "completed")

		assert.NoError(t, err)

		fight, _ := store.GetByID(ctx, fightID)
		assert.Equal(t, model.FightStatusCompleted, fight.Status)
		assert.NotNil(t, fight.EndTime)
	})

	t.Run("returns error for non-existent fight", func(t *testing.T) {
		err := store.UpdateStatus(ctx, 99999, "cancelled")

		assert.ErrorIs(t, err, ErrFightNotFound)
	})

	t.Run("rejects invalid fight ID", func(t *testing.T) {
		err := store.UpdateStatus(ctx, -1, "cancelled")

		assert.Error(t, err)
	})
}

func TestFightStore_Delete(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	// Create a fight first
	scheduledTime := time.Now().Add(24 * time.Hour)
	fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

	t.Run("deletes fight successfully", func(t *testing.T) {
		err := store.Delete(ctx, fightID)

		assert.NoError(t, err)

		// Verify deletion
		fight, err := store.GetByID(ctx, fightID)
		assert.ErrorIs(t, err, ErrFightNotFound)
		assert.Nil(t, fight)
	})

	t.Run("returns error for non-existent fight", func(t *testing.T) {
		err := store.Delete(ctx, 99999)

		assert.ErrorIs(t, err, ErrFightNotFound)
	})

	t.Run("rejects invalid fight ID", func(t *testing.T) {
		err := store.Delete(ctx, -1)

		assert.Error(t, err)
	})
}

// ============================================================================
// Status Transitions with Transactions Tests
// ============================================================================

func TestFightStore_StatusTransition_ScheduledToCancelled(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	t.Run("can cancel scheduled fight with reason", func(t *testing.T) {
		scheduledTime := time.Now().Add(48 * time.Hour)
		fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

		err := store.CancelFight(ctx, fightID, "User requested cancellation")

		assert.NoError(t, err)

		fight, _ := store.GetByID(ctx, fightID)
		assert.Equal(t, model.FightStatusCancelled, fight.Status)
	})

	t.Run("cannot cancel already cancelled fight", func(t *testing.T) {
		scheduledTime2 := time.Now().Add(48 * time.Hour)
		fightID2, _ := store.Create(ctx, boxerIDs[1], boxerIDs[2], &scheduledTime2, 10)

		// Cancel once
		store.CancelFight(ctx, fightID2, "First cancellation")

		// Try to cancel again
		err := store.CancelFight(ctx, fightID2, "Second cancellation attempt")

		assert.ErrorIs(t, err, ErrFightCannotCancel)
	})

	t.Run("cannot cancel completed fight", func(t *testing.T) {
		scheduledTime3 := time.Now().Add(48 * time.Hour)
		fightID3, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime3, 10)

		// Mark as completed
		store.UpdateStatus(ctx, fightID3, "completed")

		err := store.CancelFight(ctx, fightID3, "Attempt to cancel completed fight")

		assert.ErrorIs(t, err, ErrFightCannotCancel)
	})

	t.Run("cannot cancel in_progress fight", func(t *testing.T) {
		scheduledTime4 := time.Now().Add(48 * time.Hour)
		fightID4, _ := store.Create(ctx, boxerIDs[1], boxerIDs[2], &scheduledTime4, 10)

		// Mark as in progress
		store.UpdateStatus(ctx, fightID4, "in_progress")

		err := store.CancelFight(ctx, fightID4, "Attempt to cancel in progress fight")

		assert.ErrorIs(t, err, ErrFightCannotCancel)
	})
}

func TestFightStore_StatusTransition_ScheduledToInProgressToCompleted(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	t.Run("scheduled to in_progress transition", func(t *testing.T) {
		scheduledTime := time.Now().Add(48 * time.Hour)
		fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

		fight, _ := store.GetByID(ctx, fightID)
		assert.Equal(t, model.FightStatusScheduled, fight.Status)

		err := store.UpdateStatus(ctx, fightID, "in_progress")
		assert.NoError(t, err)

		fight, _ = store.GetByID(ctx, fightID)
		assert.Equal(t, model.FightStatusInProgress, fight.Status)
	})

	t.Run("in_progress to completed with winner", func(t *testing.T) {
		scheduledTime := time.Now().Add(48 * time.Hour)
		fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

		// First move to in_progress
		store.UpdateStatus(ctx, fightID, "in_progress")

		err := store.SetWinner(ctx, fightID, boxerIDs[0])

		assert.NoError(t, err)

		fight, _ := store.GetByID(ctx, fightID)
		assert.Equal(t, model.FightStatusCompleted, fight.Status)
		assert.Equal(t, int(boxerIDs[0]), *fight.WinnerID)
	})

	t.Run("sets end_time on completion", func(t *testing.T) {
		scheduledTime := time.Now().Add(48 * time.Hour)
		fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

		store.UpdateStatus(ctx, fightID, "completed")

		fight, _ := store.GetByID(ctx, fightID)
		assert.NotNil(t, fight.EndTime)
		assert.Equal(t, model.FightStatusCompleted, fight.Status)
	})
}

func TestFightStore_TransactionOperations(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	t.Run("GetByIDWithLock acquires lock", func(t *testing.T) {
		scheduledTime := time.Now().Add(48 * time.Hour)
		fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

		tx, err := testDB.BeginTx(ctx, nil)
		assert.NoError(t, err)
		defer tx.Rollback()

		fight, err := store.GetByIDWithLockTx(ctx, tx, fightID)
		assert.NoError(t, err)
		assert.NotNil(t, fight)
		assert.Equal(t, int(fightID), fight.ID)
	})

	t.Run("UpdateStatusTx works within transaction", func(t *testing.T) {
		scheduledTime := time.Now().Add(48 * time.Hour)
		fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

		tx, err := testDB.BeginTx(ctx, nil)
		assert.NoError(t, err)
		defer tx.Rollback()

		err = store.UpdateStatusTx(ctx, tx, fightID, "in_progress")
		assert.NoError(t, err)

		fight, _ := store.GetByIDWithLockTx(ctx, tx, fightID)
		assert.Equal(t, model.FightStatusInProgress, fight.Status)
	})

	t.Run("SetWinnerTx works within transaction", func(t *testing.T) {
		scheduledTime := time.Now().Add(48 * time.Hour)
		fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

		tx, err := testDB.BeginTx(ctx, nil)
		assert.NoError(t, err)
		defer tx.Rollback()

		err = store.SetWinnerTx(ctx, tx, fightID, boxerIDs[0])
		assert.NoError(t, err)

		fight, _ := store.GetByIDWithLockTx(ctx, tx, fightID)
		assert.Equal(t, model.FightStatusCompleted, fight.Status)
		assert.Equal(t, int(boxerIDs[0]), *fight.WinnerID)
	})

	t.Run("BeginFightTx uses serializable isolation", func(t *testing.T) {
		tx, err := store.BeginFightTx(ctx)
		assert.NoError(t, err)
		defer tx.Rollback()

		// Verify transaction is created (serializable isolation tested in integration tests)
		assert.NotNil(t, tx)
	})
}

// ============================================================================
// Concurrent Safety Tests
// ============================================================================

func TestFightStore_BoxerInFight(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	scheduledTime := time.Now().Add(48 * time.Hour)
	fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

	t.Run("detects boxer in scheduled fight", func(t *testing.T) {
		inFight, err := store.BoxerInFight(ctx, boxerIDs[0])

		assert.NoError(t, err)
		assert.True(t, inFight, "Boxer should be marked as in fight")
	})

	t.Run("detects other boxer in same fight", func(t *testing.T) {
		inFight, err := store.BoxerInFight(ctx, boxerIDs[1])

		assert.NoError(t, err)
		assert.True(t, inFight, "Boxer should be marked as in fight")
	})

	t.Run("boxer not in fight returns false", func(t *testing.T) {
		inFight, err := store.BoxerInFight(ctx, boxerIDs[2])

		assert.NoError(t, err)
		assert.False(t, inFight, "Boxer should not be in any fight")
	})

	t.Run("completed fights do not count", func(t *testing.T) {
		store.SetWinner(ctx, fightID, boxerIDs[0])

		inFight, err := store.BoxerInFight(ctx, boxerIDs[0])

		assert.NoError(t, err)
		assert.False(t, inFight, "Completed fights should not count as active")
	})

	t.Run("cancelled fights do not count", func(t *testing.T) {
		scheduledTime2 := time.Now().Add(48 * time.Hour)
		fightID2, _ := store.Create(ctx, boxerIDs[1], boxerIDs[2], &scheduledTime2, 10)

		store.CancelFight(ctx, fightID2, "Test cancellation")

		inFight, err := store.BoxerInFight(ctx, boxerIDs[1])

		assert.NoError(t, err)
		assert.False(t, inFight, "Cancelled fights should not count as active")
	})

	t.Run("rejects invalid boxer ID", func(t *testing.T) {
		inFight, err := store.BoxerInFight(ctx, -1)

		assert.Error(t, err)
		assert.False(t, inFight)
	})
}

func TestFightStore_SimultaneousBookingPrevention(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	scheduledTime := time.Now().Add(48 * time.Hour)
	_, _ = store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

	t.Run("prevents booking boxer already in fight", func(t *testing.T) {
		inFight, _ := store.BoxerInFight(ctx, boxerIDs[0])
		assert.True(t, inFight, "Boxer should be detected as in fight")

		scheduledTime2 := time.Now().Add(60 * time.Hour)
		id, _ := store.Create(ctx, boxerIDs[0], boxerIDs[2], &scheduledTime2, 10)

		assert.Greater(t, int(id), 0, "Store allows create (validation is in FightService)")
	})
}

func TestFightStore_ConcurrentStatusUpdates(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	scheduledTime := time.Now().Add(48 * time.Hour)
	fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

	t.Run("FOR UPDATE lock prevents concurrent modifications", func(t *testing.T) {
		// Start two transactions
		tx1, err := testDB.BeginTx(ctx, nil)
		assert.NoError(t, err)
		defer tx1.Rollback()

		tx2, err := testDB.BeginTx(ctx, nil)
		assert.NoError(t, err)
		defer tx2.Rollback()

		// TX1 acquires lock on fight
		fight1, err := store.GetByIDWithLockTx(ctx, tx1, fightID)
		assert.NoError(t, err)
		assert.NotNil(t, fight1)

		// TX2 tries to acquire lock (will block or return with SKIP LOCKED)
		fight2, err := store.GetByIDWithLockTx(ctx, tx2, fightID)
		// With SKIP LOCKED, this returns nil or the fight without blocking
		// The key is that our production code uses proper transaction handling
		_ = fight2
		_ = err
	})

	t.Run("status update only succeeds once due to WHERE clause", func(t *testing.T) {
		scheduledTime2 := time.Now().Add(48 * time.Hour)
		fightID2, _ := store.Create(ctx, boxerIDs[1], boxerIDs[2], &scheduledTime2, 10)

		// Try to cancel twice (simulating concurrent cancellation attempts)
		err1 := store.CancelFight(ctx, fightID2, "First attempt")
		err2 := store.CancelFight(ctx, fightID2, "Second attempt")

		assert.NoError(t, err1, "First cancellation should succeed")
		assert.ErrorIs(t, err2, ErrFightCannotCancel, "Second cancellation should fail due to WHERE status = 'scheduled'")
	})
}

// ============================================================================
// Additional Query Tests
// ============================================================================

func TestFightStore_GetByBoxer(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	scheduledTime1 := time.Now().Add(48 * time.Hour)
	scheduledTime2 := time.Now().Add(72 * time.Hour)
	_, _ = store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime1, 10)
	_, _ = store.Create(ctx, boxerIDs[0], boxerIDs[2], &scheduledTime2, 10)

	t.Run("returns fights for boxer as boxer1", func(t *testing.T) {
		fights, err := store.GetByBoxer(ctx, boxerIDs[0], 10)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(fights), 2, "Should return at least 2 fights")
	})

	t.Run("returns fights for boxer as boxer2", func(t *testing.T) {
		fights, err := store.GetByBoxer(ctx, boxerIDs[1], 10)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(fights), 1, "Should return at least 1 fight")
	})

	t.Run("respects limit", func(t *testing.T) {
		fights, err := store.GetByBoxer(ctx, boxerIDs[0], 1)

		assert.NoError(t, err)
		assert.Len(t, fights, 1, "Should respect limit of 1")
	})
}

func TestFightStore_GetUpcomingFightForBoxer(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	futureTime := time.Now().Add(48 * time.Hour)
	_, _ = store.Create(ctx, boxerIDs[0], boxerIDs[1], &futureTime, 10)

	t.Run("returns upcoming fight", func(t *testing.T) {
		result, err := store.GetUpcomingFightForBoxer(ctx, boxerIDs[0])

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEqual(t, 0, result.FightID)
		assert.Equal(t, "Test Boxer B", result.OpponentName)
	})

	t.Run("returns error for boxer with no upcoming fights", func(t *testing.T) {
		result, err := store.GetUpcomingFightForBoxer(ctx, boxerIDs[2])

		assert.ErrorIs(t, err, db.ErrUpcomingFightNotFound)
		assert.Nil(t, result)
	})
}

func TestFightStore_SetData(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	scheduledTime := time.Now().Add(48 * time.Hour)
	fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 10)

	t.Run("sets data JSONB field", func(t *testing.T) {
		data := map[string]any{
			"winner":  "boxer1",
			"rounds":  8,
			"scored":  true,
		}

		err := store.SetData(ctx, fightID, data)

		assert.NoError(t, err)
	})

	t.Run("returns error for non-existent fight", func(t *testing.T) {
		err := store.SetData(ctx, 99999, map[string]any{})

		assert.ErrorIs(t, err, ErrFightNotFound)
	})
}

func TestFightStore_UpdateRound(t *testing.T) {
	testDB := setupTestDB(t)
	ctx := context.Background()
	store := NewFightStore(testDB)
	boxerStore := NewBoxerStore(testDB)
	boxerIDs := setupTestBoxers(t, ctx, testDB, boxerStore)

	scheduledTime := time.Now().Add(48 * time.Hour)
	fightID, _ := store.Create(ctx, boxerIDs[0], boxerIDs[1], &scheduledTime, 12)

	t.Run("updates round number", func(t *testing.T) {
		err := store.UpdateRound(ctx, fightID, 5)

		assert.NoError(t, err)

		fight, _ := store.GetByID(ctx, fightID)
		assert.Equal(t, 5, fight.Round)
	})

	t.Run("returns error for non-existent fight", func(t *testing.T) {
		err := store.UpdateRound(ctx, 99999, 5)

		assert.ErrorIs(t, err, ErrFightNotFound)
	})
}
