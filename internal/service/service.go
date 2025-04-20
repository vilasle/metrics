package service

import (
	"context"

	"github.com/vilasle/metrics/internal/metric"
)

//TODO add godoc
type MetricService interface {
	Save(context.Context, ...metric.Metric) error
	Get(ctx context.Context, metricType, name string) (metric.Metric, error)
	All(context.Context) ([]metric.Metric, error)
	Stats(context.Context) ([]metric.Metric, error)
	Ping(context.Context) error
	Close()
}

//TODO add godoc
type Collector interface {
	Collect()
	AllMetrics() []metric.Metric
	ResetCounter(string)
}

//TODO add godoc
type Sender interface {
	Send(metric.Metric) error
	SendBatch(...metric.Metric) error
	SendWithLimit(...metric.Metric) error
}
