package metric

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
)

type Metric interface {
	Name() string
	Value() string
	Type() string
	ToJSON() ([]byte, error)
	SetValue(any) error
	AddValue(any) error
}

func ParseMetric(name, value, metricType string) (Metric, error) {
	if err := isNotEmpty(name, value); err != nil {
		return nil, err
	}

	switch metricType {
	case TypeGauge:
		return parseGauge(name, value)
	case TypeCounter:
		return parseCounter(name, value)
	default:
		return nil, ErrUnknownMetricType
	}
}

func NewGaugeMetric(name string, value float64) (Metric, error) {
	return &gauge{name: name, value: value}, nil
}

func NewCounterMetric(name string, value int64) (Metric, error) {
	return &counter{name: name, value: value}, nil
}

func CreateSummedCounter(name string, metrics []Metric) (Metric, error) {
	var (
		sum  int64
		errs = make([]error, 0)
	)

	errFormat := "metric { name: %s; type: %s; value: %s} is not a counter"
	for _, c := range metrics {
		if v, ok := c.(*counter); ok {
			sum += v.value
		} else {
			errs = append(errs, fmt.Errorf(errFormat, c.Name(), c.Type(), c.Value()))
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return &counter{name: name, value: sum}, nil
}

func FromJSON(content []byte) (Metric, error) {
	object := struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}{}

	if err := json.Unmarshal(content, &object); err != nil {
		return nil, errors.Join(ErrInvalidMetric, err)
	} else if object.ID == "" {
		return nil, ErrInvalidMetric
	}

	if object.MType == TypeGauge {
		return createGaugeMetric(object.ID, object.Value)
	} else if object.MType == TypeCounter {
		return createCounterMetric(object.ID, object.Delta)
	} else {
		return nil, ErrUnknownMetricType
	}
}

func createGaugeMetric(name string, value *float64) (Metric, error) {
	if value == nil {
		return nil, ErrNotFilledValue
	}

	return &gauge{name, *value}, nil
}

func createCounterMetric(name string, value *int64) (Metric, error) {
	if value == nil {
		return nil, ErrNotFilledValue
	}

	return &counter{name, *value}, nil
}

func isNotEmpty(name, value string) error {
	if name == "" {
		return ErrEmptyName
	}
	if value == "" {
		return ErrEmptyValue
	}

	return nil
}
