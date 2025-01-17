package main

import (
	"fmt"
	"sync"

	agent "github.com/vilasle/metrics/internal/service"
)

type collectorAgent struct {
	agent.Collector
	agent.Sender
	mx *sync.Mutex
}

func NewCollectorAgent(collector agent.Collector, sender agent.Sender) collectorAgent {
	return collectorAgent{
		Collector: collector,
		Sender:    sender,
		mx:        &sync.Mutex{},
	}
}

func (a collectorAgent) Collect() {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.Collector.Collect()
}

func (a collectorAgent) Report() {
	if err := a.report(); err != nil {
		fmt.Printf("failed to report metrics: %v\n", err)
	} else {
		a.resetPoolCounter()
	}
}

func (a collectorAgent) report() error {
	a.mx.Lock()
	defer a.mx.Unlock()

	return a.SendBatch(a.AllMetrics()...)
}

func (a collectorAgent) resetPoolCounter() {
	a.Collector.ResetCounter("PollCount")
}
