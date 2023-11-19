package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/scan"
)

const OrderASC = "ASC"
const OrderDESC = "DESC"

type Database struct {
	bob.DB
	EnableDebugLogging bool
}

type Executor interface {
	scan.Queryer
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func (db *Database) InTransaction(ctx context.Context, fn func(ctx context.Context, tx Executor) error) error {
	if tx, ok := txFromCtx(ctx); ok {
		var exec Executor = tx
		if db.EnableDebugLogging {
			exec = bob.Debug(tx)
		}
		return fn(ctx, exec)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	ctx = ctxWithTx(ctx, tx)

	var exec Executor = tx
	if db.EnableDebugLogging {
		exec = bob.Debug(tx)
	}

	_, err = exec.ExecContext(ctx, "PRAGMA defer_foreign_keys = 1")
	if err != nil {
		err = fmt.Errorf("error setting foreign key check to deferred: %w", err)
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back: %w. original error: %v", rbErr, err)
		}
		return err
	}

	if err := fn(ctx, exec); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back: %w. original error: %v", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func InTransaction[R any](ctx context.Context, db *Database, fn func(ctx context.Context, tx Executor) (R, error)) (R, error) {
	var result R
	err := db.InTransaction(ctx, func(ctx context.Context, tx Executor) error {
		r, err := fn(ctx, tx)
		if err != nil {
			return err
		}

		result = r

		return nil
	})

	if err != nil {
		return result, err
	}

	return result, nil
}

type ctxTxKeyType string

const ctxTxKey = ctxTxKeyType("ctxTxKey")

func txFromCtx(ctx context.Context) (bob.Tx, bool) {
	tx, ok := ctx.Value(ctxTxKey).(bob.Tx)
	return tx, ok
}

func ctxWithTx(parent context.Context, tx bob.Tx) context.Context {
	return context.WithValue(parent, ctxTxKey, tx)
}
