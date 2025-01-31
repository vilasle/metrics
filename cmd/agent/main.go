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
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service/agent/collector"
	"github.com/vilasle/metrics/internal/service/agent/sender/rest/json"
)

type runConfig struct {
	endpoint   string
	report     time.Duration
	poll       time.Duration
	hashSumKey string
	rateLimit  int
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
			fmt.Printf("can not parse REPORT_INTERVAL %s. will use value %d\n", envReportSec, *reportSec)
		}
	}

	envPollSec := os.Getenv("POLL_INTERVAL")
	if envPollSec != "" {
		if v, err := strconv.Atoi(envPollSec); err == nil {
			pollSec = &v
		} else {
			fmt.Printf("can not parse POLL_INTERVAL %s. will use value %d\n", envReportSec, *pollSec)
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
			fmt.Printf("can not parse RATE_LIMIT %s. will use value %d\n", envReportSec, *rateLimit)
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
	conf := getConfig()

	c := collector.NewRuntimeCollector()

	metrics := defaultGaugeMetrics()
	err := c.RegisterMetric(metrics...)

	if err != nil {
		fmt.Printf("can to register metric by reason %v\n", err)
		os.Exit(1)
	}

	pollInterval := conf.poll
	reportInterval := conf.report

	c.RegisterEvent(func(c *collector.RuntimeCollector) {
		counter := c.GetCounterValue("PollCount")
		if err := counter.AddValue(1); err != nil {
			fmt.Printf("can not add value to counter %v\n", err)
		}
	})

	c.RegisterEvent(func(c *collector.RuntimeCollector) {
		gauge := c.GetGaugeValue("RandomValue")
		gauge.SetValue(rand.Float64())

		c.SetGaugeValue(gauge)

	})
	c.RegisterEvent(collectExtraMetrics)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	updateAddress := fmt.Sprintf("http://%s/updates/", conf.endpoint)

	hashKey, err := getHashKeyFromFile(conf.hashSumKey)
	if err != nil {
		fmt.Printf("can not read key from file %s by reason %v\n", conf.hashSumKey, err)
	}

	fmt.Printf("sending metrics to %s\n", updateAddress)
	fmt.Printf("pulling metrics every %d sec\n", conf.poll/time.Second)
	fmt.Printf("sending report every %d sec\n", conf.report/time.Second)
	fmt.Printf("using key %s\n", hashKey)

	fmt.Println("press ctrl+c to exit")

	sender, err := json.NewHTTPJsonSender(updateAddress, hashKey, conf.rateLimit)
	if err != nil {
		fmt.Printf("can not create sender by reason %v", err)
		os.Exit(2)
	}

	agent := NewCollectorAgent(c, sender, delay{collect: pollInterval, report: reportInterval})

	ctx, cancel := context.WithCancel(context.Background())
	go agent.Run(ctx)
	_ = ctx
	_ = agent

	<-sigint
	fmt.Println("stopping agent")
	cancel()

	time.Sleep(time.Millisecond * 1500)
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
		c.SetGaugeValue(m)
	}

}

func collectExtraMemMetrics(wg *sync.WaitGroup, metricCh chan<- metric.Metric) {
	defer wg.Done()

	if v, err := mem.VirtualMemory(); err == nil {
		metricCh <- metric.NewGaugeMetric("TotalMemory", float64(v.Total))
		metricCh <- metric.NewGaugeMetric("FreeMemory", float64(v.Free))
	} else {
		fmt.Printf("collection memory's metric was failed by reason %v\n", err)
	}
}

func collectExtraCPUMetrics(wg *sync.WaitGroup, metricCh chan<- metric.Metric) {
	defer wg.Done()

	result, err := cpu.Percent((time.Second * 3), true)
	if err != nil {
		fmt.Printf("collection cpu usage was failed by reason %v\n", err)
		return
	}

	format := "CPUutilization%d"
	for i, v := range result {
		m := fmt.Sprintf(format, i)
		metricCh <- metric.NewGaugeMetric(m, v)
	}
}
