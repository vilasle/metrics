package model

import (
	"errors"
	"strconv"
)

type Gauge float64

func (g Gauge) Type() string {
	return "gauge"
}

func (g Gauge) Value() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func GaugeFromString(s string) (Gauge, error) {
	if v, err := strconv.ParseFloat(s, 64); err != nil {
		return Gauge(0), errors.Join(err, ErrConvertMetricFromString)
	} else {
		return Gauge(v), nil
	}
}
