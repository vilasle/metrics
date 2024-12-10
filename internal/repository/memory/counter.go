package memory

import (
	"sync"

	"github.com/vilasle/metrics/internal/model"
)

type MetricCounterMemoryRepository[T model.Counter] struct {
	mx      *sync.RWMutex
	metrics map[string][]model.Counter
}

func NewMetricCounterMemoryRepository() *MetricCounterMemoryRepository[model.Counter] {
	return &MetricCounterMemoryRepository[model.Counter]{
		metrics: make(map[string][]model.Counter),
		mx:      &sync.RWMutex{},
	}
}

func (m *MetricCounterMemoryRepository[T]) Save(name string, metric model.Counter) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	if _, ok := m.metrics[name]; !ok {
		m.metrics[name] = make([]model.Counter, 0)
	}
	m.metrics[name] = append(m.metrics[name], metric)

	return nil
}

func (m *MetricCounterMemoryRepository[T]) Get(name string) (model.Counter, error) {
	//All() can not return error because ignore the error
	metric, _ := m.All()
	if v, ok := metric[name]; ok {
		return v, nil
	}
	return 0, nil
}

func (m *MetricCounterMemoryRepository[T]) All() (map[string]model.Counter, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	result := make(map[string]model.Counter)
	for k, v := range m.metrics {
		result[k] = v[len(v)-1]
	}
	return result, nil
}
