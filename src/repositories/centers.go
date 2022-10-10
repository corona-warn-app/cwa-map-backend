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
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/domain"
	"context"
	"errors"
	"fmt"
	"github.com/doug-martin/goqu"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

const DistanceUnit = 111.045

type SearchParameters struct {
	Appointment     *domain.AppointmentType
	TestKind        *domain.TestKind
	DCC             *bool
	IncludeOutdated *bool
}

type PagedCentersResult struct {
	PagedResult
	Result []domain.Center
}

type CenterStatistics struct {
	TotalCount     int
	DccCount       int
	InvisibleCount int
}

type Centers interface {
	Repository
	FindByUUID(ctx context.Context, uuid string) (domain.Center, error)
	Delete(ctx context.Context, center domain.Center) error

	FindByBounds(ctx context.Context, target domain.Bounds, params SearchParameters, limit uint) ([]domain.Center, error)

	// FindByOperatorAndUserReference find the center for the given operator und user reference
	FindByOperatorAndUserReference(ctx context.Context, operator, number string) (domain.Center, error)

	FindByOperator(ctx context.Context, operator string, search string, page PageRequest) (PagedCentersResult, error)

	// Save persists the given center
	Save(ctx context.Context, center *domain.Center) error

	// SaveMultiple saves the given centers
	SaveMultiple(ctx context.Context, center []domain.Center) ([]domain.Center, error)

	DeleteByOperator(ctx context.Context, operator string) error

	FindAll() ([]domain.Center, error)

	FindStatistics(ctx context.Context) (CenterStatistics, error)

	FindCentersForNotification(ctx context.Context, lastUpdateAge, renotifyInterval int) ([]domain.Center, error)
}

type centersRepository struct {
	postgresqlRepository
}

func NewCentersRepository(db *gorm.DB) Centers {
	return &centersRepository{
		postgresqlRepository{db: db},
	}
}

func (r *centersRepository) FindAll() ([]domain.Center, error) {
	var centers []domain.Center
	err := r.db.
		Preload("Operator").
		Order("operator_uuid, uuid").
		Find(&centers).Error
	return centers, err
}

func (r *centersRepository) Delete(ctx context.Context, center domain.Center) error {
	return r.db.Delete(&center).Error
}

func (r *centersRepository) FindByUUID(ctx context.Context, uuid string) (domain.Center, error) {
	var center domain.Center
	err := r.GetTX(ctx).Model(&domain.Center{}).
		Preload("Operator").
		Where("uuid = ?", uuid).
		First(&center).Error
	return center, err
}

func (r *centersRepository) DeleteByOperator(ctx context.Context, operator string) error {
	return r.GetTX(ctx).Exec("DELETE FROM centers WHERE operator_uuid = ?", operator).Error
}

func (r *centersRepository) Save(ctx context.Context, center *domain.Center) error {
	if util.IsNilOrEmpty(&center.UUID) {
		if newUUID, err := uuid.NewUUID(); err == nil {
			center.UUID = newUUID.String()
		} else {
			return err
		}

	} else {
		if _, err := r.FindByUUID(ctx, center.UUID); errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
	}

	return r.GetTX(ctx).Save(center).Error
}

func (r *centersRepository) SaveMultiple(ctx context.Context, centers []domain.Center) ([]domain.Center, error) {
	result := make([]domain.Center, len(centers))
	err := r.db.Transaction(func(tx *gorm.DB) error {
		repo := NewCentersRepository(tx)
		for i, center := range centers {
			if err := repo.Save(ctx, &center); err != nil {
				return err
			} else {
				result[i] = center
			}
		}
		return nil
	})

	return result, err
}

func (r *centersRepository) FindByOperator(ctx context.Context, operator string, search string, page PageRequest) (PagedCentersResult, error) {
	baseQuery := r.db.Model(&domain.Center{}).
		Where("operator_uuid = ?", operator)

	if search != "" {
		baseQuery.Where("name ilike ? or address ilike ?", "%"+search+"%", "%"+search+"%")
	}
	baseQuery = baseQuery.Order("user_reference")

	result := PagedCentersResult{}
	if err := baseQuery.Count(&result.Count).Error; err != nil {
		return result, err
	}

	err := baseQuery.
		Offset(page.Page * page.Size).
		Limit(page.Size).
		Find(&result.Result).
		Error

	return result, err
}

