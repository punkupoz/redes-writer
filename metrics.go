package redes_writer

import (
	"github.com/prometheus/client_golang/prometheus"
)

type (
	metricCollector struct{
		Histogram histogram
	}

	histogram struct {
		ProcessTime prometheus.Histogram
	}
)

func NewMetricCollector(cnf *Config) (*metricCollector, error) {
	return newMetricCollector(cnf)
}

func newMetricCollector(cnf *Config) (*metricCollector, error) {
	collectors := metricCollector{
		Histogram: histogram{
			ProcessTime: prometheus.NewHistogram(prometheus.HistogramOpts{
				Namespace: "redes_writer",
				Name:      "process_time",
				Help:      "of time of operations processes",
				Buckets: []float64{0.025, 0.05, 0.075, 0.1, 0.125}, //TODO: Configurate suitable bucket size
			}),
		},
	}

	if cnf.Prometheus.ProcessTime {
		prometheus.MustRegister(collectors.Histogram.ProcessTime)
	}

	return &collectors, nil
}