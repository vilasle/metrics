package service

import (
	"context"

	"github.com/vilasle/metrics/internal/metric"
)

type MetricService interface {
	Save(context.Context, ...metric.Metric) error
	Get(ctx context.Context, metricType, name string) (metric.Metric, error)
	All(context.Context) ([]metric.Metric, error)
	Stats(context.Context) ([]metric.Metric, error)
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
	SendBatch(...metric.Metric) error
	SendWithLimit(...metric.Metric) error
}
