package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
)

type gaugeStorage map[string]metric.Metric

type counterStorage map[string][]metric.Metric

type MemoryMetricRepository struct {
	mxGauge   *sync.Mutex
	gauges    gaugeStorage
	mxCounter *sync.Mutex
	counters  counterStorage
}

func NewMetricRepository() *MemoryMetricRepository {
	return &MemoryMetricRepository{
		mxGauge:   &sync.Mutex{},
		gauges:    make(gaugeStorage),
		mxCounter: &sync.Mutex{},
		counters:  make(counterStorage),
	}
}

func (r *MemoryMetricRepository) Save(ctx context.Context, entity ...metric.Metric) error {
	switch len(entity) {
	case 0:
		return repository.ErrEmptySetOfMetric
	case 1:
		e := entity[0]
		return r.getSaver(e.Type()).save(e)
	default:
		//TODO wrap it
		return r.saveAll(entity...)
	}
}

func (r *MemoryMetricRepository) Get(ctx context.Context, metricType string, filterName ...string) ([]metric.Metric, error) {
	return r.getGetter(metricType).get(filterName...)
}

func (r *MemoryMetricRepository) Ping(ctx context.Context) error {
	return nil
}

func (r *MemoryMetricRepository) Close() {}

func (r *MemoryMetricRepository) getSaver(metricType string) saver {
	if metricType == metric.TypeGauge {
		return gaugeSaver{storage: r.gauges, mx: r.mxGauge}
	} else if metricType == metric.TypeCounter {
		return counterSaver{storage: r.counters, mx: r.mxCounter}
	}
	return unknownSaver{}
}

func (r *MemoryMetricRepository) saveAll(entity ...metric.Metric) error {
	errs := make([]error, 0)

	for _, e := range entity {
		if err := r.getSaver(e.Type()).save(e); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (r *MemoryMetricRepository) getGetter(metricType string) getter {
	if metricType == metric.TypeGauge {
		return gaugeGetter{storage: r.gauges, mx: r.mxGauge}
	} else if metricType == metric.TypeCounter {
		return counterGetter{storage: r.counters, mx: r.mxCounter}
	}
	return unknownGetter{}
}
