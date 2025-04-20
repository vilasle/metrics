package server

import (
	"context"
	"errors"
	"time"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
	"github.com/vilasle/metrics/internal/service"
)

//TODO add godoc
type MetricService struct {
	storage repository.MetricRepository
}

//TODO add godoc
func NewMetricService(storage repository.MetricRepository) *MetricService {
	return &MetricService{storage: storage}
}

//TODO add godoc
func (s MetricService) Save(ctx context.Context, entity ...metric.Metric) error {
	if err := s.storage.Save(ctx, entity...); err != nil {
		return errors.Join(service.ErrStorage, err)
	}
	return nil
}

//TODO add godoc
func (s MetricService) Get(ctx context.Context, metricType, name string) (metric.Metric, error) {
	metrics, err := s.storage.Get(ctx, metricType, name)
	if err != nil {
		return nil, errors.Join(service.ErrStorage, err)
	}
	if len(metrics) == 0 {
		return nil, service.ErrMetricIsNotExist
	}

	if metricType == metric.TypeGauge {
		return metrics[0], nil
	}

	if m, err := metric.CreateSummedCounter(name, metrics); err == nil {
		return m, nil
	} else {
		return nil, errors.Join(service.ErrStorage, err)
	}
}

//TODO add godoc
func (s MetricService) All(ctx context.Context) ([]metric.Metric, error) {
	allGauges, allCounters, err := s.all(ctx)
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
			return nil, errors.Join(service.ErrStorage, err)
		}
	}
	return rs, nil
}

//TODO add godoc
func (s MetricService) Stats(ctx context.Context) ([]metric.Metric, error) {
	allGauges, allCounters, err := s.all(ctx)
	if err != nil {
		return nil, errors.Join(service.ErrStorage, err)
	}

	rs := make([]metric.Metric, 0, len(allGauges)+len(allCounters))
	rs = append(rs, allGauges...)
	rs = append(rs, allCounters...)

	return rs, nil
}

//TODO add godoc
func (s MetricService) Ping(ctx context.Context) error {
	newCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := s.storage.Ping(newCtx)
	if err != nil {
		logger.Errorw("ping database failed", "error", err)
	}
	return err
}

//TODO add godoc
func (s MetricService) Close() {
	s.storage.Close()
}

func (s MetricService) all(ctx context.Context) (gauges, counters []metric.Metric, err error) {
	allGauges, err := s.storage.Get(ctx, metric.TypeGauge)
	if err != nil {
		return nil, nil, errors.Join(service.ErrStorage, err)
	}

	allCounters, err := s.storage.Get(ctx, metric.TypeCounter)
	if err != nil {
		return nil, nil, errors.Join(service.ErrStorage, err)
	}

	return allGauges, allCounters, nil
}
