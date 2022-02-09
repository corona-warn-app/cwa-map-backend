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
	"gorm.io/gorm"
)

type SystemSettings interface {
	FindValue(ctx context.Context, key string) (*string, error)
	FindValueWithDefault(ctx context.Context, key, value string) (string, error)
}

type systemSettingsRepository struct {
	postgresqlRepository
	db *gorm.DB
	// Save persists the given center
}

func NewSystemSettingsRepository(db *gorm.DB) SystemSettings {
	return &systemSettingsRepository{
		postgresqlRepository: postgresqlRepository{db: db},
		db:                   db,
	}
}

func (s *systemSettingsRepository) FindValue(ctx context.Context, key string) (*string, error) {
	var setting domain.SystemSetting
	err := s.GetTX(ctx).
		Model(domain.SystemSetting{}).
		Where("config_key = ?", key).
		First(&setting).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return setting.ConfigValue, err
}

func (s *systemSettingsRepository) FindValueWithDefault(ctx context.Context, key, value string) (string, error) {
	val, err := s.FindValue(ctx, key)
	if err != nil {
		return "", err
	}

	if val == nil || *val == "" {
		return value, nil
	}
	return *val, nil
}
