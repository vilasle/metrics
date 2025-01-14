package metric

import "errors"

var ErrConvertingRawValue = errors.New("error converting raw value")
var ErrUnknownMetricType = errors.New("unknown metric type")
var ErrNotFilledValue = errors.New("not filled value")
var ErrEmptyName = errors.New("name of metric is empty")
var ErrEmptyValue = errors.New("value of metric is empty")
var ErrInvalidMetric = errors.New("invalid metric data")