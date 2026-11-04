package domain

import (
	"context"
	"database/sql"
)

// IsolationLevel represents a database transaction isolation level.
type IsolationLevel int

const (
	// LevelDefault uses the database's default isolation level.
	LevelDefault IsolationLevel = iota
	// LevelReadUncommitted allows dirty reads (not supported by all databases).
	LevelReadUncommitted
	// LevelReadCommitted prevents dirty reads (PostgreSQL default).
	LevelReadCommitted
	// LevelRepeatableRead prevents dirty and non-repeatable reads.
	LevelRepeatableRead
	// LevelSerializable provides the highest isolation (recommended for complex transactions).
	LevelSerializable
)

// ToSQL converts IsolationLevel to sql.Level.
func (i IsolationLevel) ToSQL() sql.IsolationLevel {
	switch i {
	case LevelReadUncommitted:
		return sql.LevelReadUncommitted
	case LevelReadCommitted:
		return sql.LevelReadCommitted
	case LevelRepeatableRead:
		return sql.LevelRepeatableRead
	case LevelSerializable:
		return sql.LevelSerializable
	default:
		return sql.LevelDefault
	}
}

// TransactionOptions configures database transaction behavior.
type TransactionOptions struct {
	// Isolation is the isolation level for the transaction.
	// Defaults to LevelSerializable if not specified (best for complex business logic).
	Isolation IsolationLevel
	// ReadOnly indicates if the transaction is read-only.
	ReadOnly bool
}

// WithIsolation creates a TransactionOptions with the given isolation level.
func WithIsolation(level IsolationLevel) *TransactionOptions {
	return &TransactionOptions{
		Isolation: level,
	}
}

// WithSerializable returns default transaction options with serializable isolation.
// This is the recommended isolation level for complex business logic to prevent
// phantom reads and ensure maximum consistency.
func WithSerializable() *TransactionOptions {
	return &TransactionOptions{
		Isolation: LevelSerializable,
	}
}

// WithReadOnly returns transaction options for a read-only transaction.
// Useful for snapshot queries that need consistent reads.
func WithReadOnly() *TransactionOptions {
	return &TransactionOptions{
		ReadOnly: true,
	}
}

// TransactionFunc is the signature for functions that execute within a transaction.
// The function receives the transaction and should return an error if execution fails.
// If an error is returned, the transaction will be rolled back.
type TransactionFunc func(tx *sql.Tx) error

// ExecInTransaction executes a function within a database transaction with automatic rollback/commit.
//
// Behavior:
//   - Begins a transaction with the specified options
//   - Executes fn(tx) within the transaction
//   - Commits if fn returns nil
//   - Rolls back if fn returns an error or panics
//   - Returns any error from fn or transaction operations
//
// Usage example:
//
//	err := domain.ExecInTransaction(db, func(tx *sql.Tx) error {
//		// Lock and fetch resources
//		boxer, err := boxerStore.GetByIDWithLockTx(ctx, tx, boxerID)
//		if err != nil { return err }
//
//		// Perform business logic
//		// ...
//
//		// Update state
//		return boxerStore.UpdateTx(ctx, tx, boxer)
//	}, domain.WithSerializable())
func ExecInTransaction(db *sql.DB, fn TransactionFunc, opts *TransactionOptions) error {
	if opts == nil {
		opts = WithSerializable() // Default to serializable for safety
	}

	txOptions := &sql.TxOptions{
		Isolation: opts.Isolation.ToSQL(),
		ReadOnly:  opts.ReadOnly,
	}

	tx, err := db.BeginTx(context.Background(), txOptions)
	if err != nil {
		return err
	}

	// Ensure rollback on error or panic, safe to call even after commit
	defer func() {
		if err := recover(); err != nil {
			_ = tx.Rollback() // Ignore rollback error, panic will continue
			panic(err)
		}
	}()

	// Execute the function within the transaction
	if err := fn(tx); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return err // Return original error, rollback error logged internally
		}
		return err
	}

	// Commit the transaction
	return tx.Commit()
}

// ExecInTransactionWithContext executes a function within a database transaction with context support.
// This is useful for transactions that need to respect deadlines or cancellation.
//
// The context is used for:
//   - Transaction lifecycle (BeginTx, Rollback, Commit)
//   - Passed to the function for use in store operations
//
// Usage example:
//
//	ctx := context.Background()
//	err := domain.ExecInTransactionWithContext(ctx, db, func(ctx context.Context, tx *sql.Tx) error {
//		boxer, err := boxerStore.GetByIDWithLockTx(ctx, tx, boxerID)
//		if err != nil { return err }
//		return boxerStore.UpdateTx(ctx, tx, boxer)
//	}, domain.WithSerializable())
type TransactionFuncWithContext func(ctx context.Context, tx *sql.Tx) error

func ExecInTransactionWithContext(ctx context.Context, db *sql.DB, fn TransactionFuncWithContext, opts *TransactionOptions) error {
	if opts == nil {
		opts = WithSerializable() // Default to serializable for safety
	}

	txOptions := &sql.TxOptions{
		Isolation: opts.Isolation.ToSQL(),
		ReadOnly:  opts.ReadOnly,
	}

	tx, err := db.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}

	// Ensure rollback on error or panic, safe to call even after commit
	defer func() {
		if err := recover(); err != nil {
			_ = tx.Rollback() // Ignore rollback error, panic will continue
			panic(err)
		}
	}()

	// Execute the function within the transaction
	if err := fn(ctx, tx); err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return err // Return original error, rollback error logged internally
		}
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
