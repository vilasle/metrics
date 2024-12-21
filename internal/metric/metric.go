package metric

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/vilasle/metrics/internal/model"
)

type Metric interface {
	Name() string
	Type() string
	Value() string
	ToJSON() ([]byte, error)
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

func (m GaugeMetric) ToJSON() ([]byte, error) {
	metric := struct {
		ID    string  `json:"id"`
		MType string  `json:"type"`
		Value float64 `json:"value"`
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

func (m CounterMetric) ToJSON() ([]byte, error) {
	metric := struct {
		ID    string `json:"id"`
		MType string `json:"type"`
		Delta int64  `json:"delta"`
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

func FromJSON(content []byte) (RawMetric, error) {
	object := struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}{}
	err := json.Unmarshal(content, &object)
	if err != nil {
		return RawMetric{}, errors.Join(ErrInvalidMetric, err)
	}

	if object.ID == "" {
		return RawMetric{}, ErrInvalidMetric
	}

	if object.MType == "gauge" {
		return newGaugeRawMetric(object.ID, object.Value)
	} else if object.MType == "counter" {
		return newCounterRawMetric(object.ID, object.Delta)
	} else {
		return RawMetric{}, ErrInvalidMetric
	}
}

func newGaugeRawMetric(name string, value *float64) (RawMetric, error) {
	if value == nil {
		return RawMetric{Name: name, Kind: "gauge"}, ErrNotFilledValue
	}

	return RawMetric{
		Name:  name,
		Kind:  "gauge",
		Value: strconv.FormatFloat(*value, 'f', -1, 64),
	}, nil
}

func newCounterRawMetric(name string, value *int64) (RawMetric, error) {
	if value == nil {
		return RawMetric{Name: name, Kind: "counter"}, ErrNotFilledValue
	}

	return RawMetric{
		Name:  name,
		Kind:  "counter",
		Value: strconv.FormatInt(*value, 10),
	}, nil
}
