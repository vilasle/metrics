package service

import (
	"context"

	"github.com/vilasle/metrics/internal/metric"
)

type MetricService interface {
	Save(...metric.Metric) error
	Get(metricType, name string) (metric.Metric, error)
	All() ([]metric.Metric, error)
	Stats() ([]metric.Metric, error)
	Ping(context.Context) error
	Close()
}

type Collector interface {
	Collect()
	AllMetrics() []metric.Metric
	ResetCounter(string)
}

type Sender interface {
	Send(metric.Metric) error
}
