package database

import (
	"context"
	"fmt"

	"github.com/stephenafamo/bob"
)

const OrderASC = "ASC"
const OrderDESC = "DESC"

type Database struct {
	bob.DB
}

func (db *Database) InTransaction(ctx context.Context, fn func(ctx context.Context, tx bob.Tx) error) error {
	if tx, ok := txFromCtx(ctx); ok {
		return fn(ctx, tx)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	ctx = ctxWithTx(ctx, tx)

	if err := fn(ctx, tx); err != nil {
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

func InTransaction[R any](ctx context.Context, db *Database, fn func(ctx context.Context, tx bob.Tx) (R, error)) (R, error) {
	var result R
	err := db.InTransaction(ctx, func(ctx context.Context, tx bob.Tx) error {
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
