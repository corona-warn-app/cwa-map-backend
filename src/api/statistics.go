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
