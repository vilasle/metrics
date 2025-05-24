package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/service"
	"github.com/vilasle/metrics/internal/service/agent/collector"
	"github.com/vilasle/metrics/internal/service/agent/sender/http"
	"github.com/vilasle/metrics/internal/version"
)

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
	signal.Notify(sigint, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

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

	wg := &sync.WaitGroup{}
	sender.Start(ctx, wg)
	go agent.run(ctx, wg)

	<-sigint
	logger.Info("stopping agent")
	cancel()

	wg.Wait()
	
	logger.Info("agent stopped")

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

func registerEvents(c *collector.RuntimeCollector, events ...func(service.Collector)) {
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
