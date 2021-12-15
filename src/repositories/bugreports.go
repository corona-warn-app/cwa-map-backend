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
	Subject string
	Count   uint
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

	IncrementReportCount(ctx context.Context, subject string) error
	GetStatistics(ctx context.Context) ([]ReportStatistics, error)
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
	err := b.GetTX(ctx).Find(&reports).Error
	return reports, err
}

func (b *bugReportsRepository) IncrementReportCount(ctx context.Context, subject string) error {
	return b.GetTX(ctx).Exec("insert into report_statistics (subject, count)"+
		"VALUES (?, 1)"+
		"on conflict (subject) "+
		"do update set count = report_statistics.count + 1", subject).Error
}

func (b *bugReportsRepository) GetStatistics(ctx context.Context) ([]ReportStatistics, error) {
	var statistics []ReportStatistics
	err := b.GetTX(ctx).
		Raw("select * from report_statistics").
		Scan(&statistics).Error

	return statistics, err
}
