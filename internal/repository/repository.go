package repository

import (
	"github.com/vilasle/metrics/internal/model"
)

type MetricRepository[T model.Gauge | model.Counter] interface {
	Save(string, T) error
	Get(string) (T, error)
	All() (map[string]T, error)
	AllAsIs() (map[string][]T, error)
}


type Dumper[T model.Gauge | model.Counter] interface {
	Dump(map[string]T) (error)
	DumpMetric(T) (error)
}