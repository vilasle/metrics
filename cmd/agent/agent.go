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
//TODO add godoc
func NewCollectorAgent(collector agent.Collector, sender agent.Sender, delaySetting delay) collectorAgent {
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

//TODO add godoc
func (a collectorAgent) Run(ctx context.Context) {
	newCtx, cancel := context.WithCancel(ctx)

	go a.CollectWithContext(newCtx)
	go a.ReportWithContext(newCtx)

	<-ctx.Done()
	logger.Debug("got cancel from main")
	cancel()
}

//TODO add godoc
func (a collectorAgent) Collect() {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.Collector.Collect()
}

//TODO add godoc
func (a collectorAgent) CollectWithContext(ctx context.Context) {
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

//TODO add godoc
func (a collectorAgent) Report() {
	if err := a.report(); err != nil {
		logger.Error("failed to report metrics", "err", err)
	} else {
		a.resetPoolCounter()
	}
}

//TODO add godoc
func (a collectorAgent) ReportWithContext(ctx context.Context) {
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

func (a collectorAgent) report() (err error) {
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
