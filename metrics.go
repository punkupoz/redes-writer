package redes_writer

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type (
	histogram struct {
		prometheus.Histogram
	}
)

// Return a metric collector that ready to be registered into /metrics endpoint
func newHistogram(name string, help string, buckets []float64) *histogram {
	collectors := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "redes_writer",
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	})

	return &histogram{collectors}
}

// Register metrics collector, based on configuration
func (m *histogram) register(active bool) *histogram {
	if active {
		err := prometheus.Register(m.Histogram)
		if err != nil {
			logrus.Warn(err)
		}
	}
	return m
}
