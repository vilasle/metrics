package metric

import (
	"errors"

	"github.com/vilasle/metrics/internal/model"
	"github.com/vilasle/metrics/internal/repository"
)

type gaugeSaver struct {
	metric     RawMetric
	repository repository.MetricRepository[model.Gauge]
}

func NewGaugeSaver(metric RawMetric, repository repository.MetricRepository[model.Gauge]) gaugeSaver {
	return gaugeSaver{metric: metric, repository: repository}
}

func (s gaugeSaver) Save() error {
	if value, err := model.GaugeFromString(s.metric.Value); err == nil {
		return s.repository.Save(s.metric.Name, value)
	} else {
		return errors.Join(err, ErrConvertingRawValue)
	}
}

type counterSaver struct {
	metric     RawMetric
	repository repository.MetricRepository[model.Counter]
}

func NewCounterSaver(metric RawMetric, repository repository.MetricRepository[model.Counter]) counterSaver {
	return counterSaver{metric: metric, repository: repository}
}

func (s counterSaver) Save() error {
	if value, err := model.CounterFromString(s.metric.Value); err == nil {
		return s.repository.Save(s.metric.Name, value)
	} else {
		return errors.Join(err, ErrConvertingRawValue)
	}
}
