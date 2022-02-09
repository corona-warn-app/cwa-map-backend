/*
 *   Corona-Warn-App / cwa-map-backend
 *
 *   (C) 2020, T-Systems International GmbH
 *
 *   Deutsche Telekom AG and all other contributors /
 *   copyright owners license this file to you under the Apache
 *   License, Version 2.0 (the "License"); you may not use this
 *   file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing,
 *   software distributed under the License is distributed on an
 *   "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *   KIND, either express or implied.  See the License for the
 *   specific language governing permissions and limitations
 *   under the License.
 */

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
