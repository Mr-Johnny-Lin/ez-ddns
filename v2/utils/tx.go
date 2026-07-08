package utils

import (
	"context"
	"database/sql"
)

type TxPropagation int

const (
	PropagationRequired     TxPropagation = iota
	PropagationRequiresNew  TxPropagation = iota
	PropagationSupports     TxPropagation = iota
	PropagationNotSupported TxPropagation = iota
	PropagationMandatory    TxPropagation = iota
	PropagationNever        TxPropagation = iota
)

type txContext struct {
	tx          *sql.Tx
	isNested    bool
	propagation TxPropagation
}

type txKey struct{}

type propagationKey struct{}

func ContextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, &txContext{
		tx:       tx,
		isNested: false,
	})
}

func ContextWithTxAndPropagation(ctx context.Context, tx *sql.Tx, propagation TxPropagation) context.Context {
	return context.WithValue(ctx, txKey{}, &txContext{
		tx:          tx,
		isNested:    false,
		propagation: propagation,
	})
}

func ContextWithPropagation(ctx context.Context, propagation TxPropagation) context.Context {
	return context.WithValue(ctx, propagationKey{}, propagation)
}

func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	txCtx, ok := ctx.Value(txKey{}).(*txContext)
	if !ok || txCtx == nil {
		return nil, false
	}
	return txCtx.tx, true
}

func MustTxFromContext(ctx context.Context) *sql.Tx {
	tx, ok := TxFromContext(ctx)
	if !ok {
		panic("no transaction found in context")
	}
	return tx
}

func IsTxNested(ctx context.Context) bool {
	txCtx, ok := ctx.Value(txKey{}).(*txContext)
	if !ok || txCtx == nil {
		return false
	}
	return txCtx.isNested
}

type TxExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func GetExecutor(ctx context.Context, db *sql.DB) TxExecutor {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return db
}

func BeginTxWithPropagation(ctx context.Context, db *sql.DB) (context.Context, *sql.Tx, func(error) error, error) {
	propagation := PropagationRequired
	if p, ok := ctx.Value(propagationKey{}).(TxPropagation); ok {
		propagation = p
	}

	switch propagation {
	case PropagationRequired:
		if existingTx, ok := TxFromContext(ctx); ok {
			return ctx, existingTx, func(error) error { return nil }, nil
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ctx, nil, nil, err
		}
		newCtx := context.WithValue(ctx, txKey{}, &txContext{
			tx:       tx,
			isNested: false,
		})
		return newCtx, tx, func(err error) error {
			if err != nil {
				return tx.Rollback()
			}
			return tx.Commit()
		}, nil

	case PropagationRequiresNew:
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ctx, nil, nil, err
		}
		newCtx := context.WithValue(ctx, txKey{}, &txContext{
			tx:       tx,
			isNested: true,
		})
		return newCtx, tx, func(err error) error {
			if err != nil {
				return tx.Rollback()
			}
			return tx.Commit()
		}, nil

	case PropagationSupports:
		if existingTx, ok := TxFromContext(ctx); ok {
			return ctx, existingTx, func(error) error { return nil }, nil
		}
		return ctx, nil, func(error) error { return nil }, nil

	case PropagationNotSupported:
		return ctx, nil, func(error) error { return nil }, nil

	case PropagationMandatory:
		if existingTx, ok := TxFromContext(ctx); ok {
			return ctx, existingTx, func(error) error { return nil }, nil
		}
		return ctx, nil, nil, ErrNoTransaction

	case PropagationNever:
		if _, ok := TxFromContext(ctx); ok {
			return ctx, nil, nil, ErrTransactionExists
		}
		return ctx, nil, func(error) error { return nil }, nil

	default:
		return ctx, nil, nil, ErrInvalidPropagation
	}
}

var (
	ErrNoTransaction      = &txError{"no transaction found"}
	ErrTransactionExists  = &txError{"transaction already exists"}
	ErrInvalidPropagation = &txError{"invalid transaction propagation"}
)

type txError struct {
	msg string
}

func (e *txError) Error() string {
	return e.msg
}
