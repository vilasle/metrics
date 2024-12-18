package metric

import "errors"

var ErrConvertingRawValue = errors.New("error converting raw value")
var ErrInvalidMetricType = errors.New("invalid metric type")
var ErrInvalidMetric = errors.New("invalid metric value")
