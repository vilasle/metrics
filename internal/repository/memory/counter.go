package memory

import (
	"github.com/vilasle/metrics/internal/model"
)

type MetricCounterMemoryRepository[T model.Counter] struct {
	metrics map[string][]model.Counter
}

func NewMetricCounterMemoryRepository() *MetricCounterMemoryRepository[model.Counter] {
	return &MetricCounterMemoryRepository[model.Counter]{
		metrics: make(map[string][]model.Counter),
	}
}

func (m *MetricCounterMemoryRepository[T]) Save(name string, metric model.Counter) error {
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
	result := make(map[string]model.Counter)
	for k, v := range m.metrics {
		result[k] = sum(v)
	}
	return result, nil
}

func sum(num []model.Counter) model.Counter {
	s := model.Counter(0)
	for _, num := range num {
		s += num
	}
	return s
}
