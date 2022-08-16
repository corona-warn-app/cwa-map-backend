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
	"com.t-systems-mms.cwa/core/api"
	"com.t-systems-mms.cwa/core/security"
	"com.t-systems-mms.cwa/core/util"
	"com.t-systems-mms.cwa/domain"
	"com.t-systems-mms.cwa/repositories"
	"encoding/csv"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
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

		r.Get("/reports", statistics.getReportStatistics)
		r.Get("/reports/centers", statistics.getCenterReportsStatistics)
	})
	return statistics
}

func (c *Statistics) getCenterReportsStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := c.reportsRepository.GetCenterStatistics(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Error getting report statistics")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	subjects := make([]string, 0)
	operators := make(map[string]domain.Operator)
	centers := make(map[string]domain.Center)
	data := make(map[string]map[string]map[string]uint)
	for _, value := range stats {
		knownSubject := false
		for _, s := range subjects {
			if s == value.Subject {
				knownSubject = true
				break
			}
		}

		if !knownSubject {
			subjects = append(subjects, value.Subject)
		}

		if _, ok := data[value.OperatorUUID]; !ok {
			data[value.OperatorUUID] = make(map[string]map[string]uint)
			operators[value.OperatorUUID] = *value.Operator
		}

		if _, ok := data[value.OperatorUUID][value.CenterUUID]; !ok {
			data[value.OperatorUUID][value.CenterUUID] = make(map[string]uint)
			centers[value.CenterUUID] = *value.Center
		}

		data[value.OperatorUUID][value.CenterUUID][value.Subject] = value.Count
	}

	w.Header().Set("Content-Type", "text/csv")
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		logrus.WithError(err).Error("Error writing BOM")
		return
	}

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	headers := []string{"partner_uuid", "partner_number", "partner_name", "center_uuid", "center_name"}
	headers = append(headers, subjects...)
	if err := csvWriter.Write(headers); err != nil {
		logrus.WithError(err).Error("Error writing response")
		return
	}

	for operator, operatorCenters := range data {
		for center, entry := range operatorCenters {
			columns := []string{
				operator,
				util.PtrToString(operators[operator].OperatorNumber, ""),
				operators[operator].Name,
				center,
				centers[center].Name,
			}
			for _, subject := range subjects {
				columns = append(columns, strconv.Itoa(int(entry[subject])))
			}

			if err := csvWriter.Write(columns); err != nil {
				logrus.WithError(err).Error("Error writing response")
				return
			}
		}
	}

	csvWriter.Flush()
}

func (c *Statistics) getReportStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := c.reportsRepository.GetStatistics(r.Context())
	if err != nil {
		logrus.WithError(err).Error("Error getting report statistics")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	subjects := make([]string, 0)
	operators := make(map[string]domain.Operator)
	data := make(map[string]map[string]uint)
	for _, value := range stats {
		knownSubject := false
		for _, s := range subjects {
			if s == value.Subject {
				knownSubject = true
				break
			}
		}

		if !knownSubject {
			subjects = append(subjects, value.Subject)
		}

		if _, ok := data[value.OperatorUUID]; !ok {
			data[value.OperatorUUID] = make(map[string]uint)
			operators[value.OperatorUUID] = *value.Operator
		}

		data[value.OperatorUUID][value.Subject] = value.Count
	}

	w.Header().Set("Content-Type", "text/csv")
	if _, err := w.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		logrus.WithError(err).Error("Error writing BOM")
		return
	}

	csvWriter := csv.NewWriter(w)
	csvWriter.Comma = ';'

	headers := []string{"partner_uuid", "partner_number", "partner_name"}
	headers = append(headers, subjects...)
	if err := csvWriter.Write(headers); err != nil {
		logrus.WithError(err).Error("Error writing response")
		return
	}

	for operator, entry := range data {
		columns := []string{operator, util.PtrToString(operators[operator].OperatorNumber, ""), operators[operator].Name}
		for _, subject := range subjects {
			columns = append(columns, strconv.Itoa(int(entry[subject])))
		}

		if err := csvWriter.Write(columns); err != nil {
			logrus.WithError(err).Error("Error writing response")
			return
		}
	}

	csvWriter.Flush()
}
