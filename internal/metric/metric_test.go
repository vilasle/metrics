package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vilasle/metrics/internal/model"
	"github.com/vilasle/metrics/internal/repository"
	"github.com/vilasle/metrics/internal/repository/memory"
)

func Test_gaugeSaver_save(t *testing.T) {
	type fields struct {
		metric     RawMetric
		repository repository.MetricRepository[model.Gauge]
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "correct value",
			fields: fields{
				metric:     NewRawMetric("test", "gauge", "1234"),
				repository: memory.NewMetricGaugeMemoryRepository(),
			},
			wantErr: false,
		},
		{
			name: "invalid value",
			fields: fields{
				metric:     NewRawMetric("test", "gauge", "gdfgdf"),
				repository: memory.NewMetricGaugeMemoryRepository(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewGaugeSaver(tt.fields.metric, tt.fields.repository)
			if err := s.Save(); (err != nil) != tt.wantErr {
				t.Errorf("gaugeSaver.save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_counterSaver_save(t *testing.T) {
	type fields struct {
		metric     RawMetric
		repository repository.MetricRepository[model.Counter]
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "correct value",
			fields: fields{
				metric:     NewRawMetric("test", "counter", "1234"),
				repository: memory.NewMetricCounterMemoryRepository(),
			},
			wantErr: false,
		},
		{
			name: "invalid value",
			fields: fields{
				metric:     NewRawMetric("test", "counter", "gdfgdf"),
				repository: memory.NewMetricCounterMemoryRepository(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewCounterSaver(tt.fields.metric, tt.fields.repository)
			if err := s.Save(); (err != nil) != tt.wantErr {
				t.Errorf("counterSaver.save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGaugeMetric_Name(t *testing.T) {
	type fields struct {
		name  string
		value model.Gauge
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "gauge value 65",
			fields: fields{
				name:  "test",
				value: 65,
			},
			want: "test",
		},
		{
			name: "gauge value 150",
			fields: fields{
				name:  "test1",
				value: 150,
			},
			want: "test1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewGaugeMetric(tt.fields.name, float64(tt.fields.value))
			if got := m.Name(); got != tt.want {
				t.Errorf("GaugeMetric.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGaugeMetric_Type(t *testing.T) {
	type fields struct {
		name  string
		value model.Gauge
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "gauge",
			fields: fields{
				name:  "test",
				value: 0,
			},
			want: "gauge",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewGaugeMetric(tt.fields.name, float64(tt.fields.value))
			if got := m.Type(); got != tt.want {
				t.Errorf("GaugeMetric.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGaugeMetric_Value(t *testing.T) {
	type fields struct {
		name  string
		value model.Gauge
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "counter value 65",
			fields: fields{
				name:  "test",
				value: 65.00,
			},
			want: "65",
		},
		{
			name: "counter value 150",
			fields: fields{
				name:  "test",
				value: 150,
			},
			want: "150",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewGaugeMetric(tt.fields.name, float64(tt.fields.value))
			if got := m.Value(); got != tt.want {
				t.Errorf("GaugeMetric.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGaugeMetric_SetValue(t *testing.T) {
	type fields struct {
		name  string
		value model.Gauge
	}
	type args struct {
		v float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "gauge value 65",
			fields: fields{
				name:  "test",
				value: 65,
			},
			args: args{
				v: 100,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &GaugeMetric{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			m.SetValue(tt.args.v)

			if tt.args.v != float64(m.value) {
				t.Errorf("GaugeMetric.SetValue() = %v, want %v", m.value, tt.args.v)
			}

		})
	}
}

func TestCounterMetric_Name(t *testing.T) {
	type fields struct {
		name  string
		value model.Counter
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "counter value 65",
			fields: fields{
				name:  "test",
				value: 65,
			},
			want: "test",
		},
		{
			name: "counter value 150",
			fields: fields{
				name:  "test",
				value: 150,
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCounterMetric(tt.fields.name, int64(tt.fields.value))
			if got := m.Name(); got != tt.want {
				t.Errorf("CounterMetric.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCounterMetric_Value(t *testing.T) {
	type fields struct {
		name  string
		value model.Counter
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "counter value 65",
			fields: fields{
				name:  "test",
				value: 65,
			},
			want: "65",
		},
		{
			name: "counter value 150",
			fields: fields{
				name:  "test",
				value: 150,
			},
			want: "150",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCounterMetric(tt.fields.name, int64(tt.fields.value))
			if got := m.Value(); got != tt.want {
				t.Errorf("CounterMetric.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCounterMetric_Type(t *testing.T) {
	type fields struct {
		name  string
		value model.Counter
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "counter",
			fields: fields{
				name:  "test",
				value: 0,
			},
			want: "counter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCounterMetric(tt.fields.name, int64(tt.fields.value))
			if got := m.Type(); got != tt.want {
				t.Errorf("CounterMetric.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCounterMetric_Increment(t *testing.T) {
	type fields struct {
		name  string
		value model.Counter
	}
	tests := []struct {
		name   string
		fields fields
		count  int
		want   int
	}{
		{
			name: "increment 1",
			fields: fields{
				name:  "test",
				value: 0,
			},
			count: 1,
			want:  1,
		},
		{
			name: "increment 10",
			fields: fields{
				name:  "test",
				value: 0,
			},
			count: 10,
			want:  10,
		},
		{
			name: "increment 1000",
			fields: fields{
				name:  "test",
				value: 0,
			},
			count: 1000,
			want:  1000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCounterMetric(tt.fields.name, int64(tt.fields.value))
			for i := 0; i < tt.count; i++ {
				m.Increment()
			}

			assert.Equal(t, tt.want, int(m.value))

		})
	}
}
