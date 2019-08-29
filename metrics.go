package redes_writer

import (
	"github.com/prometheus/client_golang/prometheus"
)

type (
	metricCollector struct {
		Histogram histogram
	}

	// Histogram metrics
	histogram struct {
		ProcessTime prometheus.Histogram
	}
)

func NewMetricCollector() (*metricCollector, error) {
	return newMetricCollector()
}

// Return a metric collector that ready to be registered into /metrics endpoint
func newMetricCollector() (*metricCollector, error) {
	collectors := metricCollector{
		Histogram: histogram{
			ProcessTime: prometheus.NewHistogram(prometheus.HistogramOpts{
				Namespace: "redes_writer",
				Name:      "process_time",
				Help:      "of time of operations processes",
				Buckets:   []float64{0.025, 0.05, 0.075, 0.1, 0.125}, //TODO: Configurate suitable bucket size
			}),
		},
	}

	return &collectors, nil
}

// Register metrics collector, based on configuration
func (mc *metricCollector) Register(cnf *Config) {
	if cnf.Prometheus.ProcessTime {
		prometheus.MustRegister(mc.Histogram.ProcessTime)
	}
}
