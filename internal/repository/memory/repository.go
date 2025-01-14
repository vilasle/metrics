package memory

import (
	"sync"

	"github.com/vilasle/metrics/internal/metric"
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

func (r *MemoryMetricRepository) Save(entity metric.Metric) error {
	return r.getSaver(entity.Type()).save(entity)
}

func (r *MemoryMetricRepository) Get(metricType string, filterName ...string) ([]metric.Metric, error) {
	return r.getGetter(metricType).get(filterName...)
}

func (r *MemoryMetricRepository) getSaver(metricType string) saver {
	if metricType == metric.TypeGauge {
		return gaugeSaver{storage: r.gauges, mx: r.mxGauge}
	} else if metricType == metric.TypeCounter {
		return counterSaver{storage: r.counters, mx: r.mxCounter}
	}
	return unknownSaver{}
}

func (r *MemoryMetricRepository) getGetter(metricType string) getter {
	if metricType == metric.TypeGauge {
		return gaugeGetter{storage: r.gauges, mx: r.mxGauge}
	} else if metricType == metric.TypeCounter {
		return counterGetter{storage: r.counters, mx: r.mxCounter}
	}
	return unknownGetter{}
}
