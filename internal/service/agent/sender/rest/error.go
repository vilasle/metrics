package rest

import "errors"

var ErrWrongMetricName = errors.New("wrong metric name")
var ErrWrongMetricTypeOrValue = errors.New("wrong metric type or value")
