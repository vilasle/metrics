package metric

import (
	"encoding/json"

	"github.com/vilasle/metrics/internal/model"
)

type Metric interface {
	Name() string
	Type() string
	Value() string
	ToJson() ([]byte, error)
}

type GaugeMetric struct {
	name  string
	value model.Gauge
}

func NewGaugeMetric(name string, value float64) GaugeMetric {
	return GaugeMetric{name: name, value: model.Gauge(value)}
}

func (m GaugeMetric) Name() string {
	return m.name
}

func (m GaugeMetric) Type() string {
	return m.value.Type()
}

func (m GaugeMetric) Value() string {
	return m.value.Value()
}

func (m GaugeMetric) ToJson() ([]byte, error) {
	metric := struct {
		ID    string  `json:"id"`
		MType string  `json:"type"`
		Value float64 `json:"value,omitempty"`
	}{
		ID:    m.name,
		MType: m.value.Type(),
		Value: float64(m.value),
	}
	return json.Marshal(metric)
}

func (m *GaugeMetric) SetValue(v float64) {
	m.value = model.Gauge(v)
}

type CounterMetric struct {
	name  string
	value model.Counter
}

func NewCounterMetric(name string, value int64) CounterMetric {
	return CounterMetric{name: name, value: model.Counter(value)}
}

func (m CounterMetric) Name() string {
	return m.name
}

func (m CounterMetric) Value() string {
	return m.value.Value()
}

func (m CounterMetric) Type() string {
	return m.value.Type()
}

func (m CounterMetric) ToJson() ([]byte, error) {
	metric := struct {
		ID    string `json:"id"`
		MType string `json:"type"`
		Delta int64  `json:"value,omitempty"`
	}{
		ID:    m.name,
		MType: m.value.Type(),
		Delta: int64(m.value),
	}
	return json.Marshal(metric)
}



func (m *CounterMetric) Increment() {
	m.value++
}
