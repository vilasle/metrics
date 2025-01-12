package metric

import (
	"encoding/json"
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
	if v, ok := val.(int64); ok {
		c.value += v
	} else {
		//TODO define error
		return fmt.Errorf("value is not int64")
	}
	return nil
}

func (c *counter) SetValue(val any) error {
	if v, ok := val.(int64); ok {
		c.value = v
	} else {
		//TODO define error
		return fmt.Errorf("value is not int64")
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

func newCounter(name string, value string) (*counter, error) {
	if v, err := strconv.ParseInt(value, 10, 64); err == nil {
		return &counter{name: name, value: v}, nil
	} else {
		//TODO define error
		return nil, err
	}
}
