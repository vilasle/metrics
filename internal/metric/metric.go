package metric

import (
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

func NewMetric(name, value, metricType string) (Metric, error) {
	switch metricType {
	case TypeGauge:
		return newGauge(name, value)
	case TypeCounter:
		return newCounter(name, value)
	default:
		//TODO define error
		return nil, fmt.Errorf("invalid metric type")
	}
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
	//TODO implement it
	panic("not implemented")
}
