package service

import "github.com/vilasle/metrics/internal/metric"

//server interfaces

type StorageService interface {
	Save(metric.RawMetric) error
	Get(name string, kind string) (metric.Metric, error)
	AllMetrics() ([]metric.Metric, error)
}

//agent interfaces

type Collector interface {
	Collect()
	AllMetrics() []metric.Metric
}

type Sender interface {
	Send(metric.Metric) error
}
