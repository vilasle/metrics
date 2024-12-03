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
	a.mx.Lock()
	defer a.mx.Unlock()

	for _, metric := range a.Collector.AllMetrics() {
		if err := a.Sender.Send(metric); err != nil {
			fmt.Printf("can not send metric report by reason %v", err)
		}
	}
}
