package db

import (
	"context"
	"fmt"
	"golang.org/x/exp/slog"

	"github.com/jackc/pgx/v5"
)

// A TxFn is a function that will be called with an initialized `Transaction` object
// that can be used for executing statements and queries against a database.
type TxFn func(tx pgx.Tx) error

// WithTransaction creates a new transaction and handles rollback/commit based on the
// error object returned by the `TxFn` or when it panics.
func WithTransaction(ctx context.Context, fn TxFn) error {
	tx, beginErr := Pool.Begin(ctx)
	if beginErr != nil {
		slog.WarnContext(ctx, "cannot begin db transaction")
		return fmt.Errorf("tx error: %w", beginErr)
	}

	defer func() {
		if p := recover(); p != nil {
			slog.WarnContext(ctx, "db panic (rollback)", "panic", p)
			rollErr := tx.Rollback(ctx)
			if rollErr != nil {
				slog.WarnContext(ctx, "cannot rollback db transaction")
				return
			}
			panic(p)
		}
	}()

	callErr := fn(tx)

	if callErr != nil {
		slog.WarnContext(ctx, "db error (rollback)", "err", callErr)
		rollErr := tx.Rollback(ctx)
		if rollErr != nil {
			slog.WarnContext(ctx, "cannot rollback db transaction)", "err", rollErr)
			// return the call (root cause) error and not transaction error
			return fmt.Errorf("tx rollback error: %s, cause: %w", rollErr.Error(), callErr)
		}
		return fmt.Errorf("tx error: %w", callErr)
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		slog.WarnContext(ctx, "cannot commit db transaction)", "err", commitErr)
		return fmt.Errorf("db commit error: %w", commitErr)
	}

	return nil
}
