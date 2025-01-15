package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type counter struct {
	name  string
	value int64
}

func (c counter) Name() string {
	return c.name
}

func (c counter) Value() string {
	return strconv.FormatInt(c.value, 10)
}

func (c counter) Type() string {
	return TypeCounter
}

func (c *counter) AddValue(val any) error {
	switch v := val.(type) {
	case int64:
		c.value += v
	case int:
		c.value += int64(v)
	default:
		return fmt.Errorf("value is %T, expect int64 or float64", val)
	}
	return nil
}

func (c *counter) SetValue(val any) error {
	if v, ok := val.(int64); ok {
		c.value = v
	} else {
		return fmt.Errorf("value is %T, expect int64", val)
	}
	return nil
}

func (c counter) ToJSON() ([]byte, error) {
	metric := struct {
		ID    string `json:"id"`
		MType string `json:"type"`
		Value int64  `json:"delta"`
	}{
		ID:    c.name,
		MType: c.Type(),
		Value: c.value,
	}
	return json.Marshal(metric)
}

func (c counter) String() string {
	return fmt.Sprintf("{type: %s; name: %s; value: %d}", c.Type(), c.name, c.value)
}

func parseCounter(name string, value string) (*counter, error) {
	if v, err := strconv.ParseInt(value, 10, 64); err == nil {
		return &counter{name: name, value: v}, nil
	} else {
		return nil, errors.Join(err, ErrConvertingRawValue)
	}
}
