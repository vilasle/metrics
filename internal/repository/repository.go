package repository

import "github.com/vilasle/metrics/internal/metric"

type MetricRepository interface {
	Save(metric.Metric) error
	Get(metricType string, filterName ...string) ([]metric.Metric, error)
}
