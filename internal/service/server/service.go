package server

import (
	"errors"
	"fmt"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/model"
	"github.com/vilasle/metrics/internal/repository"
	"github.com/vilasle/metrics/internal/service"
)

const (
	keyGauge   = "gauge"
	keyCounter = "counter"
)

type StorageService struct {
	gaugeStorage   repository.MetricRepository[model.Gauge]
	counterStorage repository.MetricRepository[model.Counter]
}

type metricSaver interface {
	Save() error
}

type unknownSaver struct {
	kind string
}

func (c unknownSaver) Save() error {
	return errors.Join(service.ErrUnknownKind, fmt.Errorf("unknown type %s", c.kind))
}

func NewStorageService(
	gaugeStorage repository.MetricRepository[model.Gauge],
	counterStorage repository.MetricRepository[model.Counter]) *StorageService {

	return &StorageService{
		gaugeStorage:   gaugeStorage,
		counterStorage: counterStorage,
	}
}

func (s StorageService) Save(data metric.RawMetric) error {
	if err := s.checkInput(data); err != nil {
		return err
	}
	return s.save(data)
}

func (s StorageService) Get(name string, kind string) (metric.Metric, error) {
	all, err := s.getAllByKind(kind)
	if err != nil {
		return nil, err
	}

	if metric, ok := all[name]; ok {
		return metric, nil
	}

	return nil, service.ErrMetricIsNotExist
}

func (s StorageService) getAllByKind(kind string) (map[string]metric.Metric, error) {
	switch kind {
	case keyGauge:
		return s.getGaugeMetrics()
	case keyCounter:
		return s.getCounterMetrics()
	default:
		return map[string]metric.Metric{}, service.ErrUnknownKind
	}
}

func (s StorageService) getGaugeMetrics() (map[string]metric.Metric, error) {
	metrics, err := s.gaugeStorage.All()
	if err != nil {
		return map[string]metric.Metric{}, err
	}

	result := make(map[string]metric.Metric, len(metrics))
	for name, value := range metrics {
		result[name] = metric.NewGaugeMetric(name, float64(value))
	}
	return result, nil
}

func (s StorageService) getCounterMetrics() (map[string]metric.Metric, error) {
	metrics, err := s.counterStorage.All()
	if err != nil {
		return map[string]metric.Metric{}, err
	}

	result := make(map[string]metric.Metric, len(metrics))
	for name, value := range metrics {
		result[name] = metric.NewCounterMetric(name, int64(value))
	}
	return result, nil
}

func (s StorageService) AllMetrics() ([]metric.Metric, error) {
	errs := make([]error, 0, 2)
	result := make([]metric.Metric, 0)
	
	if metrics, err := s.getGaugeMetrics(); err == nil {
		result = append(result, asSlice(metrics)...)
	} else {
		errs = append(errs, err)
	}

	if metrics, err := s.getCounterMetrics(); err == nil {
		result = append(result, asSlice(metrics)...)
	} else {
		errs = append(errs, err)
	}
	return result, errors.Join(errs...)
}

func (s StorageService) save(data metric.RawMetric) error {
	saver := s.getSaverByType(data)

	err := saver.Save()
	if errors.Is(err, metric.ErrConvertingRawValue) {
		return errors.Join(err, service.ErrInvalidValue)
	}

	return err
}

func (s StorageService) checkInput(data metric.RawMetric) error {
	if data.Kind == "" {
		return service.ErrEmptyKind
	}

	if data.Name == "" {
		return service.ErrEmptyName
	}

	if data.Value == "" {
		return service.ErrEmptyValue
	}
	return nil
}

func (s StorageService) getSaverByType(data metric.RawMetric) metricSaver {
	switch data.Kind {
	case keyGauge:
		return metric.NewGaugeSaver(data, s.gaugeStorage)
	case keyCounter:
		return metric.NewCounterSaver(data, s.counterStorage)
	default:
		return unknownSaver{kind: data.Kind}
	}
}

func asSlice(m map[string]metric.Metric) []metric.Metric {
	result := make([]metric.Metric, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
