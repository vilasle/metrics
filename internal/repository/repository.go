package repository

import (
	"context"

	"github.com/vilasle/metrics/internal/metric"
)
//TODO add godoc
type MetricRepository interface {
	Save(context.Context, ...metric.Metric) error
	Get(ctx context.Context, metricType string, filterName ...string) ([]metric.Metric, error)
	Ping(ctx context.Context) error
	Close()
}
