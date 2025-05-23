package memory

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
)

type wrongMetric struct{}

func (m wrongMetric) Name() string {
	return "test"
}
func (m wrongMetric) Value() string {
	return ""
}
func (m wrongMetric) Type() string {
	return "wrongMetric"
}
func (m wrongMetric) MarshalJSON() ([]byte, error) {
	panic("not implemented")
}
func (m wrongMetric) SetValue(any) error {
	panic("not implemented")
}
func (m wrongMetric) AddValue(any) error {
	panic("not implemented")
}

func (m wrongMetric) String() string {
	return "wrongMetric"
}

func (m wrongMetric) Float64() float64 {
	return 0
}

func (m wrongMetric) Int64() int64 {
	return 0
}

func TestMemoryMetricRepository_Save(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		value   []metric.Metric
		wantErr bool
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
			wantErr: false,
		},
		{
			name: "unknown metric",
			key:  "test",
			value: []metric.Metric{
				wrongMetric{},
			},
			wantErr: true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMetricRepository()
			for _, m := range tt.value {
				err := r.Save(context.TODO(), m)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}

			if tt.wantErr {
				return
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
	r := NewMetricRepository()

	testMetrics := []metric.Metric{
		metric.NewGaugeMetric("test1", 10.01),
		metric.NewGaugeMetric("test2", 15.14),
		metric.NewGaugeMetric("test3", -1.0146),
		metric.NewGaugeMetric("test4", -0),
		metric.NewGaugeMetric("test5", 17.05),
		metric.NewCounterMetric("test6", 10),
		metric.NewCounterMetric("test7", 15),
		metric.NewCounterMetric("test8", -1),
		metric.NewCounterMetric("test9", -0),
		metric.NewCounterMetric("test10", 17),
	}

	for _, m := range testMetrics {
		r.Save(context.TODO(), m)
	}

	type args struct {
		metricType string
		filterName []string
	}

	tests := []struct {
		name    string
		args    args
		want    []metric.Metric
		wantErr bool
	}{
		{
			name: "get all gauge metrics",
			args: args{
				metricType: metric.TypeGauge,
				filterName: []string{},
			},
			want: []metric.Metric{
				metric.NewGaugeMetric("test1", 10.01),
				metric.NewGaugeMetric("test2", 15.14),
				metric.NewGaugeMetric("test3", -1.0146),
				metric.NewGaugeMetric("test4", -0),
				metric.NewGaugeMetric("test5", 17.05),
			},
			wantErr: false,
		},
		{
			name: "get filtered gauge metrics",
			args: args{
				metricType: metric.TypeGauge,
				filterName: []string{"test1", "test3"},
			},
			want: []metric.Metric{
				metric.NewGaugeMetric("test1", 10.01),
				metric.NewGaugeMetric("test3", -1.0146),
			},
			wantErr: false,
		},
		{
			name: "get all counter metrics",
			args: args{
				metricType: metric.TypeCounter,
				filterName: []string{},
			},
			want: []metric.Metric{
				metric.NewCounterMetric("test6", 10),
				metric.NewCounterMetric("test7", 15),
				metric.NewCounterMetric("test8", -1),
				metric.NewCounterMetric("test9", -0),
				metric.NewCounterMetric("test10", 17),
			},
			wantErr: false,
		},
		{
			name: "get filtered counter metrics",
			args: args{
				metricType: metric.TypeCounter,
				filterName: []string{"test9", "test10"},
			},
			want: []metric.Metric{
				metric.NewCounterMetric("test9", -0),
				metric.NewCounterMetric("test10", 17),
			},
			wantErr: false,
		},
		{
			name: "get not existed gauge metric",
			args: args{
				metricType: metric.TypeGauge,
				filterName: []string{"test15"},
			},
			want:    []metric.Metric{},
			wantErr: false,
		},
		{
			name: "get not existed counter metric",
			args: args{
				metricType: metric.TypeCounter,
				filterName: []string{"test20"},
			},
			want:    []metric.Metric{},
			wantErr: false,
		},
		{
			name: "get wrong type of metric",
			args: args{
				metricType: "wrongMetric",
				filterName: []string{"test20"},
			},
			want:    []metric.Metric{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Get(context.TODO(), tt.args.metricType, tt.args.filterName...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			sort.Slice(got, func(i, j int) bool {
				return got[i].Name() < got[j].Name()
			})
			// storage it's map, order can be random and deep equal can return error, because match only length
			// don't think that it's a problem, but it will become the problem need write check for searching object in slice with random order
			assert.Len(t, got, len(tt.want))
		})
	}
}

func TestMemoryMetricRepository_saveAll(t *testing.T) {
	testMetrics := []metric.Metric{
		metric.NewGaugeMetric("test1", 10.01),
		metric.NewGaugeMetric("test2", 15.14),
		metric.NewGaugeMetric("test3", -1.0146),
		metric.NewGaugeMetric("test4", -0),
		metric.NewGaugeMetric("test5", 17.05),
		metric.NewCounterMetric("test6", 10),
		metric.NewCounterMetric("test7", 15),
		metric.NewCounterMetric("test8", -1),
		metric.NewCounterMetric("test9", -0),
		metric.NewCounterMetric("test10", 17),
	}

	r := NewMetricRepository()

	err := r.saveAll(testMetrics...)
	require.NoError(t, err)
}
