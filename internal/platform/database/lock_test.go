//go:build integration

package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/mormm/boxing/internal/platform/config"
)

func setupLockTestDB(t *testing.T) *sql.DB {
	testCfg, err := config.LoadTestDBConfig()
	if err != nil {
		t.Fatalf("Failed to load test DB config: %v", err)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=boxing sslmode=disable",
		testCfg.Host,
		testCfg.Port,
		testCfg.User,
		testCfg.Password,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("Failed to ping database: %v", err)
	}

	return db
}

func TestAcquireWorkerAuthority(t *testing.T) {
	db := setupLockTestDB(t)
	defer db.Close()

	acquirer := NewLockAcquirer(db)
	ctx := context.Background()

	// First acquisition should succeed
	err := acquirer.AcquireWorkerAuthority(ctx)
	if err != nil {
		t.Errorf("First lock acquisition should succeed, got error: %v", err)
	}

	// Verify lock is held
	held, err := acquirer.IsWorkerAuthorityHeld(ctx)
	if err != nil {
		t.Fatalf("Failed to check lock status: %v", err)
	}
	if !held {
		t.Error("Lock should be held after AcquireWorkerAuthority")
	}

	// Release the lock
	err = acquirer.ReleaseWorkerAuthority(ctx)
	if err != nil {
		t.Errorf("Failed to release lock: %v", err)
	}
}

func TestTryAcquireWorkerAuthority(t *testing.T) {
	db := setupLockTestDB(t)
	defer db.Close()

	acquirer1 := NewLockAcquirer(db)
	ctx := context.Background()

	// First try-acquire should succeed
	acquired, err := acquirer1.TryAcquireWorkerAuthority(ctx)
	if err != nil {
		t.Fatalf("TryAcquire failed: %v", err)
	}
	if !acquired {
		t.Error("First try-acquire should succeed")
	}

	// Create second acquirer on same connection (will also succeed since it's per-connection)
	acquirer2 := NewLockAcquirer(db)
	acquired2, err := acquirer2.TryAcquireWorkerAuthority(ctx)
	if err != nil {
		t.Fatalf("Second TryAcquire failed: %v", err)
	}
	// Same connection can acquire multiple times (ref count)
	if !acquired2 {
		t.Error("Second try-acquire on same connection should succeed")
	}

	// Release both
	acquirer1.ReleaseWorkerAuthority(ctx)
	acquirer2.ReleaseWorkerAuthority(ctx)
}

func TestReleaseWorkerAuthority_NotHeld(t *testing.T) {
	db := setupLockTestDB(t)
	defer db.Close()

	acquirer := NewLockAcquirer(db)
	ctx := context.Background()

	// Try to release without acquiring
	err := acquirer.ReleaseWorkerAuthority(ctx)
	if err != ErrLockNotHeld {
		t.Errorf("Expected ErrLockNotHeld, got: %v", err)
	}
}
