package repository

import "errors"

var (
	ErrUnknownMetricType = errors.New("unknown metric type")
	ErrInitializeMetadata = errors.New("failed to initialize metadata")
)
