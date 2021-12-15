package api

import (
	"com.t-systems-mms.cwa/api/model"
	"com.t-systems-mms.cwa/core/api"
	"com.t-systems-mms.cwa/core/security"
	"com.t-systems-mms.cwa/repositories"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"net/http"
)

type Statistics struct {
	chi.Router
	reportsRepository repositories.BugReports
}

func NewStatisticsAPI(reportsRepository repositories.BugReports, auth *jwtauth.JWTAuth) *Statistics {
	statistics := &Statistics{
		Router:            chi.NewRouter(),
		reportsRepository: reportsRepository,
	}

	statistics.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth))
		r.Use(jwtauth.Authenticator)
		r.Use(api.RequireRole(security.RoleAdmin))

		r.Get("/reports", api.Handle(statistics.getReportStatistics))
	})
	return statistics
}

func (c *Statistics) getReportStatistics(_ http.ResponseWriter, r *http.Request) (interface{}, error) {
	stats, err := c.reportsRepository.GetStatistics(r.Context())
	if err != nil {
		return nil, err
	}

	result := make([]model.ReportStatisticsDTO, len(stats))
	for i, stat := range stats {
		result[i] = *model.ReportStatisticsDTO{}.FromModel(stat)
	}
	return result, nil
}
