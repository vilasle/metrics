package repository

import (
	"context"

	"github.com/vilasle/metrics/internal/metric"
)
//MetricRepository is the interface that group methods for work with metrics' storage
type MetricRepository interface {
	Save(context.Context, ...metric.Metric) error
	Get(ctx context.Context, metricType string, filterName ...string) ([]metric.Metric, error)
	Ping(ctx context.Context) error
	Close()
}
