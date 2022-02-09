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

package main

import (
	"com.t-systems-mms.cwa/repositories"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var (
	centersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cwa_map_total_centers_count",
		Help: "The total count of centers",
	})

	dccCentersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cwa_map_dcc_centers_count",
		Help: "The count of centers with dcc enabled",
	})

	invisibleCentersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cwa_map_invisible_centers_count",
		Help: "The count of invisible centers",
	})

	operatorsCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "cwa_map_partners_count",
		Help: "The count of partners",
	})
)

func initMetricsHandler(centers repositories.Centers, operators repositories.Operators) http.Handler {
	go recordMetrics(centers, operators)
	return promhttp.Handler()
}

func recordMetrics(centers repositories.Centers, operators repositories.Operators) {
	for {
		logrus.Debug("Collecting system metrics")
		centerStats, err := centers.FindStatistics(context.Background())
		if err != nil {
			logrus.WithError(err).Error("Error getting center statistics")
		} else {
			centersCount.Set(float64(centerStats.TotalCount))
			dccCentersCount.Set(float64(centerStats.DccCount))
			invisibleCentersCount.Set(float64(centerStats.InvisibleCount))
		}

		operatorStats, err := operators.FindStatistics(context.Background())
		if err != nil {
			logrus.WithError(err).Error("Error getting operator statistics")
		} else {
			operatorsCount.Set(float64(operatorStats.TotalCount))
		}
		time.Sleep(1 * time.Hour)
	}
}
