package service

import "github.com/vilasle/metrics/internal/metric"

//server interfaces

type StorageService interface {
	Save(metric.RawMetric) error
	Get(name string, kind string) (metric.Metric, error)
	//metrics for client
	AllMetrics() ([]metric.Metric, error)
	//metrics as is
	AllMetricsAsIs() ([]metric.Metric, error)
}

//agent interfaces

type Collector interface {
	Collect()
	AllMetrics() []metric.Metric
	ResetCounter(string) error
}

type Sender interface {
	Send(metric.Metric) error
}
