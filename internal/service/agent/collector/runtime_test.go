package collector

import (
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

func TestRuntimeCollector_RegisterEvent(t *testing.T) {
	type fields struct {
		counters map[string]metric.CounterMetric
		gauges   map[string]metric.GaugeMetric
		metrics  []string
		events   []eventHandler
	}
	type args struct {
		event eventHandler
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RuntimeCollector{
				counters: tt.fields.counters,
				gauges:   tt.fields.gauges,
				metrics:  tt.fields.metrics,
				events:   tt.fields.events,
			}
			c.RegisterEvent(tt.args.event)
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
				counter.Increment()
				c.SetCounterValue(counter)
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
	// tests := []struct {
	// 	name   string
	// 	metrics []string
	// }{
	// 	//TODO make up how test function without error
	// 	{
	// 		name: "check metrics which will be collected",
	// 		metrics: []string{"Alloc", "Frees"},
	// 	},

	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		c :=NewRuntimeCollector()

	// 		c.RegisterMetric(tt.metrics...)
	// 		c.Co

	// 		if got := c.GetCounterValue(tt.args.name); !reflect.DeepEqual(got, tt.want) {
	// 			t.Errorf("RuntimeCollector.GetCounterValue() = %v, want %v", got, tt.want)
	// 		}
	// 	})
	// }
}

func TestRuntimeCollector_SetCounterValue(t *testing.T) {
	type fields struct {
		counters map[string]metric.CounterMetric
		gauges   map[string]metric.GaugeMetric
		metrics  []string
		events   []eventHandler
	}
	type args struct {
		counter metric.CounterMetric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RuntimeCollector{
				counters: tt.fields.counters,
				gauges:   tt.fields.gauges,
				metrics:  tt.fields.metrics,
				events:   tt.fields.events,
			}
			c.SetCounterValue(tt.args.counter)
		})
	}
}

func TestRuntimeCollector_execEvents(t *testing.T) {
	type fields struct {
		counters map[string]metric.CounterMetric
		gauges   map[string]metric.GaugeMetric
		metrics  []string
		events   []eventHandler
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RuntimeCollector{
				counters: tt.fields.counters,
				gauges:   tt.fields.gauges,
				metrics:  tt.fields.metrics,
				events:   tt.fields.events,
			}
			c.execEvents()
		})
	}
}
