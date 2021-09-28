package repositories

import (
	"context"
	"gorm.io/gorm"
)

const transactionKey = "transactionKey"

type Repository interface {
	UseTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type postgresqlRepository struct {
	db *gorm.DB
}

func (r *postgresqlRepository) UseTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, hasTx := ctx.Value(transactionKey).(*gorm.DB)
	if !hasTx {
		tx = r.db.Begin()
		ctx = context.WithValue(ctx, transactionKey, tx)
	}

	if err := fn(ctx); err != nil {
		tx.Rollback()
		return err
	}

	if !hasTx {
		return tx.Commit().Error
	}

	return nil
}

func (r *postgresqlRepository) GetTX(ctx context.Context) *gorm.DB {
	tx, hasTx := ctx.Value(transactionKey).(*gorm.DB)
	if hasTx {
		return tx
	}
	return r.db
}
