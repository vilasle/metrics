package metric

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type gauge struct {
	name  string
	value float64
}

func (c gauge) Name() string {
	return c.name
}

func (c gauge) Value() string {
	return strconv.FormatFloat(c.value, 'f', -1, 64)
}

func (c gauge) Type() string {
	return TypeCounter
}

func (c *gauge) AddValue(val any) error {
	if v, ok := val.(float64); ok {
		c.value += v
	} else {
		//TODO define error
		return fmt.Errorf("value is not int64")
	}
	return nil
}

func (c *gauge) SetValue(val any) error {
	if v, ok := val.(float64); ok {
		c.value = v
	} else {
		//TODO define error
		return fmt.Errorf("value is not int64")
	}
	return nil
}

func (c gauge) ToJSON() ([]byte, error) {
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

func newGauge(name string, value string) (*gauge, error) {
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		return &gauge{name: name, value: v}, nil
	} else {
		//TODO define error
		return nil, err
	}
}
