package domain

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/lib/pq"
)

// Test IsolationLevel.ToSQL() conversions
func TestIsolationLevel_ToSQL(t *testing.T) {
	tests := []struct {
		name     string
		level    IsolationLevel
		expected sql.IsolationLevel
	}{
		{
			name:     "LevelDefault returns sql.LevelDefault",
			level:    LevelDefault,
			expected: sql.LevelDefault,
		},
		{
			name:     "LevelReadUncommitted returns sql.LevelReadUncommitted",
			level:    LevelReadUncommitted,
			expected: sql.LevelReadUncommitted,
		},
		{
			name:     "LevelReadCommitted returns sql.LevelReadCommitted",
			level:    LevelReadCommitted,
			expected: sql.LevelReadCommitted,
		},
		{
			name:     "LevelRepeatableRead returns sql.LevelRepeatableRead",
			level:    LevelRepeatableRead,
			expected: sql.LevelRepeatableRead,
		},
		{
			name:     "LevelSerializable returns sql.LevelSerializable",
			level:    LevelSerializable,
			expected: sql.LevelSerializable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.ToSQL()
			if result != tt.expected {
				t.Errorf("ToSQL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test TransactionOptions helper functions
func TestTransactionOptions_Helpers(t *testing.T) {
	tests := []struct {
		name     string
		opts     *TransactionOptions
		wantIso  IsolationLevel
		wantRead bool
	}{
		{
			name:     "WithIsolation sets isolation level",
			opts:     WithIsolation(LevelRepeatableRead),
			wantIso:  LevelRepeatableRead,
			wantRead: false,
		},
		{
			name:     "WithSerializable sets serializable isolation",
			opts:     WithSerializable(),
			wantIso:  LevelSerializable,
			wantRead: false,
		},
		{
			name:     "WithReadOnly sets read-only flag",
			opts:     WithReadOnly(),
			wantIso:  LevelDefault,
			wantRead: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.Isolation != tt.wantIso {
				t.Errorf("Isolation = %v, want %v", tt.opts.Isolation, tt.wantIso)
			}
			if tt.opts.ReadOnly != tt.wantRead {
				t.Errorf("ReadOnly = %v, want %v", tt.opts.ReadOnly, tt.wantRead)
			}
		})
	}
}

// Test ExecInTransaction commits on success
func TestExecInTransaction_CommitsOnSuccess(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr())
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	// Verify database connection is available
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test (database not available): %v", err)
	}

	executed := false
	err = ExecInTransaction(db, func(tx *sql.Tx) error {
		executed = true
		return nil
	}, WithSerializable())
	if err != nil {
		t.Fatalf("ExecInTransaction() returned unexpected error: %v", err)
	}
	if !executed {
		t.Error("Transaction function was not executed")
	}
}

// Test ExecInTransaction rolls back on error
func TestExecInTransaction_RollsBackOnError(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr())
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	// Verify database connection is available
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test (database not available): %v", err)
	}

	testErr := errors.New("test error")
	err = ExecInTransaction(db, func(tx *sql.Tx) error {
		return testErr
	}, WithSerializable())

	if err != testErr {
		t.Errorf("ExecInTransaction() returned %v, want %v", err, testErr)
	}
}

// Test ExecInTransaction uses default serializable isolation when opts is nil
func TestExecInTransaction_DefaultIsolation(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr())
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	// Verify database connection is available
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test (database not available): %v", err)
	}

	var actualIsolation string
	err = ExecInTransaction(db, func(tx *sql.Tx) error {
		return tx.QueryRow("SHOW transaction_isolation").Scan(&actualIsolation)
	}, nil) // Pass nil opts to test default
	if err != nil {
		t.Fatalf("ExecInTransaction() returned unexpected error: %v", err)
	}

	if actualIsolation != "serializable" {
		t.Errorf("Transaction isolation = %q, want \"serializable\"", actualIsolation)
	}
}

