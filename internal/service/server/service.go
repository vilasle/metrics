package server

import (
	"errors"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
	"github.com/vilasle/metrics/internal/service"
)

type MetricService struct {
	storage repository.MetricRepository
}

func NewMetricService(storage repository.MetricRepository) *MetricService {
	return &MetricService{storage: storage}
}

func (s MetricService) Save(entity metric.Metric) error {
	if err := s.storage.Save(entity); err != nil {
		return errors.Join(service.ErrStorage, err)
	}
	return nil
}

func (s MetricService) Get(metricType, name string) (metric.Metric, error) {
	metrics, err := s.storage.Get(metricType, name)
	if err != nil {
		return nil, errors.Join(service.ErrStorage, err)
	}
	if len(metrics) == 0 {
		return nil, service.ErrMetricIsNotExist
	}
	return metrics[0], nil
}

func (s MetricService) All() ([]metric.Metric, error) {
	allGauges, allCounters, err := s.all()
	if err != nil {
		return nil, errors.Join(service.ErrStorage, err)
	}

	rs := make([]metric.Metric, 0, len(allGauges)+len(allCounters))
	rs = append(rs, allGauges...)

	counters := make(map[string][]metric.Metric)
	for _, m := range allCounters {
		if _, ok := counters[m.Name()]; !ok {
			counters[m.Name()] = make([]metric.Metric, 0)
		}
		counters[m.Name()] = append(counters[m.Name()], m)
	}

	for name, metrics := range counters {
		if v, err := metric.CreateSummedCounter(name, metrics); err == nil {
			rs = append(rs, v)
		} else {
			//TODO wrap error
			return nil, err
		}
	}
	return rs, nil
}

func (s MetricService) Stats() ([]metric.Metric, error) {
	allGauges, allCounters, err := s.all()
	if err != nil {
		return nil, errors.Join(service.ErrStorage, err)
	}

	rs := make([]metric.Metric, 0, len(allGauges)+len(allCounters))
	rs = append(rs, allGauges...)
	rs = append(rs, allCounters...)

	return rs, nil
}

func (s MetricService) all() (gauges, counters []metric.Metric, err error) {
	allGauges, err := s.storage.Get(metric.TypeGauge)
	if err != nil {
		return nil, nil, errors.Join(service.ErrStorage, err)
	}

	allCounters, err := s.storage.Get(metric.TypeCounter)
	if err != nil {
		return nil, nil, errors.Join(service.ErrStorage, err)
	}

	return allGauges, allCounters, nil
}
