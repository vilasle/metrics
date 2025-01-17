package main

import (
	"fmt"
	"sync"
	"time"

	agent "github.com/vilasle/metrics/internal/service"
)

type collectorAgent struct {
	agent.Collector
	agent.Sender
	mx     *sync.Mutex
	repeat []time.Duration
}

func NewCollectorAgent(collector agent.Collector, sender agent.Sender) collectorAgent {
	return collectorAgent{
		Collector: collector,
		Sender:    sender,
		mx:        &sync.Mutex{},
		repeat: []time.Duration{
			time.Second * 1,
			time.Second * 3,
			time.Second * 5,
		},
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

func (a collectorAgent) report() (err error) {
	a.mx.Lock()
	defer a.mx.Unlock()
	for _, d := range a.repeat {
		if err = a.SendBatch(a.AllMetrics()...); err != nil {
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
