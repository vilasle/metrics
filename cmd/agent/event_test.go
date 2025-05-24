package main

import (
	"runtime"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/vilasle/metrics/internal/metric"
)

func Test_incrementPollCounter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	collector := NewMockCollector(ctrl)
	sender := NewMockSender(ctrl)

	c := newCollectorAgent(collector, sender, newDelay(time.Second*2, time.Second*1))

	m1 := metric.NewCounterMetric("PollCount", 0)
	m2 := metric.NewCounterMetric("PollCount", 1)

	collector.EXPECT().GetCounterValue("PollCount").Return(m1)
	collector.EXPECT().SetValue(m1)

	collector.EXPECT().GetCounterValue("PollCount").Return(m2)

	incrementPollCounter(c)

	m := c.GetCounterValue("PollCount")

	assert.Equal(t, m2, m)
}

func Test_collectRandomValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	collector := NewMockCollector(ctrl)
	sender := NewMockSender(ctrl)

	c := newCollectorAgent(collector, sender, newDelay(time.Second*2, time.Second*1))

	m1 := metric.NewGaugeMetric("RandomValue", 1)
	m2 := metric.NewGaugeMetric("RandomValue", 2)

	collector.EXPECT().GetGaugeValue("RandomValue").Return(m1)
	collector.EXPECT().SetValue(gomock.Any())

	collector.EXPECT().GetCounterValue("RandomValue").Return(m2)

	collectRandomValue(c)

	m := c.GetCounterValue("RandomValue")

	assert.Equal(t, m2, m)
}

func Test_collectExtraMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	collector := NewMockCollector(ctrl)
	sender := NewMockSender(ctrl)

	c := newCollectorAgent(collector, sender, newDelay(time.Second*2, time.Second*1))

	qty := runtime.NumCPU()

	collector.EXPECT().SetValue(gomock.Any()).Times(qty + 2)

	collectExtraMetrics(c)
}
