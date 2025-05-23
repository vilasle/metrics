package service

import (
	"context"

	"github.com/vilasle/metrics/internal/metric"
)

// MetricService is the interface that group methods for work with metrics
type MetricService interface {
	Save(context.Context, ...metric.Metric) error
	Get(ctx context.Context, metricType, name string) (metric.Metric, error)
	All(context.Context) ([]metric.Metric, error)
	Stats(context.Context) ([]metric.Metric, error)
	Ping(context.Context) error
	Close()
}

// Collector is the interface that group methods for collection of metrics
type Collector interface {
	Collect()
	AllMetrics() []metric.Metric
	ResetCounter(string)
}

// Sender is the interface that group methods for sending of metrics to server
type Sender interface {
	Send(...metric.Metric) error
}
