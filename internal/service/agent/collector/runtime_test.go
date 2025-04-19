package collector

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vilasle/metrics/internal/metric"
)

func TestRuntimeCollector_RegisterMetric(t *testing.T) {
	type args struct {
		metrics []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "register metrics",
			args: args{
				metrics: []string{"test_metric"},
			},
			wantErr: true,
		},
		{
			name: "duplicate metrics",
			args: args{
				metrics: []string{"Alloc", "Frees"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewRuntimeCollector()
			if err := c.RegisterMetric(tt.args.metrics...); (err != nil) != tt.wantErr {
				t.Errorf("RuntimeCollector.RegisterMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRuntimeCollector_Collect(t *testing.T) {
	tests := []struct {
		name        string
		metrics     []string
		pushMetrics []string
	}{
		{
			name:    "check metrics which will be collected",
			metrics: []string{"Alloc", "Frees"},
		},
		{
			name:    "there are not metrics",
			metrics: []string{},
		},
		{
			name:        "there are invalid metrics",
			metrics:     []string{},
			pushMetrics: []string{"test_metric"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewRuntimeCollector()

			c.RegisterMetric(tt.metrics...)

			c.metrics = append(c.metrics, tt.pushMetrics...)

			c.Collect()
			for _, v := range tt.metrics {
				_, ok := c.gauges[v]
				assert.Equal(t, true, ok)
			}
		})
	}
}

func TestRuntimeCollector_AllMetrics(t *testing.T) {
	tests := []struct {
		name           string
		metrics        []string
		wantCount      int
		eventCollector eventHandler
	}{
		{
			name:      "must have 2 metrics",
			metrics:   []string{"Alloc", "Frees"},
			wantCount: 2,
			eventCollector: func(c *RuntimeCollector) {
				counter := c.GetCounterValue("test")
				counter.AddValue(1)
				c.SetValue(counter)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewRuntimeCollector()
			c.RegisterMetric(tt.metrics...)

			c.RegisterEvent(tt.eventCollector)

			c.Collect()

			c.execEvents()

			got := c.AllMetrics()
			assert.Len(t, got, tt.wantCount+1)
		})
	}
}

func TestRuntimeCollector_GetCounterValue(t *testing.T) {
	tests := []struct {
		name      string
		metrics   map[string]metric.Metric
		wantKey   string
		wantValue metric.Metric
	}{
		{
			name: "check metrics which will be collected",
			metrics: map[string]metric.Metric{
				"test1": metric.NewCounterMetric("test1", 15),
				"test2": metric.NewCounterMetric("test2", 30),
				"test3": metric.NewCounterMetric("test3", 45),
				"test4": metric.NewCounterMetric("test4", 55),
				"test5": metric.NewCounterMetric("test5", 65),
			},
			wantKey:   "test3",
			wantValue: metric.NewCounterMetric("test3", 45),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := RuntimeCollector{
				counters: tt.metrics,
				mxMetric: &sync.Mutex{},
			}
			got := c.GetCounterValue(tt.wantKey)
			assert.Equal(t, tt.wantValue, got)
		})
	}
}

func TestRuntimeCollector_SetGetGaugeValue(t *testing.T) {
	tests := []struct {
		name    string
		storage map[string]metric.Metric
		update  metric.Metric
	}{
		{
			name: "check metrics which will be collected",
			storage: map[string]metric.Metric{
				"test1": metric.NewGaugeMetric("test1", 15),
				"test2": metric.NewGaugeMetric("test2", 55),
				"test3": metric.NewGaugeMetric("test3", 45),
				"test4": metric.NewGaugeMetric("test4", 65),
				"test5": metric.NewGaugeMetric("test5", 3135),
				"test6": metric.NewGaugeMetric("test6", 3455),
			},
			update: metric.NewGaugeMetric("test1", 105.673),
		},
		{
			name: "set metric which does not exists yet",
			storage: map[string]metric.Metric{
				"test1": metric.NewGaugeMetric("test1", 15),
				"test2": metric.NewGaugeMetric("test2", 55),
				"test3": metric.NewGaugeMetric("test3", 45),
				"test4": metric.NewGaugeMetric("test4", 65),
				"test5": metric.NewGaugeMetric("test5", 3135),
				"test6": metric.NewGaugeMetric("test6", 3455),
			},
			update: metric.NewGaugeMetric("test1564", 105.673),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RuntimeCollector{
				gauges:   tt.storage,
				mxMetric: &sync.Mutex{},
			}
			c.GetGaugeValue(tt.update.Name())

			c.SetValue(tt.update)

			assert.Equal(t, tt.update, c.GetGaugeValue(tt.update.Name()))
		})
	}
}

func TestRuntimeCollector_ResetGetCounterValue(t *testing.T) {
	tests := []struct {
		name      string
		metrics   map[string]metric.Metric
		wantKey   string
		wantValue metric.Metric
	}{
		{
			name: "check metrics which will be collected",
			metrics: map[string]metric.Metric{
				"test1": metric.NewCounterMetric("test1", 15),
				"test2": metric.NewCounterMetric("test2", 30),
				"test3": metric.NewCounterMetric("test3", 45),
				"test4": metric.NewCounterMetric("test4", 55),
				"test5": metric.NewCounterMetric("test5", 65),
			},
			wantKey:   "test3",
			wantValue: metric.NewCounterMetric("test3", 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := RuntimeCollector{
				counters: tt.metrics,
				mxMetric: &sync.Mutex{},
			}
			c.ResetCounter(tt.wantKey)

			got := c.GetCounterValue(tt.wantKey)
			assert.Equal(t, tt.wantValue, got)
		})
	}
}

func BenchmarkRuntimeCollector_Collect(b *testing.B) {
	defaultMetrics := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
		"GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects",
		"HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse",
		"MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC",
		"NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
		"StackSys", "Sys", "TotalAlloc",
	}
	c := NewRuntimeCollector()

	c.RegisterMetric(defaultMetrics...)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Collect()
	}
}

