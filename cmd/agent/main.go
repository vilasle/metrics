package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/service/agent/collector"
	"github.com/vilasle/metrics/internal/service/agent/sender/http"
	"github.com/vilasle/metrics/internal/version"
)

type runConfig struct {
	report     time.Duration
	poll       time.Duration
	rateLimit  int
	endpoint   string
	hashSumKey string
	cryptoKey  string
}

func getConfig() runConfig {
	endpoint := flag.String("a", "localhost:8080", "endpoint to send metrics")
	reportSec := flag.Int("r", 10, "timeout(sec) for sending report to server")
	pollSec := flag.Int("p", 2, "timeout(sec) for polling metrics")
	hashSumKey := flag.String("k", "", "path to key for hash sum")
	rateLimit := flag.Int("l", 1, "rate limit for sending metrics")
	cryptoKey := flag.String("crypto-key", "", "path to public key")

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

	envPublicKey := os.Getenv("CRYPTO_KEY")
	if envPublicKey != "" {
		cryptoKey = &envPublicKey
	}

	return runConfig{
		endpoint:   *endpoint,
		poll:       time.Second * time.Duration(*pollSec),
		report:     time.Second * time.Duration(*reportSec),
		hashSumKey: *hashSumKey,
		rateLimit:  *rateLimit,
		cryptoKey:  *cryptoKey,
	}
}

var buildVersion, buildDate, buildCommit string

func main() {
	version.ShowVersion(buildVersion, buildDate, buildCommit)

	logger.Init(os.Stdout, false)

	conf := getConfig()

	c := collector.NewRuntimeCollector()

	if err := c.RegisterMetric(defaultGaugeMetrics()...); err != nil {
		logger.Fatal("can to register metric", "error", err)
	}

	registerEvents(c, incrementPollCounter, collectExtraMetrics, collectRandomValue)

	ctx, cancel := context.WithCancel(context.Background())

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	addr := fmt.Sprintf("http://%s/update/", conf.endpoint)

	logger.Debug("starting agent",
		"address", addr,
		"pollInterval", conf.poll/time.Second,
		"reportInterval", conf.report/time.Second,
	)

	sender, err := createSender(conf.hashSumKey, conf.cryptoKey, addr, conf.rateLimit)
	if err != nil {
		logger.Fatal("can not create sender", "err", err)
	}

	agent := newCollectorAgent(c, sender, newDelay(conf.poll, conf.report))

	sender.Start(ctx)
	go agent.run(ctx)

	<-sigint
	logger.Info("stopping agent")
	cancel()

	time.Sleep(time.Millisecond * 1500)
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

func registerEvents(c *collector.RuntimeCollector, events ...func(*collector.RuntimeCollector)) {
	for _, event := range events {
		c.RegisterEvent(event)
	}
}

func createSender(hashPath, cryptoKeyPath, addr string, rateLimit int) (*http.HTTPSender, error) {
	hashKey, err := getHashKeyFromFile(hashPath)
	if err != nil {
		logger.Error("can not read key from file", "file", hashPath, "error", err)
	}

	publicKey, err := getPublicKeyFromFile(cryptoKeyPath)
	if err != nil {
		logger.Error("can not read public key from file", "file", cryptoKeyPath, "error", err)
	}

	bodyWriter := http.NewJSONWriter(
		http.WithCalculateHashSum(hashKey),
		http.WithEncryption(publicKey),
		http.WithCompressing(),
	)

	maker, err := http.NewJSONRequestMaker(addr, bodyWriter)
	if err != nil {
		return nil, errors.Join(err, errors.New("can not create request maker"))
	}

	return http.NewHTTPSender(
		maker,
		http.WithRateLimit(rateLimit),
	), nil

}

func getHashKeyFromFile(path string) ([]byte, error) {
	if content, err := os.ReadFile(path); err == nil {
		return content, err
	} else {
		return nil, err
	}
}

func getPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	if path == "" {
		return nil, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	publicBlock, _ := pem.Decode(content)
	return x509.ParsePKCS1PublicKey(publicBlock.Bytes)
}
