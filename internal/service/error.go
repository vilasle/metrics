package service

import "errors"

var ErrMetricIsNotExist = errors.New("metric is not exist")
var ErrStorage = errors.New("the repository finished work with error")
