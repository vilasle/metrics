package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service/agent/collector"
)

func collectExtraMetrics(c *collector.RuntimeCollector) {
	wg := &sync.WaitGroup{}

	metricCh := make(chan metric.Metric, 2+runtime.NumCPU())

	wg.Add(2)

	go collectExtraMemMetrics(wg, metricCh)
	go collectExtraCPUMetrics(wg, metricCh)

	wg.Wait()

	close(metricCh)

	for m := range metricCh {
		c.SetValue(m)
	}

}

func collectExtraMemMetrics(wg *sync.WaitGroup, metricCh chan<- metric.Metric) {
	defer wg.Done()

	if v, err := mem.VirtualMemory(); err == nil {
		metricCh <- metric.NewGaugeMetric("TotalMemory", float64(v.Total))
		metricCh <- metric.NewGaugeMetric("FreeMemory", float64(v.Free))
	} else {
		logger.Error("collection memory's metric was failed", "err", err)
	}
}

func collectExtraCPUMetrics(wg *sync.WaitGroup, metricCh chan<- metric.Metric) {
	defer wg.Done()

	result, err := cpu.Percent((time.Second * 3), true)
	if err != nil {
		logger.Error("collection cpu usage was failed", "err", err)
		return
	}

	format := "CPUutilization%d"
	for i, v := range result {
		m := fmt.Sprintf(format, i)
		metricCh <- metric.NewGaugeMetric(m, v)
	}
}

func collectRandomValue(c *collector.RuntimeCollector) {
	gauge := c.GetGaugeValue("RandomValue")
	gauge.SetValue(rand.Float64())

	c.SetValue(gauge)
}

func incrementPollCounter(c *collector.RuntimeCollector) {
	counter := c.GetCounterValue("PollCount")
	if err := counter.AddValue(1); err != nil {
		logger.Error("can not add value to counter", "err", err)
	}
	c.SetValue(counter)
}