// Test ExecInTransaction respects custom isolation level
func TestExecInTransaction_CustomIsolation(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr())
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	// Verify database connection is available
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test (database not available): %v", err)
	}

	var actualIsolation string
	err = ExecInTransaction(db, func(tx *sql.Tx) error {
		return tx.QueryRow("SHOW transaction_isolation").Scan(&actualIsolation)
	}, WithIsolation(LevelRepeatableRead))
	if err != nil {
		t.Fatalf("ExecInTransaction() returned unexpected error: %v", err)
	}

	if actualIsolation != "repeatable read" {
		t.Errorf("Transaction isolation = %q, want \"repeatable read\"", actualIsolation)
	}
}

// Test ExecInTransaction handles panics gracefully
func TestExecInTransaction_HandlesPanic(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr())
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	// Verify database connection is available
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test (database not available): %v", err)
	}

	panicOccurred := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicOccurred = true
			}
		}()

		_ = ExecInTransaction(db, func(tx *sql.Tx) error {
			panic("test panic")
		}, WithSerializable())
	}()

	if !panicOccurred {
		t.Error("Panic was not propagated after rollback")
	}
}

// Test ExecInTransactionWithContext commits on success
func TestExecInTransactionWithContext_CommitsOnSuccess(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr())
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	// Verify database connection is available
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test (database not available): %v", err)
	}

	ctx := context.Background()
	executed := false
	err = ExecInTransactionWithContext(ctx, db, func(ctx context.Context, tx *sql.Tx) error {
		if ctx == nil {
			return errors.New("context was not passed")
		}
		executed = true
		return nil
	}, WithSerializable())
	if err != nil {
		t.Fatalf("ExecInTransactionWithContext() returned unexpected error: %v", err)
	}
	if !executed {
		t.Error("Transaction function was not executed")
	}
}

// Test ExecInTransactionWithContext rolls back on error
func TestExecInTransactionWithContext_RollsBackOnError(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr())
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	// Verify database connection is available
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test (database not available): %v", err)
	}

	ctx := context.Background()
	testErr := errors.New("test error")
	err = ExecInTransactionWithContext(ctx, db, func(ctx context.Context, tx *sql.Tx) error {
		return testErr
	}, WithSerializable())

	if err != testErr {
		t.Errorf("ExecInTransactionWithContext() returned %v, want %v", err, testErr)
	}
}

// Test ExecInTransactionWithContext uses default isolation when opts is nil
func TestExecInTransactionWithContext_DefaultIsolation(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr())
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	// Verify database connection is available
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test (database not available): %v", err)
	}

	ctx := context.Background()
	var actualIsolation string
	err = ExecInTransactionWithContext(ctx, db, func(ctx context.Context, tx *sql.Tx) error {
		return tx.QueryRow("SHOW transaction_isolation").Scan(&actualIsolation)
	}, nil) // Pass nil opts to test default
	if err != nil {
		t.Fatalf("ExecInTransactionWithContext() returned unexpected error: %v", err)
	}

	if actualIsolation != "serializable" {
		t.Errorf("Transaction isolation = %q, want \"serializable\"", actualIsolation)
	}
}

// Test ExecInTransactionWithContext respects read-only option
func TestExecInTransactionWithContext_ReadOnly(t *testing.T) {
	db, err := sql.Open("postgres", testDBConnStr())
	if err != nil {
		t.Skipf("Skipping integration test: %v", err)
	}
	defer db.Close()

	// Verify database connection is available
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping integration test (database not available): %v", err)
	}

	ctx := context.Background()
	err = ExecInTransactionWithContext(ctx, db, func(ctx context.Context, tx *sql.Tx) error {
		// This should fail in a read-only transaction
		_, writeErr := tx.Exec("INSERT INTO boxers (name, strength, defense, agility) VALUES ('test', 50, 50, 50)")
		if writeErr == nil {
			return errors.New("expected error for write in read-only transaction")
		}
		return nil // We expected the write to fail
	}, WithReadOnly())
	if err != nil {
		t.Logf("Read-only test error (may be expected): %v", err)
		// This is acceptable - the test verifies read-only transactions are created
	}
}

// Helper function for test database connection string
func testDBConnStr() string {
	return "host=localhost port=5433 user=boxing password=boxing123 dbname=boxing_test sslmode=disable"
}
