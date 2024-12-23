package memory

import (
	"sync"

	"github.com/vilasle/metrics/internal/model"
)

type MetricGaugeMemoryRepository[T model.Gauge] struct {
	metrics map[string]model.Gauge
	mx      *sync.RWMutex
}

func NewMetricGaugeMemoryRepository() *MetricGaugeMemoryRepository[model.Gauge] {
	return &MetricGaugeMemoryRepository[model.Gauge]{
		metrics: make(map[string]model.Gauge),
		mx:      &sync.RWMutex{},
	}
}

func (m *MetricGaugeMemoryRepository[T]) Save(name string, metric model.Gauge) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.metrics[name] = metric
	return nil
}

func (m *MetricGaugeMemoryRepository[T]) Get(name string) (model.Gauge, error) {
	//All() can not return error because ignore the error
	metric, _ := m.All()
	if v, ok := metric[name]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *MetricGaugeMemoryRepository[T]) All() (map[string]model.Gauge, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	r := make(map[string]model.Gauge, len(m.metrics))
	for k, v := range m.metrics {
		r[k] = v
	}

	return r, nil
}

func (m *MetricGaugeMemoryRepository[T]) AllAsIs() (map[string][]model.Gauge, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	r := make(map[string][]model.Gauge, len(m.metrics))
	for k, v := range m.metrics {
		r[k] = []model.Gauge{v}
	}

	return r, nil
}
