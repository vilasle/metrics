package collector

import (
	"errors"
	"reflect"
	"runtime"
	"sync"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
)

type eventHandler func(c *RuntimeCollector)

type RuntimeCollector struct {
	counters map[string]metric.Metric
	gauges   map[string]metric.Metric
	metrics  []string
	events   []eventHandler
	mxMetric *sync.Mutex
}

func NewRuntimeCollector() *RuntimeCollector {
	return &RuntimeCollector{
		counters: make(map[string]metric.Metric, 0),
		gauges:   make(map[string]metric.Metric, 0),
		metrics:  make([]string, 0),
		events:   make([]eventHandler, 0),
		mxMetric: &sync.Mutex{},
	}
}

func (c *RuntimeCollector) RegisterMetric(metrics ...string) error {
	errs := make([]error, 0)

	value := reflect.ValueOf(runtime.MemStats{})

	for _, v := range metrics {
		fld := value.FieldByName(v)
		if fld.IsValid() {
			c.metrics = append(c.metrics, v)
		} else {
			errs = append(errs, errors.Join(errors.New("invalid metric"), errors.New(v)))
		}
	}

	return errors.Join(errs...)
}

func (c *RuntimeCollector) RegisterEvent(event eventHandler) {
	c.events = append(c.events, event)
}

func (c *RuntimeCollector) Collect() {
	if len(c.metrics) == 0 {
		return
	}
	c.mxMetric.Lock()

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	logger.Debug("get runtime stats", "stat", ms)

	value := reflect.ValueOf(ms)
	for _, v := range c.metrics {
		fld := value.FieldByName(v)
		if !fld.IsValid() {
			break
		}

		switch fld.Kind() {
		case reflect.Uint64,
			reflect.Uint32,
			reflect.Uint16,
			reflect.Uint8:
			c.gauges[v] = metric.NewGaugeMetric(v, float64(fld.Uint()))
		case reflect.Float32, reflect.Float64:
			c.gauges[v] = metric.NewGaugeMetric(v, fld.Float())
		default:
			logger.Error("unsupported type", "type", fld.Kind().String())
		}
	}
	c.mxMetric.Unlock()

	c.execEvents()
}

func (c *RuntimeCollector) execEvents() {
	for _, v := range c.events {
		v(c)
	}
}

func (c *RuntimeCollector) AllMetrics() []metric.Metric {
	c.mxMetric.Lock()
	defer c.mxMetric.Unlock()

	metrics := make([]metric.Metric, len(c.gauges)+len(c.counters))

	var i int
	for _, v := range c.gauges {
		metrics[i] = v
		i++
	}

	for _, v := range c.counters {
		metrics[i] = v
		i++
	}

	return metrics
}

func (c *RuntimeCollector) GetCounterValue(name string) metric.Metric {
	c.mxMetric.Lock()
	defer c.mxMetric.Unlock()

	if v, ok := c.counters[name]; ok {
		return v
	} else {
		return metric.NewCounterMetric(name, 0)
	}
}

func (c *RuntimeCollector) GetGaugeValue(name string) metric.Metric {
	c.mxMetric.Lock()
	defer c.mxMetric.Unlock()

	if v, ok := c.gauges[name]; ok {
		return v
	}
	return metric.NewGaugeMetric(name, 0)
}

func (c *RuntimeCollector) SetValue(value metric.Metric) {
	c.mxMetric.Lock()
	switch value.Type() {
	case metric.TypeGauge:
		c.gauges[value.Name()] = value
	case metric.TypeCounter:
		c.counters[value.Name()] = value
	}
	c.mxMetric.Unlock()
}

func (c *RuntimeCollector) ResetCounter(counterName string) {
	m := metric.NewCounterMetric(counterName, 0)

	c.mxMetric.Lock()
	c.counters[counterName] = m
	c.mxMetric.Unlock()
}
