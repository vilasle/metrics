package memory

import "github.com/vilasle/metrics/internal/model"

type MetricGaugeMemoryRepository[T model.Gauge] struct {
	metrics map[string]model.Gauge
}

func NewMetricGaugeMemoryRepository() *MetricGaugeMemoryRepository[model.Gauge] {
	return &MetricGaugeMemoryRepository[model.Gauge]{
		metrics: make(map[string]model.Gauge),
	}
}

func (m *MetricGaugeMemoryRepository[T]) Save(name string, metric model.Gauge) error {
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
	// result := make(map[string]model.Gauge)
	// for k, v := range m.metrics {
	// 	result[k] = v.(model.Gauge)
	// }
	// return result, nil
	return m.metrics, nil
}
