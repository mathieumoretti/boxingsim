package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Advisory lock key for world simulation worker authority.
// Using a large integer to avoid collisions with other advisory locks.
// Must fit within PostgreSQL oid (32-bit unsigned) range for pg_locks table queries.
// Maximum oid value is 4294967295 (~4 billion).
// Key is derived from: 0x0000_0001_0000_0001 (fits in 32-bit, still large)
const WorkerAuthorityLockKey = 16777217

var (
	// ErrLockAcquisitionFailed is returned when unable to acquire advisory lock
	ErrLockAcquisitionFailed = fmt.Errorf("failed to acquire worker authority lock")

	// ErrLockNotHeld is returned when trying to release a lock not held by this connection
	ErrLockNotHeld = fmt.Errorf("worker authority lock is not held by this connection")
)

// LockAcquirer provides PostgreSQL advisory lock utilities for worker authority control.
//
// PostgreSQL advisory locks are application-managed locks that are independent of
// database objects (tables, rows). They are identified by a 64-bit integer key and
// are automatically released when the connection closes or the transaction ends
// (for xact locks).
//
// Reference: https://www.postgresql.org/docs/current/functions-admin.html#FUNCTIONS-ADVISORY-LOCKS
type LockAcquirer struct {
	db *sql.DB
}

// NewLockAcquirer creates a new lock acquirer for the given database connection.
func NewLockAcquirer(db *sql.DB) *LockAcquirer {
	return &LockAcquirer{db: db}
}

// AcquireWorkerAuthority attempts to acquire the worker authority advisory lock.
// Uses pg_advisory_lock which:
//   - Blocks until lock is available (for exclusive authority)
//   - Is session-scoped (not transaction-scoped) for worker continuity
//   - Automatically releases when connection closes
func (la *LockAcquirer) AcquireWorkerAuthority(ctx context.Context) error {
	_, err := la.db.ExecContext(ctx, "SELECT pg_advisory_lock($1)", WorkerAuthorityLockKey)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrLockAcquisitionFailed, err)
	}
	return nil
}

// ReleaseWorkerAuthority releases the worker authority advisory lock.
// Uses pg_advisory_unlock which immediately releases the lock if held by this connection.
func (la *LockAcquirer) ReleaseWorkerAuthority(ctx context.Context) error {
	var released bool
	err := la.db.QueryRowContext(ctx, "SELECT pg_advisory_unlock($1)", WorkerAuthorityLockKey).Scan(&released)
	if err != nil {
		return fmt.Errorf("failed to release worker authority lock: %v", err)
	}

	if !released {
		return ErrLockNotHeld
	}

	return nil
}

// TryAcquireWorkerAuthority attempts to acquire the lock without blocking.
// Returns true if acquired, false if another worker holds it.
// Uses pg_try_advisory_lock which returns immediately.
func (la *LockAcquirer) TryAcquireWorkerAuthority(ctx context.Context) (bool, error) {
	var acquired bool
	err := la.db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", WorkerAuthorityLockKey).Scan(&acquired)
	if err != nil {
		return false, fmt.Errorf("failed to try acquiring lock: %v", err)
	}
	return acquired, nil
}

// IsWorkerAuthorityHeld checks if this connection holds the worker authority lock.
// Queries pg_locks system view to check for advisory locks held by the current backend pid.
func (la *LockAcquirer) IsWorkerAuthorityHeld(ctx context.Context) (bool, error) {
	var count int
	err := la.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM pg_locks WHERE locktype = 'advisory' AND pid = pg_backend_pid() AND objid = $1",
		WorkerAuthorityLockKey,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check lock status: %v", err)
	}
	return count > 0, nil
}
