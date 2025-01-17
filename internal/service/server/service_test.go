package server

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
	"github.com/vilasle/metrics/internal/repository/memory"
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

func TestMetricService_Save(t *testing.T) {
	type fields struct {
		//FIXME mock interface for getting errors
		storage repository.MetricRepository
	}
	type args struct {
		entity metric.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "saving gauge metrics",
			fields: fields{
				storage: memory.NewMetricRepository(),
			},
			args: args{
				entity: metric.NewGaugeMetric("test1", 1.123),
			},
			wantErr: false,
		},
		{
			name: "saving counter metrics",
			fields: fields{
				storage: memory.NewMetricRepository(),
			},
			args: args{
				entity: metric.NewCounterMetric("test1", 1),
			},
			wantErr: false,
		},
		{
			name: "saving wrong metrics",
			fields: fields{
				storage: memory.NewMetricRepository(),
			},
			args: args{
				entity: wrongMetric{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MetricService{
				storage: tt.fields.storage,
			}
			err := s.Save(tt.args.entity)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricService_Get(t *testing.T) {
	type fields struct {
		//FIXME mock interface for getting errors
		storage repository.MetricRepository
	}

	type args struct {
		metricType string
		name       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    metric.Metric
		wantErr bool
	}{
		{
			name: "getting gauge metrics",
			fields: fields{
				storage: memory.NewMetricRepository(),
			},
			args: args{
				metricType: "gauge",
				name:       "test1",
			},
			want: metric.NewGaugeMetric("test1", 1.123),
		},
		{
			name: "getting counter metrics",
			fields: fields{
				storage: memory.NewMetricRepository(),
			},
			args: args{
				metricType: "counter",
				name:       "test2",
			},
			want: metric.NewCounterMetric("test2", 1),
		},
		{
			name: "getting wrong metrics",
			fields: fields{
				storage: memory.NewMetricRepository(),
			},
			args: args{
				metricType: "wrong",
				name:       "test3",
			},
			want:    wrongMetric{},
			wantErr: true,
		},
		{
			name: "getting non-existing metrics",
			fields: fields{
				storage: memory.NewMetricRepository(),
			},
			args: args{
				metricType: "gauge",
				name:       "test4",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := MetricService{
				storage: tt.fields.storage,
			}
			if !tt.wantErr {
				err := s.Save(tt.want)
				require.NoError(t, err)
			}

			got, err := s.Get(tt.args.metricType, tt.args.name)

			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				assert.NoError(t, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricService.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricService_All(t *testing.T) {
	type fields struct {
		//FIXME mock interface for getting errors
		storage repository.MetricRepository
	}
	tests := []struct {
		name    string
		fields  fields
		input   []metric.Metric
		want    []metric.Metric
		wantErr bool
	}{
		{
			name: "getting all metrics",
			fields: fields{
				storage: memory.NewMetricRepository(),
			},
			input: []metric.Metric{
				metric.NewGaugeMetric("test1", 1.123),
				metric.NewCounterMetric("test2", 1),
				metric.NewCounterMetric("test2", 2),
				metric.NewCounterMetric("test2", 3),
			},
			want: []metric.Metric{
				metric.NewGaugeMetric("test1", 1.123),
				metric.NewCounterMetric("test2", 6),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if !tt.wantErr {
				for _, m := range tt.input {
					err := tt.fields.storage.Save(m)
					require.NoError(t, err)
				}
			}

			s := MetricService{
				storage: tt.fields.storage,
			}
			got, err := s.All()
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricService.All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricService.All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricService_Stats(t *testing.T) {
	type fields struct {
		//FIXME mock interface for getting errors
		storage repository.MetricRepository
	}
	tests := []struct {
		name    string
		fields  fields
		input   []metric.Metric
		want    []metric.Metric
		wantErr bool
	}{
		{
			name: "getting all metrics without service processing",
			fields: fields{
				storage: memory.NewMetricRepository(),
			},
			input: []metric.Metric{
				metric.NewGaugeMetric("test1", 1.123),
				metric.NewCounterMetric("test2", 1),
				metric.NewCounterMetric("test2", 2),
				metric.NewCounterMetric("test2", 3),
			},
			want: []metric.Metric{
				metric.NewGaugeMetric("test1", 1.123),
				metric.NewCounterMetric("test2", 1),
				metric.NewCounterMetric("test2", 2),
				metric.NewCounterMetric("test2", 3),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				for _, m := range tt.input {
					err := tt.fields.storage.Save(m)
					require.NoError(t, err)
				}
			}

			s := MetricService{
				storage: tt.fields.storage,
			}
			got, err := s.Stats()
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricService.Stats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MetricService.Stats() = %v, want %v", got, tt.want)
			}
		})
	}
}
