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

//Metric provided way for getting information about metric and change value
type Metric interface {
	//Name returns name of metric
	Name() string

	//Value returns value of metric as string
	Value() string
	
	//Type returns type of metric
	Type() string

	//SetValue sets value to metric, accepted types depends on metric type
	SetValue(any) error

	//AddValue adds value to metric, accept int64 or float64
	AddValue(any) error

	//Float64 returns float64 representation of metric
	Float64() float64
	
	//Int64 returns int64 representation of metric
	Int64() int64
	
	//String returns string representation of metric
	//representation string likes {type: metric_type; name: metric_name; value: metric_value}
	String() string

	//MarshalJSON returns json representation of metric
	MarshalJSON() ([]byte, error)
}

//ParseMetric returns Metric or error 
//checking process include define metricType and check that name and value are not empty
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

//NewGaugeMetric returns new gauge metric
func NewGaugeMetric(name string, value float64) Metric {
	return &gauge{name: name, value: value}
}

//NewCounterMetric returns new counter metric
func NewCounterMetric(name string, value int64) Metric {
	return &counter{name: name, value: value}
}

//CreateSummedCounter join metrics to one and return this. 
//New value for metric is sum of value of metrics from slice   
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

//FromJSON parse metric from json string and return Metric or error
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

//FromJSONArray parse metrics from json array and return slice of Metric or error
func FromJSONArray(content []byte) ([]Metric, error) {
	rs := make([]Metric, 0)
	errs := make([]error, 0)
	objects := []struct {
		ID    string   `json:"id"`
		MType string   `json:"type"`
		Delta *int64   `json:"delta,omitempty"`
		Value *float64 `json:"value,omitempty"`
	}{}

	if err := json.Unmarshal(content, &objects); err != nil {
		return nil, errors.Join(ErrInvalidMetric, err)
	}

	for _, object := range objects {
		if object.ID == "" {
			errs = append(errs, errors.Join(ErrInvalidMetric, fmt.Errorf("%v", object)))
		}

		if object.MType == TypeGauge {
			if m, err := createGaugeMetric(object.ID, object.Value); err == nil {
				rs = append(rs, m)
			} else {
				errs = append(errs, errors.Join(err, fmt.Errorf("%v", object)))
			}
		} else if object.MType == TypeCounter {
			if m, err := createCounterMetric(object.ID, object.Delta); err == nil {
				rs = append(rs, m)
			} else {
				errs = append(errs, errors.Join(err, fmt.Errorf("%v", object)))
			}
		} else {
			errs = append(errs, errors.Join(ErrUnknownMetricType, fmt.Errorf("%v", object)))
		}
	}

	return rs, errors.Join(errs...)
}

func createGaugeMetric(name string, value *float64) (Metric, error) {
	var (
		v   float64
		err error
	)
	if value != nil {
		v = *value
	} else {
		err = ErrEmptyValue
	}
	return NewGaugeMetric(name, v), err
}

func createCounterMetric(name string, value *int64) (Metric, error) {
	var (
		v   int64
		err error
	)
	if value != nil {
		v = *value
	} else {
		err = ErrEmptyValue
	}
	return NewCounterMetric(name, v), err
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
