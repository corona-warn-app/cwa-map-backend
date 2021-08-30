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
