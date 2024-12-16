package server

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/model"
	"github.com/vilasle/metrics/internal/repository"
	"github.com/vilasle/metrics/internal/repository/memory"
	"github.com/vilasle/metrics/internal/service"
)

func TestStorageService_Save(t *testing.T) {
	type fields struct {
		gaugeStorage   repository.MetricRepository[model.Gauge]
		counterStorage repository.MetricRepository[model.Counter]
	}

	_fields := fields{
		gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
		counterStorage: memory.NewMetricCounterMemoryRepository(),
	}

	type args struct {
		metric metric.RawMetric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		err    error
	}{
		{
			name:   "problems with input, did not fill name of value",
			fields: _fields,
			args: args{
				metric: metric.NewRawMetric("", keyGauge, "12.45"),
			},
			err: service.ErrEmptyName,
		},
		{
			name:   "problems with input, did not fill kind of metric",
			fields: _fields,
			args: args{
				metric: metric.NewRawMetric("test", "", "12.45"),
			},
			err: service.ErrEmptyKind,
		},
		{
			name:   "problems with input, did not fill value of metric",
			fields: _fields,
			args: args{
				metric: metric.NewRawMetric("test", keyGauge, ""),
			},
			err: service.ErrEmptyValue,
		},
		{
			name:   "metric is filled but contents unknown kind",
			fields: _fields,
			args: args{
				metric: metric.NewRawMetric("test", "test", "123.4"),
			},
			err: service.ErrUnknownKind,
		},
		{
			name:   "gauge metric is filled and kind is right",
			fields: _fields,
			args: args{
				metric: metric.NewRawMetric("test", keyGauge, "123.4"),
			},
			err: nil,
		},
		{
			name:   "gauge metric is filled and kind is right, value has wrong type",
			fields: _fields,
			args: args{
				metric: metric.NewRawMetric("test", keyGauge, "test"),
			},
			err: model.ErrConvertMetricFromString,
		},
		{
			name:   "counter metric is filled and kind is right",
			fields: _fields,
			args: args{
				metric: metric.NewRawMetric("test", keyCounter, "144"),
			},
			err: nil,
		},
		{
			name:   "counter metric is filled and kind is right, value has wrong type",
			fields: _fields,
			args: args{
				metric: metric.NewRawMetric("test", keyCounter, "test"),
			},
			err: model.ErrConvertMetricFromString,
		},
		{
			name:   "unknown kind of metric",
			fields: _fields,
			args: args{
				metric: metric.NewRawMetric("test", "test", "test"),
			},
			err: service.ErrUnknownKind,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStorageService(tt.fields.gaugeStorage, tt.fields.counterStorage)
			err := s.Save(tt.args.metric)
			if tt.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, s.Save(tt.args.metric), tt.err.Error())
			}
		})
	}
}

func TestStorageService_checkInput(t *testing.T) {
	type fields struct {
		gaugeStorage   repository.MetricRepository[model.Gauge]
		counterStorage repository.MetricRepository[model.Counter]
	}

	_fields := fields{
		gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
		counterStorage: memory.NewMetricCounterMemoryRepository(),
	}

	type args struct {
		data metric.RawMetric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		err    error
	}{
		{
			name:   "all is filled",
			fields: _fields,
			args: args{
				data: metric.NewRawMetric("test", keyGauge, "12.45"),
			},
			err: nil,
		},
		{
			name:   "empty name",
			fields: _fields,
			args: args{
				data: metric.NewRawMetric("", keyGauge, "12.45"),
			},
			err: service.ErrEmptyName,
		},
		{
			name:   "empty kind",
			fields: _fields,
			args: args{
				data: metric.NewRawMetric("test", "", "12.45"),
			},
			err: service.ErrEmptyKind,
		},
		{
			name:   "empty value",
			fields: _fields,
			args: args{
				data: metric.NewRawMetric("test", keyGauge, ""),
			},
			err: service.ErrEmptyValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StorageService{
				gaugeStorage:   tt.fields.gaugeStorage,
				counterStorage: tt.fields.counterStorage,
			}
			err := s.checkInput(tt.args.data)

			if tt.err == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.err.Error())
			}
		})
	}
}

func TestStorageService_getSaverByType(t *testing.T) {
	type fields struct {
		gaugeStorage   repository.MetricRepository[model.Gauge]
		counterStorage repository.MetricRepository[model.Counter]
	}

	_fields := fields{
		gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
		counterStorage: memory.NewMetricCounterMemoryRepository(),
	}

	type args struct {
		data metric.RawMetric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   metricSaver
	}{
		{
			name:   "gauge",
			fields: _fields,
			args: args{
				data: metric.NewRawMetric("test", keyGauge, "12.45"),
			},
			want: metric.NewGaugeSaver(
				metric.NewRawMetric("test", keyGauge, "12.45"),
				_fields.gaugeStorage),
		},
		{
			name:   "counter",
			fields: _fields,
			args: args{
				data: metric.NewRawMetric("test", keyCounter, "12"),
			},
			want: metric.NewCounterSaver(
				metric.NewRawMetric("test", keyCounter, "12"),
				_fields.counterStorage),
		},
		{
			name:   "unknown saver",
			fields: _fields,
			args: args{
				data: metric.NewRawMetric("test", "test", "12"),
			},
			want: unknownSaver{
				kind: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StorageService{
				gaugeStorage:   tt.fields.gaugeStorage,
				counterStorage: tt.fields.counterStorage,
			}
			got := s.getSaverByType(tt.args.data)

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("StorageService.getSaverByType() = %v, want %v", got, tt.want)
			}
		})
	}
}
