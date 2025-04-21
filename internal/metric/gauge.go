package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type gauge struct {
	name  string
	value float64
}

//Value returns name of metric
func (c gauge) Name() string {
	return c.name
}

//Value returns value as string
func (c gauge) Value() string {
	return strconv.FormatFloat(c.value, 'f', -1, 64)
}

//Type returns type of metric
func (c gauge) Type() string {
	return TypeGauge
}

//AddValue adds value to metric, accept int64 or float64
func (c *gauge) AddValue(val any) error {
	if v, ok := val.(float64); ok {
		c.value += v
	} else {
		return fmt.Errorf("value is %T, expect float64", val)
	}
	return nil
}

//SetValue sets value to metric, accept float64
func (c *gauge) SetValue(val any) error {
	if v, ok := val.(float64); ok {
		c.value = v
	} else {
		return fmt.Errorf("value is %T, expect float64", val)
	}
	return nil
}

//MarshalJSON returns json representation of metric
func (c gauge) MarshalJSON() ([]byte, error) {
	metric := struct {
		ID    string  `json:"id"`
		MType string  `json:"type"`
		Value float64 `json:"value"`
	}{
		ID:    c.name,
		MType: c.Type(),
		Value: c.value,
	}
	return json.Marshal(metric)
}

//String returns string representation of metric
//representation string likes {type: metric_type; name: metric_name; value: metric_value}
func (c gauge) String() string {
	return fmt.Sprintf("{type: %s; name: %s; value: %f}", c.Type(), c.name, c.value)
}

//Float64 returns float64 representation of metric
func (c gauge) Float64() float64 {
	return c.value
}

//Int64 returns int64 representation of metric
func (c gauge) Int64() int64 {
	return int64(c.value)
}

func parseGauge(name string, value string) (*gauge, error) {
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return &gauge{name: name, value: v}, nil
	} else {
		return nil, errors.Join(err, ErrConvertingRawValue)
	}
}
