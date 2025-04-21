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

//RuntimeCollector provided way for collection runtime metrics 
// with options registration extra events where can add extra metrics or make postprocess collected metrics  
type RuntimeCollector struct {
	counters map[string]metric.Metric
	gauges   map[string]metric.Metric
	metrics  []string
	events   []eventHandler
	mxMetric *sync.Mutex
}

//NewRuntimeCollector returns new instance of RuntimeCollector
func NewRuntimeCollector() *RuntimeCollector {
	return &RuntimeCollector{
		counters: make(map[string]metric.Metric, 0),
		gauges:   make(map[string]metric.Metric, 0),
		metrics:  make([]string, 0),
		events:   make([]eventHandler, 0),
		mxMetric: &sync.Mutex{},
	}
}

//RegisterMetric register new metric and check that golang runtime provide information about this 
//return error if metric is not valid
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

//RegisterEvent register new event handler, events will raise after function Collect()
func (c *RuntimeCollector) RegisterEvent(event eventHandler) {
	c.events = append(c.events, event)
}

//Collect reads stats from runtime.MemStats, transform it to suited metrics and stores they in collections
//raises events after collecting 
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

//AllMetrics collects counter and gauge to slice of metric and return this
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

//GetCounterValue finds counter by name and returns it
func (c *RuntimeCollector) GetCounterValue(name string) metric.Metric {
	c.mxMetric.Lock()
	defer c.mxMetric.Unlock()

	if v, ok := c.counters[name]; ok {
		return v
	} else {
		return metric.NewCounterMetric(name, 0)
	}
}

//GetGaugeValue finds gauge by name and returns it
func (c *RuntimeCollector) GetGaugeValue(name string) metric.Metric {
	c.mxMetric.Lock()
	defer c.mxMetric.Unlock()

	if v, ok := c.gauges[name]; ok {
		return v
	}
	return metric.NewGaugeMetric(name, 0)
}

//SetValue replaces metric on collections to metric which are passed to input
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

//ResetCounter finds metric by name and replaces it to metric with zero value 
func (c *RuntimeCollector) ResetCounter(counterName string) {
	m := metric.NewCounterMetric(counterName, 0)

	c.mxMetric.Lock()
	c.counters[counterName] = m
	c.mxMetric.Unlock()
}

func (c *RuntimeCollector) execEvents() {
	for _, v := range c.events {
		v(c)
	}
}


