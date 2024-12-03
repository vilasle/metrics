package model

import (
	"errors"
	"strconv"
)

type Counter int64

func (c Counter) Type() string {
	return "counter"
}

func (c Counter) Value() string {
	return strconv.FormatInt(int64(c), 10)
}

func CounterFromString(s string) (Counter, error) {
	if v, err := strconv.ParseInt(s, 10, 64); err != nil {
		return Counter(0), errors.Join(err, ErrConvertMetricFromString)
	} else {
		return Counter(v), nil
	}
}