func (r *centersRepository) Update(ctx context.Context, center domain.Center) (domain.Center, error) {
	err := r.db.Save(&center).Error
	return center, err
}

// FindByBounds finds centers within the given bounds
func (r *centersRepository) FindByBounds(ctx context.Context, target domain.Bounds, params SearchParameters, limit uint) ([]domain.Center, error) {
	// build base query restriction
	builder := goqu.From("centers").Where(
		goqu.I("latitude").Between(goqu.RangeVal{
			Start: target.SouthWest.Latitude,
			End:   target.NorthEast.Latitude,
		}),
		goqu.I("longitude").Between(goqu.RangeVal{
			Start: target.SouthWest.Longitude,
			End:   target.NorthEast.Longitude,
		}),
		goqu.Or(
			goqu.I("enter_date").IsNull(),
			goqu.I("enter_date").Lte(time.Now()),
		),
		goqu.Or(
			goqu.I("leave_date").IsNull(),
			goqu.I("leave_date").Gte(time.Now()),
		),
		goqu.I("visible").IsNotFalse(),
	)

	if params.DCC != nil && *params.DCC {
		builder = builder.Where(goqu.I("dcc").Eq(true))
	}
	if params.Appointment != nil {
		builder = builder.Where(goqu.I("appointment").Eq(*params.Appointment))
	}
	if params.TestKind != nil {
		builder = builder.Where(goqu.L("test_kinds @> ARRAY[?]::varchar[]", *params.TestKind))
	}
	if params.IncludeOutdated == nil || *params.IncludeOutdated == false {
		builder = builder.Where(goqu.L("last_update > now() - INTERVAL '4 weeks'"))
	}

	countQuery := builder.Select(goqu.COUNT("*"))
	var count uint
	if sql, args, err := countQuery.ToSql(); err == nil {
		if err := r.db.Raw(sql, args).Scan(&count).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	var result []domain.Center
	if count == 0 {
		return result, nil
	}

	fraction := float64(limit) / float64(count)
	resultsQuery := builder.
		Where(goqu.I("ranking").Lte(fraction)).
		Select("centers.*"). //, goqu.L("haversine(latitude, longitude, ?, ?)", target.Latitude, target.Longitude).As("distance")).
		Order(goqu.I("centers.uuid").Asc())

	sql, args, err := resultsQuery.ToSql()
	tx := r.db.Raw(sql, args...)
	err = tx.Preload("Operator").
		Find(&result).
		Error

	return result, err
}

func (r *centersRepository) FindByOperatorAndUserReference(ctx context.Context, operator, userReference string) (domain.Center, error) {
	var center domain.Center
	err := r.GetTX(ctx).Where("operator_uuid = ? and user_reference = ?", operator, userReference).First(&center).Error
	return center, err
}

func (r *centersRepository) FindStatistics(ctx context.Context) (CenterStatistics, error) {
	var statistics CenterStatistics
	err := r.GetTX(ctx).
		Raw("select (select count(*) from centers) as total_count, (select count(*) from centers where dcc = true) as dcc_count, (select count(*) from centers where visible != true) as invisible_count").
		First(&statistics).Error

	return statistics, err
}

func (r *centersRepository) FindCentersForNotification(ctx context.Context, lastUpdateAge, renotifyInterval int) ([]domain.Center, error) {
	var result []domain.Center
	err := r.GetTX(ctx).
		Raw(fmt.Sprintf(`
				select c.*
  from centers c
         join operators o on c.operator_uuid = o.uuid
  where (c.visible != false and (c.enter_date is null or c.enter_date < now()) and (c.leave_date is null or c.leave_date > now())) 
  and o.bug_reports_receiver = 'center'
  and c.last_update < now() - interval '%d weeks'
  and ((c.notified < now() - interval '%d weeks') or c.notified is null)`, lastUpdateAge, renotifyInterval)).
		Find(&result).
		Error

	return result, err
}
