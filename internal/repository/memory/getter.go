package memory

import (
	"sync"

	"github.com/vilasle/metrics/internal/metric"
)

type getter interface {
	get(nameFilter ...string) ([]metric.Metric, error)
}

type gaugeGetter struct {
	mx      *sync.Mutex
	storage gaugeStorage
}

func (g gaugeGetter) get(nameFilter ...string) ([]metric.Metric, error) {
	g.mx.Lock()
	defer g.mx.Unlock()

	if len(nameFilter) == 0 {
		return g.all(), nil
	}

	metrics := make([]metric.Metric, 0)
	for _, name := range nameFilter {
		if m, ok := g.storage[name]; ok {
			metrics = append(metrics, m)
		}
	}
	return metrics, nil

}

func (g gaugeGetter) all() []metric.Metric {
	rs := make([]metric.Metric, 0, len(g.storage))
	for _, m := range g.storage {
		rs = append(rs, m)
	}
	return rs
}

type counterGetter struct {
	mx      *sync.Mutex
	storage counterStorage
}

func (g counterGetter) get(nameFilter ...string) ([]metric.Metric, error) {
	g.mx.Lock()
	defer g.mx.Unlock()

	if len(nameFilter) == 0 {
		return g.all(), nil
	}

	metrics := make([]metric.Metric, 0)
	for _, name := range nameFilter {
		if m, ok := g.storage[name]; ok {
			metrics = append(metrics, m...)
		}
	}
	return metrics, nil

}

func (g counterGetter) all() []metric.Metric {
	rs := make([]metric.Metric, 0, len(g.storage))
	for _, m := range g.storage {
		rs = append(rs, m...)
	}
	return rs
}

type unknownGetter struct{}

func (g unknownGetter) get(nameFilter ...string) ([]metric.Metric, error) {
	return nil, metric.ErrUnknownMetricType
}
