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
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type ReportStatistics struct {
	Subject      string
	OperatorUUID string
	Operator     *domain.Operator `gorm:"foreignKey:OperatorUUID"`
	Count        uint
}

type ReportCenterStatistics struct {
	Subject      string
	OperatorUUID string
	Operator     *domain.Operator `gorm:"foreignKey:OperatorUUID"`
	CenterUUID   string
	Center       *domain.Center `gorm:"foreignKey:CenterUUID"`
	Count        uint
}

type BugReports interface {
	Repository
	Save(ctx context.Context, center *domain.BugReport) error
	FindAll(ctx context.Context) ([]domain.BugReport, error)
	DeleteAll(ctx context.Context) error
	DeleteByLeader(ctx context.Context, leader string) error
	UpdateLeaderForAll(ctx context.Context, leader string) error
	FindAllByLeader(ctx context.Context, leader string) ([]domain.BugReport, error)
	ResetLeader(ctx context.Context, leader string) error

	IncrementReportCount(ctx context.Context, operatorUUID, centerUUID, subject string) error
	GetStatistics(ctx context.Context) ([]ReportStatistics, error)
	GetCenterStatistics(ctx context.Context) ([]ReportCenterStatistics, error)
}

type bugReportsRepository struct {
	postgresqlRepository
	db *gorm.DB
	// Save persists the given center
}

func NewBugReportsRepository(db *gorm.DB) BugReports {
	return &bugReportsRepository{
		postgresqlRepository: postgresqlRepository{db: db},
		db:                   db,
	}
}

func (b *bugReportsRepository) Save(ctx context.Context, report *domain.BugReport) error {
	if util.IsNilOrEmpty(&report.UUID) {
		if id, err := uuid.NewUUID(); err != nil {
			return err
		} else {
			report.UUID = id.String()
			report.Created = time.Now()
		}
	}
	return b.db.Save(report).Error
}

func (b *bugReportsRepository) DeleteAll(ctx context.Context) error {
	return b.GetTX(ctx).Exec("DELETE FROM bug_reports").Error
}

func (b *bugReportsRepository) DeleteByLeader(ctx context.Context, leader string) error {
	return b.GetTX(ctx).Exec("DELETE FROM bug_reports where leader = ?", leader).Error
}

func (b *bugReportsRepository) UpdateLeaderForAll(ctx context.Context, leader string) error {
	return b.GetTX(ctx).Exec("UPDATE bug_reports SET leader = ? WHERE leader IS NULL", leader).Error
}

func (b *bugReportsRepository) FindAllByLeader(ctx context.Context, leader string) ([]domain.BugReport, error) {
	var reports []domain.BugReport
	err := b.GetTX(ctx).Where("leader = ?", leader).Find(&reports).Error
	return reports, err
}

func (b *bugReportsRepository) ResetLeader(ctx context.Context, leader string) error {
	return b.GetTX(ctx).Exec("UPDATE bug_reports SET leader = NULL WHERE leader = ?", leader).Error
}

func (b *bugReportsRepository) FindAll(ctx context.Context) ([]domain.BugReport, error) {
	var reports []domain.BugReport
	err := b.GetTX(ctx).
		Find(&reports).Error
	return reports, err
}

func (b *bugReportsRepository) IncrementReportCount(ctx context.Context, operatorUUID, centerUUID, subject string) error {
	err := b.GetTX(ctx).Exec("insert into report_statistics (operator_uuid, subject, count) "+
		"VALUES (?, ?, 1)"+
		"on conflict on constraint report_statistics_pk "+
		"do update set count = report_statistics.count + 1", operatorUUID, subject).Error

	if err != nil {
		return err
	}

	return b.GetTX(ctx).Exec("insert into report_center_statistics (operator_uuid, center_uuid, subject, count) "+
		"VALUES (?, ?, ?, 1)"+
		"on conflict on constraint report_center_statistics_pk "+
		"do update set count = report_center_statistics.count + 1", operatorUUID, centerUUID, subject).Error
}

func (b *bugReportsRepository) GetStatistics(ctx context.Context) ([]ReportStatistics, error) {
	var statistics []ReportStatistics
	err := b.GetTX(ctx).
		Preload("Operator").
		Order("operator_uuid").
		Find(&statistics).
		Error

	return statistics, err
}

func (b *bugReportsRepository) GetCenterStatistics(ctx context.Context) ([]ReportCenterStatistics, error) {
	var statistics []ReportCenterStatistics
	err := b.GetTX(ctx).
		Preload("Operator").
		Preload("Center").
		Order("operator_uuid").
		Find(&statistics).
		Error

	return statistics, err
}
