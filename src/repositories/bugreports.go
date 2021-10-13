package repositories

import (
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/domain"
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type BugReports interface {
	Repository
	Save(ctx context.Context, center *domain.BugReport) error
	FindAll(ctx context.Context) ([]domain.BugReport, error)
	DeleteAll(ctx context.Context) error
	DeleteByReceiver(ctx context.Context, receiver string) error
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

func (b *bugReportsRepository) DeleteByReceiver(ctx context.Context, receiver string) error {
	return b.GetTX(ctx).Exec("DELETE FROM bug_reports where email = ?", receiver).Error
}

func (b *bugReportsRepository) FindAll(ctx context.Context) ([]domain.BugReport, error) {
	var reports []domain.BugReport
	err := b.GetTX(ctx).Find(&reports).Error
	return reports, err
}