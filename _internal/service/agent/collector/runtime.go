package collector

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"

	"github.com/vilasle/metrics/internal/metric"
)

type eventHandler func(c *RuntimeCollector)

type RuntimeCollector struct {
	counters map[string]metric.CounterMetric
	gauges   map[string]metric.GaugeMetric
	metrics  []string
	events   []eventHandler
}

func NewRuntimeCollector() *RuntimeCollector {
	return &RuntimeCollector{
		counters: make(map[string]metric.CounterMetric, 0),
		gauges:   make(map[string]metric.GaugeMetric, 0),
		metrics:  make([]string, 0),
		events:   make([]eventHandler, 0),
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

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	value := reflect.ValueOf(ms)
	for _, v := range c.metrics {
		fld := value.FieldByName(v)
		if !fld.IsValid() {
			return
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
			fmt.Printf("unsupported type %s\n", fld.Kind().String())
		}
	}
	c.execEvents()

}

func (c *RuntimeCollector) execEvents() {
	for _, v := range c.events {
		v(c)
	}
}

func (c *RuntimeCollector) AllMetrics() []metric.Metric {
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

func (c *RuntimeCollector) GetCounterValue(name string) metric.CounterMetric {
	if v, ok := c.counters[name]; ok {
		return v
	} else {
		return metric.NewCounterMetric(name, 0)
	}
}

func (c *RuntimeCollector) SetCounterValue(counter metric.CounterMetric) {
	c.counters[counter.Name()] = counter
}

func (c *RuntimeCollector) GetGaugeValue(name string) metric.GaugeMetric {
	if v, ok := c.gauges[name]; ok {
		return v
	} else {
		return metric.NewGaugeMetric(name, 0)
	}
}

func (c *RuntimeCollector) SetGaugeValue(gauge metric.GaugeMetric) {
	c.gauges[gauge.Name()] = gauge
}

func (c *RuntimeCollector) ResetCounter(counterName string) error {
	c.SetCounterValue(metric.NewCounterMetric(counterName, 0))
	return nil
}
