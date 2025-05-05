package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"time"

	"math/rand"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service/agent/collector"
	"github.com/vilasle/metrics/internal/service/agent/sender/rest/json"
)

type runConfig struct {
	report     time.Duration
	poll       time.Duration
	rateLimit  int
	endpoint   string
	hashSumKey string
}

func getConfig() runConfig {
	endpoint := flag.String("a", "localhost:8080", "endpoint to send metrics")
	reportSec := flag.Int("r", 10, "timeout(sec) for sending report to server")
	pollSec := flag.Int("p", 2, "timeout(sec) for polling metrics")
	hashSumKey := flag.String("k", "", "path to key for hash sum")
	rateLimit := flag.Int("l", 1, "rate limit for sending metrics")

	flag.Parse()

	envEndpoint := os.Getenv("ADDRESS")
	if envEndpoint != "" {
		endpoint = &envEndpoint
	}

	envReportSec := os.Getenv("REPORT_INTERVAL")
	if envReportSec != "" {
		if v, err := strconv.Atoi(envReportSec); err == nil {
			reportSec = &v
		} else {
			logger.Warn("can not parse env, will use default value",
				"env", "REPORT_INTERVAL",
				"value", envReportSec,
				"default", *reportSec)
		}
	}

	envPollSec := os.Getenv("POLL_INTERVAL")
	if envPollSec != "" {
		if v, err := strconv.Atoi(envPollSec); err == nil {
			pollSec = &v
		} else {
			logger.Warn("can not parse env, will use default value",
				"env", "POLL_INTERVAL",
				"value", envPollSec,
				"default", *pollSec)
		}
	}

	envHashSumKey := os.Getenv("KEY")
	if envHashSumKey != "" {
		hashSumKey = &envHashSumKey
	}

	envRateLimit := os.Getenv("RATE_LIMIT")
	if envRateLimit != "" {
		if v, err := strconv.Atoi(envRateLimit); err == nil {
			rateLimit = &v
		} else {
			logger.Warn("can not parse env, will use default value",
				"env", "RATE_LIMIT",
				"value", envRateLimit,
				"default", *rateLimit)
		}
	}

	return runConfig{
		endpoint:   *endpoint,
		poll:       time.Second * time.Duration(*pollSec),
		report:     time.Second * time.Duration(*reportSec),
		hashSumKey: *hashSumKey,
		rateLimit:  *rateLimit,
	}
}

func main() {
	logger.Init(os.Stdout, false)

	conf := getConfig()

	c := collector.NewRuntimeCollector()

	metrics := defaultGaugeMetrics()
	err := c.RegisterMetric(metrics...)

	if err != nil {
		logger.Fatal("can to register metric", "error", err)
	}

	pollInterval := conf.poll
	reportInterval := conf.report

	c.RegisterEvent(incrementPollCounter)

	c.RegisterEvent(func(c *collector.RuntimeCollector) {
		gauge := c.GetGaugeValue("RandomValue")
		gauge.SetValue(rand.Float64())

		c.SetValue(gauge)

	})
	c.RegisterEvent(collectExtraMetrics)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	updateAddress := fmt.Sprintf("http://%s/update/", conf.endpoint)

	hashKey, err := getHashKeyFromFile(conf.hashSumKey)
	if err != nil {
		logger.Error("can not read key from file", "file", conf.hashSumKey, "error", err)
	}

	logger.Debug("starting agent",
		"address", updateAddress,
		"pollInterval", conf.poll/time.Second,
		"reportInterval", conf.report/time.Second,
		"key", hashKey)

	sender, err := json.NewHTTPJsonSender(updateAddress, hashKey, conf.rateLimit)
	if err != nil {
		logger.Fatal("can not create sender", "err", err)
	}

	agent := newCollectorAgent(c, sender, delay{collect: pollInterval, report: reportInterval})

	ctx, cancel := context.WithCancel(context.Background())
	go agent.run(ctx)

	<-sigint
	logger.Info("stopping agent")
	cancel()

	time.Sleep(time.Millisecond * 1500)
}

func incrementPollCounter(c *collector.RuntimeCollector) {
	counter := c.GetCounterValue("PollCount")
	if err := counter.AddValue(1); err != nil {
		logger.Error("can not add value to counter", "err", err)
	}
	c.SetValue(counter)
}

func getHashKeyFromFile(path string) (string, error) {
	fd, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	if content, err := io.ReadAll(fd); err == nil {
		return string(content), err
	} else {
		return "", err
	}
}

func defaultGaugeMetrics() []string {
	return []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}
}

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
