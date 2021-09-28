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
