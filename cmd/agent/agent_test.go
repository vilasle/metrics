package main

import (
	"context"
	"sync"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
)

func Test_collectorAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	collector := NewMockCollector(ctrl)
	collector.EXPECT().Collect().AnyTimes()
	collector.EXPECT().AllMetrics().AnyTimes()
	collector.EXPECT().ResetCounter("PollCount").AnyTimes()

	sender := NewMockSender(ctrl)
	sender.EXPECT().Send(gomock.Any()).AnyTimes()
	sender.EXPECT().Close()

	agent := newCollectorAgent(
		collector, sender,
		newDelay(time.Second*2, time.Second*1))

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	agent.run(ctx, wg)

	time.Sleep(time.Second * 5)

	cancel()

	time.Sleep(time.Second * 2)

}
