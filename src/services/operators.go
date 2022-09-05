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

package services

import (
	"com.t-systems-mms.cwa/core/security"
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/repositories"
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"time"
)

type Operators interface {
	// GetCurrentOperator gets the currently authenticated operator
	// If there is an authenticated context, but no operator exists for this subject, it will be created
	GetCurrentOperator(ctx context.Context) (domain.Operator, error)
	OperatorNotificationScheduler()
	ConfirmNotification(ctx context.Context, token string) error
}

type OperatorsServiceConfig struct {
	NotificationInterval int
	MaxLastUpdateAge     int
	RenotifyInterval     int
}

type operatorsService struct {
	operators   repositories.Operators
	config      OperatorsServiceConfig
	mailService MailService
}

func NewOperatorsService(operators repositories.Operators, config OperatorsServiceConfig, mailService MailService) Operators {
	return &operatorsService{
		operators:   operators,
		config:      config,
		mailService: mailService,
	}
}

func (o *operatorsService) GetCurrentOperator(ctx context.Context) (domain.Operator, error) {
	token, err := security.GetTokenFromContext(ctx)
	if err != nil {
		return domain.Operator{}, err
	}

	// TODO store operator in context
	return o.operators.GetOrCreateByToken(ctx, token)
}

func (o *operatorsService) ProcessOperatorNotification(ctx context.Context, operator domain.Operator) error {
	if util.IsNilOrEmpty(operator.Email) {
		return errors.New("missing email")
	}
	logrus.WithField("operator", operator.UUID).Info("Processing operator notification")

	if operator.NotificationToken == nil {
		token := uuid.New().String()
		operator.NotificationToken = &token
	}

	if err := o.mailService.ProcessTemplate(ctx,
		*operator.Email,
		"operator.notification.template",
		"operator.notification.subject",
		operator); err != nil {
		return err
	}

	notified := time.Now()
	operator.Notified = &notified

	_, err := o.operators.Save(ctx, operator)
	return err
}

func (o *operatorsService) ProcessOperatorNotifications(ctx context.Context) error {
	operators, err := o.operators.FindOperatorsForNotification(ctx, o.config.MaxLastUpdateAge, o.config.RenotifyInterval)
	if err != nil {
		logrus.WithError(err).Error("Error getting operators")
		return err
	}
	logrus.WithField("count", len(operators)).Info("Processing operator verifications")
	for _, operator := range operators {
		if err := o.ProcessOperatorNotification(ctx, operator); err != nil {
			logrus.WithError(err).Error("Error processing operator notification")
		}
	}
	return nil
}

func (o *operatorsService) OperatorNotificationScheduler() {
	for {
		if err := o.ProcessOperatorNotifications(context.Background()); err != nil {
			logrus.WithError(err).Error("Error processing operator notification")
		}
		time.Sleep(time.Duration(o.config.NotificationInterval) * time.Hour)
	}
}

func (o *operatorsService) ConfirmNotification(ctx context.Context, token string) error {
	operator, err := o.operators.FindByNotificationToken(ctx, token)
	if err != nil {
		return err
	}

	operator.NotificationToken = nil
	_, err = o.operators.Save(ctx, operator)
	return err
}
