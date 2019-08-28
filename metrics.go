package redes_writer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metricCollector struct{
	gauges gauge
}

type gauge struct {
	opsQueued prometheus.Gauge
}

func NewMetricCollector() *metricCollector {
	return newMetricCollector()
}

func newMetricCollector() *metricCollector {
	collectors := metricCollector{
		gauges: gauge{
			opsQueued: promauto.NewGauge(prometheus.GaugeOpts{
				Namespace: "redes_writer",
				Name:      "bulk_processing",
				Help:      "Number of bulks is being processes",
			}),
		},
	}

	return &collectors
}