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
	"com.t-systems-mms.cwa/domain"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OperatorsStatistics struct {
	TotalCount int
}

type Operators interface {
	Repository
	FindById(ctx context.Context, id string) (domain.Operator, error)
	GetOrCreateByToken(ctx context.Context, subject jwt.Token) (domain.Operator, error)
	Save(ctx context.Context, operator domain.Operator) (domain.Operator, error)
	FindStatistics(ctx context.Context) (OperatorsStatistics, error)
	FindAll(ctx context.Context) ([]domain.Operator, error)
	Delete(ctx context.Context, id string) error
	FindOperatorsForNotification(ctx context.Context, lastUpdateAge, renotifyInterval int) ([]domain.Operator, error)
	FindByNotificationToken(ctx context.Context, token string) (domain.Operator, error)
}

type operatorsRepository struct {
	postgresqlRepository
}

func NewOperatorsRepository(db *gorm.DB) Operators {
	return &operatorsRepository{
		postgresqlRepository{db: db},
	}
}

func (r *operatorsRepository) Save(ctx context.Context, operator domain.Operator) (domain.Operator, error) {
	err := r.db.Save(&operator).Error
	return operator, err
}

func (r *operatorsRepository) FindById(ctx context.Context, id string) (domain.Operator, error) {
	if id == "" {
		return domain.Operator{}, errors.New("invalid subject")
	}

	var operator domain.Operator
	err := r.db.Model(&domain.Operator{}).Where("uuid = ?", id).First(&operator).Error
	return operator, err
}

func (r *operatorsRepository) GetOperatorBySubject(ctx context.Context, subject string) (domain.Operator, error) {
	if subject == "" {
		return domain.Operator{}, errors.New("invalid subject")
	}

	var operator domain.Operator
	err := r.db.Model(&domain.Operator{}).Where("subject = ?", subject).First(&operator).Error
	return operator, err
}

func (r *operatorsRepository) GetOrCreateByToken(ctx context.Context, token jwt.Token) (domain.Operator, error) {
	if token == nil {
		return domain.Operator{}, errors.New("invalid subject")
	}
	subject := token.Subject()

	id, err := uuid.NewUUID()
	if err != nil {
		logrus.WithError(err).Error("Error creating uuid")
		return domain.Operator{}, err
	}

	operator, err := r.GetOperatorBySubject(ctx, subject)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create the operator with default settings
		name := ""
		if entry, ok := token.Get("name"); ok {
			name = entry.(string)
		}

		var operatorNumber *string
		if entry, ok := token.Get("preferred_username"); ok {
			tmp := entry.(string)
			operatorNumber = &tmp
		}

		var email *string
		if entry, ok := token.Get("email"); ok {
			tmp := entry.(string)
			email = &tmp
		}

		receiver := domain.ReportReceiverOperator
		return r.Save(ctx, domain.Operator{
			UUID:               id.String(),
			OperatorNumber:     operatorNumber,
			Name:               name,
			Subject:            &subject,
			BugReportsReceiver: &receiver,
			Email:              email,
		})
	}
	return operator, nil
}

func (r *operatorsRepository) FindStatistics(ctx context.Context) (OperatorsStatistics, error) {
	var statistics OperatorsStatistics
	err := r.db.Raw("select (select count(*) from operators) as total_count").
		First(&statistics).
		Error
	return statistics, err
}

func (r *operatorsRepository) FindAll(ctx context.Context) ([]domain.Operator, error) {
	var result []domain.Operator
	err := r.db.
		Model(&domain.Operator{}).
		Order("uuid").
		Find(&result).Error

	return result, err
}

func (r *operatorsRepository) FindOperatorsForNotification(ctx context.Context, lastUpdateAge, renotifyInterval int) ([]domain.Operator, error) {
	var result []domain.Operator
	err := r.GetTX(ctx).
		Raw(fmt.Sprintf(`
				select o.*
  from operators o
         join centers c on o.uuid = c.operator_uuid
  where (c.visible != false and (c.enter_date is null or c.enter_date < now()) and (c.leave_date is null or c.leave_date > now()))
  and (o.bug_reports_receiver = 'operator')
  and ((o.notified < now() - interval '%d weeks') or o.notified is null)
  and (o.notification_token is null)
  group by o.uuid
  having max(c.last_update) < now() - interval '%d weeks'`, renotifyInterval, lastUpdateAge)).
		Find(&result).
		Error

	return result, err
}

func (r *operatorsRepository) FindByNotificationToken(ctx context.Context, token string) (domain.Operator, error) {
	var result domain.Operator
	err := r.GetTX(ctx).
		Model(domain.Operator{}).
		Where("notification_token = ?", token).
		First(&result).
		Error

	return result, err
}

func (r *operatorsRepository) Delete(ctx context.Context, id string) error {
	return r.GetTX(ctx).Exec("DELETE FROM operators WHERE uuid = ?", id).Error
}
