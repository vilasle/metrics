package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"math/rand"

	"github.com/vilasle/metrics/internal/service/agent/collector"
	"github.com/vilasle/metrics/internal/service/agent/sender/rest/json"
)

type runConfig struct {
	endpoint string
	report   time.Duration
	poll     time.Duration
}

func getConfig() runConfig {
	endpoint := flag.String("a", "localhost:8080", "endpoint to send metrics")
	reportSec := flag.Int("r", 10, "timeout(sec) for sending report to server")
	pollSec := flag.Int("p", 2, "timeout(sec) for polling metrics")

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

	return runConfig{
		endpoint: *endpoint,
		poll:     time.Second * time.Duration(*pollSec),
		report:   time.Second * time.Duration(*reportSec),
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

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	pollTicker := time.NewTicker(pollInterval)
	reportTicker := time.NewTicker(reportInterval)

	updateAddress := fmt.Sprintf("http://%s/updates/", conf.endpoint)

	fmt.Printf("sending metrics to %s\n", updateAddress)
	fmt.Printf("pulling metrics every %d sec\n", conf.poll/time.Second)
	fmt.Printf("sending report every %d sec\n", conf.report/time.Second)

	fmt.Println("press ctrl+c to exit")

	sender, err := json.NewHTTPJsonSender(updateAddress)
	if err != nil {
		fmt.Printf("can not create sender by reason %v", err)
		os.Exit(2)
	}

	agent := NewCollectorAgent(c, sender)

	for run := true; run; {
		select {
		case <-pollTicker.C:
			agent.Collect()
		case <-reportTicker.C:
			reportTicker.Stop()
			agent.Report()
			reportTicker.Reset(reportInterval)
		case <-sigint:
			pollTicker.Stop()
			reportTicker.Stop()
			run = false
		}
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
