package services

import (
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/repositories"
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"strings"
	"text/template"
	"time"
)

var (
	createdBugReportsCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cwa_map_bug_reports_count",
		Help: "The total count of created bug reports",
	})
)

const (
	ConfigDefaultReportsEmail  = "reports.email.default"
	ConfigReportsEmailTemplate = "reports.email.template"
	ConfigReportsEmailSubject  = "reports.email.subject"
)

type BugReportConfig struct {
	Interval int
}

type BugReports interface {
	CreateBugReport(ctx context.Context, centerUUID, subject string, message *string) (domain.BugReport, error)

	//PublishBugReports sends all pending bug reports the appropriate receivers.
	//After sending the reports will be permanently deleted
	PublishBugReports(ctx context.Context) error

	//PublishScheduler starts the scheduler for regularly sending bug reports
	PublishScheduler()
}

type bugReportsService struct {
	config               BugReportConfig
	mailService          MailService
	centersRepository    repositories.Centers
	bugReportsRepository repositories.BugReports
	settingsRepository   repositories.SystemSettings
}

func NewBugReportsService(config BugReportConfig,
	mailService MailService,
	centersRepository repositories.Centers,
	bugReportsRepository repositories.BugReports,
	settingsRepository repositories.SystemSettings) BugReports {

	return &bugReportsService{
		config:               config,
		mailService:          mailService,
		centersRepository:    centersRepository,
		bugReportsRepository: bugReportsRepository,
		settingsRepository:   settingsRepository,
	}
}

func (s *bugReportsService) CreateBugReport(ctx context.Context, centerUUID, subject string, message *string) (domain.BugReport, error) {
	// check if center exists
	center, err := s.centersRepository.FindByUUID(ctx, centerUUID)
	if err != nil {
		return domain.BugReport{}, err
	}

	var email *string
	if center.Operator.BugReportsReceiver != nil && *center.Operator.BugReportsReceiver == "center" {
		email = center.Email
	}

	if util.IsNilOrEmpty(email) {
		email = center.Operator.Email
	}

	if util.IsNilOrEmpty(email) {
		if value, err := s.settingsRepository.FindValue(ctx, ConfigDefaultReportsEmail); err != nil {
			logrus.WithError(err).WithField("key", ConfigDefaultReportsEmail).Error("Error getting config value")
			return domain.BugReport{}, errors.New("invalid config")
		} else if value == nil {
			logrus.WithField("key", ConfigDefaultReportsEmail).Error("Config not set")
			return domain.BugReport{}, errors.New("invalid config")
		} else {
			email = value
		}
	}

	report := domain.BugReport{
		Created:       time.Now(),
		Email:         *email,
		CenterUUID:    centerUUID,
		OperatorUUID:  center.OperatorUUID,
		CenterName:    center.Name,
		CenterAddress: center.Address,
		Subject:       subject,
		Message:       message,
	}

	err = s.bugReportsRepository.Save(ctx, &report)
	createdBugReportsCount.Inc()
	return report, err
}

func (s *bugReportsService) PublishBugReports(ctx context.Context) error {
	reports, err := s.bugReportsRepository.FindAll(ctx)
	if err != nil {
		return err
	}

	// getting template
	mailTemplateString, err := s.settingsRepository.FindValue(ctx, ConfigReportsEmailTemplate)
	if err != nil {
		return err
	} else if mailTemplateString == nil {
		return errors.New("missing template")
	}

	mailSubject, err := s.settingsRepository.FindValue(ctx, ConfigReportsEmailSubject)
	if err != nil {
		return err
	} else if mailSubject == nil {
		return errors.New("missing subject")
	}

	mailTemplate, err := template.New("report").Parse(*mailTemplateString)
	if err != nil {
		return err
	}

	// collect reports by receiver
	receivers := make(map[string][]domain.BugReport)
	for _, report := range reports {
		if _, ok := receivers[report.Email]; !ok {
			receivers[report.Email] = make([]domain.BugReport, 0)
		}
		receivers[report.Email] = append(receivers[report.Email], report)
	}

	for receiver, reports := range receivers {
		// collect reports by center
		centers := make(map[string][]domain.BugReport)
		for _, report := range reports {
			if _, ok := centers[report.CenterUUID]; !ok {
				centers[report.CenterUUID] = make([]domain.BugReport, 0)
			}
			centers[report.CenterUUID] = append(centers[report.CenterUUID], report)
		}

		// render template
		buffer := strings.Builder{}
		err := mailTemplate.Execute(&buffer, struct {
			Centers map[string][]domain.BugReport
		}{
			Centers: centers,
		})
		if err != nil {
			return err
		}

		// sending mail
		err = s.mailService.SendMail(ctx, receiver, *mailSubject, "text/html", buffer.String())
		if err != nil {
			return err
		}

		// delete reports for this receiver
		err = s.bugReportsRepository.DeleteByReceiver(ctx, receiver)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *bugReportsService) PublishScheduler() {
	logrus.WithFields(logrus.Fields{"interval": s.config.Interval}).Info("PublishScheduler started")
	for {
		if err := s.PublishBugReports(context.Background()); err != nil {
			logrus.WithError(err).Error("Error publishing reports")
		}
		time.Sleep(time.Duration(s.config.Interval) * time.Minute)
	}
}
