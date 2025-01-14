package memory

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vilasle/metrics/internal/metric"
)

func TestMemoryMetricRepository_Save(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value []metric.Metric
	}{
		{
			name: "save one metric",
			key:  "test",
			value: []metric.Metric{
				metric.NewGaugeMetric("test1", 10.00),
				metric.NewCounterMetric("test1", 10),
			},
		},
		{
			name: "save several metric",
			key:  "test",
			value: []metric.Metric{
				metric.NewGaugeMetric("test1", 10.01),
				metric.NewGaugeMetric("test2", 15.14),
				metric.NewGaugeMetric("test3", -1.0146),
				metric.NewGaugeMetric("test4", -0),
				metric.NewGaugeMetric("test5", 17.05),
				metric.NewCounterMetric("test1", 10),
				metric.NewCounterMetric("test2", 15),
				metric.NewCounterMetric("test3", -1),
				metric.NewCounterMetric("test4", -0),
				metric.NewCounterMetric("test5", 17),
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMetricRepository()
			for _, m := range tt.value {
				r.Save(m)
			}

			for _, m := range tt.value {
				if m.Type() == metric.TypeGauge {
					v := r.gauges[m.Name()]
					assert.True(t, reflect.DeepEqual(v, m))
				} else if m.Type() == metric.TypeCounter {
					v := r.counters[m.Name()][0]
					assert.True(t, reflect.DeepEqual(v, m))
				}
			}
		})
	}
}

func TestMemoryMetricRepository_Get(t *testing.T) {
	type fields struct {
		mxGauge   *sync.Mutex
		gauges    gaugeStorage
		mxCounter *sync.Mutex
		counters  counterStorage
	}
	type args struct {
		metricType string
		filterName []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []metric.Metric
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MemoryMetricRepository{
				mxGauge:   tt.fields.mxGauge,
				gauges:    tt.fields.gauges,
				mxCounter: tt.fields.mxCounter,
				counters:  tt.fields.counters,
			}
			got, err := r.Get(tt.args.metricType, tt.args.filterName...)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryMetricRepository.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryMetricRepository.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryMetricRepository_getSaver(t *testing.T) {
	type fields struct {
		mxGauge   *sync.Mutex
		gauges    gaugeStorage
		mxCounter *sync.Mutex
		counters  counterStorage
	}
	type args struct {
		metricType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   saver
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MemoryMetricRepository{
				mxGauge:   tt.fields.mxGauge,
				gauges:    tt.fields.gauges,
				mxCounter: tt.fields.mxCounter,
				counters:  tt.fields.counters,
			}
			if got := r.getSaver(tt.args.metricType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryMetricRepository.getSaver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemoryMetricRepository_getGetter(t *testing.T) {
	type fields struct {
		mxGauge   *sync.Mutex
		gauges    gaugeStorage
		mxCounter *sync.Mutex
		counters  counterStorage
	}
	type args struct {
		metricType string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   getter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MemoryMetricRepository{
				mxGauge:   tt.fields.mxGauge,
				gauges:    tt.fields.gauges,
				mxCounter: tt.fields.mxCounter,
				counters:  tt.fields.counters,
			}
			if got := r.getGetter(tt.args.metricType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MemoryMetricRepository.getGetter() = %v, want %v", got, tt.want)
			}
		})
	}
}
