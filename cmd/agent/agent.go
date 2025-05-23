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

func newDelay(report, collect time.Duration) delay {
	return delay{
		report:  report,
		collect: collect,
	}
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

func (a collectorAgent) run(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(2)
	go a.collect(ctx, wg)
	go a.report(ctx, wg)
}

func (a collectorAgent) collect(ctx context.Context, wg *sync.WaitGroup) {
	t := time.NewTicker(a.collectDelay)
	defer wg.Done()
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Debug("collector got cancel signal")
			return
		case <-t.C:
			a.Collect()
		}
	}
}

func (a collectorAgent) report(ctx context.Context, wg *sync.WaitGroup) {
	t := time.NewTicker(a.reportDelay)
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			logger.Debug("reporter got cancel signal")
			logger.Debug("send collected metrics")
			t.Stop()
			a.handleReport()
			a.Sender.Close()
			return
		case <-t.C:
			t.Stop()
			a.handleReport()
			t.Reset(a.reportDelay)
		}
	}
}

func (a collectorAgent) handleReport() {
	if err := a.sendReport(); err == nil {
		a.ResetCounter("PollCount")
	} else {
		logger.Error("failed to report metrics", "err", err)
	}
}

func (a collectorAgent) sendReport() (err error) {
	a.mx.Lock()
	defer a.mx.Unlock()
	for _, d := range a.repeat {
		if err = a.Send(a.AllMetrics()...); err != nil {
			time.Sleep(d)
		} else {
			break
		}
	}
	return err
}
