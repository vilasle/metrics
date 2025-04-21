package main

import (
	"context"
	"sync"
	"time"

	"github.com/vilasle/metrics/internal/logger"
	agent "github.com/vilasle/metrics/internal/service"
)

type collectorAgent struct {
	agent.Collector
	agent.Sender
	mx           *sync.Mutex
	repeat       []time.Duration
	reportDelay  time.Duration
	collectDelay time.Duration
}

type delay struct {
	report  time.Duration
	collect time.Duration
}

func newCollectorAgent(collector agent.Collector, sender agent.Sender, delaySetting delay) collectorAgent {
	return collectorAgent{
		Collector: collector,
		Sender:    sender,
		mx:        &sync.Mutex{},
		repeat: []time.Duration{
			time.Second * 1,
			time.Second * 3,
			time.Second * 5,
		},
		reportDelay:  delaySetting.report,
		collectDelay: delaySetting.collect,
	}
}

func (a collectorAgent) run(ctx context.Context) {
	newCtx, cancel := context.WithCancel(ctx)

	go a.collectWithContext(newCtx)
	go a.reportWithContext(newCtx)

	<-ctx.Done()
	logger.Debug("got cancel from main")
	cancel()
}

func (a collectorAgent) collect() {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.Collector.Collect()
}

func (a collectorAgent) collectWithContext(ctx context.Context) {
	t := time.NewTicker(a.collectDelay)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Debug("collector got cancel signal")
			return
		case <-t.C:
			a.Collector.Collect()
		}
	}
}

func (a collectorAgent) report() {
	if err := a.sendReport(); err != nil {
		logger.Error("failed to report metrics", "err", err)
	} else {
		a.resetPoolCounter()
	}
}

func (a collectorAgent) reportWithContext(ctx context.Context) {
	t := time.NewTicker(a.reportDelay)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Debug("reporter got cancel signal")
			return
		case <-t.C:
			t.Stop()
			a.report()
			t.Reset(a.reportDelay)
		}
	}
}

func (a collectorAgent) sendReport() (err error) {
	a.mx.Lock()
	defer a.mx.Unlock()
	for _, d := range a.repeat {
		if err = a.SendWithLimit(a.AllMetrics()...); err != nil {
			time.Sleep(d)
		} else {
			break
		}
	}
	return err
}

func (a collectorAgent) resetPoolCounter() {
	a.Collector.ResetCounter("PollCount")
}
