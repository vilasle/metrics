package server

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"github.com/vilasle/metrics/internal/metric"
// 	"github.com/vilasle/metrics/internal/repository/memory"
// 	"github.com/vilasle/metrics/internal/service"
// )

// type rawMetric struct {
// 	Name  string
// 	Type  string
// 	Value string
// }

// func TestStorageService_Save(t *testing.T) {
// 	type fields struct {
// 		storage *memory.MemoryMetricRepository
// 	}

// 	_fields := fields{
// 		storage: memory.NewMetricRepository(),
// 	}

// 	type args struct {
// 		metric rawMetric
// 	}

// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args
// 		err error
// 	}{
// 		{
// 			name:   "problems with input, did not fill name of value",
// 			fields: _fields,
// 			args: args{
// 				metric: rawMetric{"", metric.TypeGauge, "12.45"},
// 			},
// 			err: service.ErrEmptyName,
// 		},
// 		{
// 			name:   "problems with input, did not fill kind of metric",
// 			fields: _fields,
// 			args: args{
// 				metric: rawMetric{"test", "", "12.45"},
// 			},
// 			err: service.ErrEmptyKind,
// 		},
// 		{
// 			name:   "problems with input, did not fill value of metric",
// 			fields: _fields,
// 			args: args{
// 				metric: rawMetric{"test", metric.TypeGauge, ""},
// 			},
// 			err: service.ErrEmptyValue,
// 		},
// 		{
// 			name:   "metric is filled but contents unknown kind",
// 			fields: _fields,
// 			args: args{
// 				metric: rawMetric{"test", "test", "123.4"},
// 			},
// 			err: service.ErrUnknownKind,
// 		},
// 		{
// 			name:   "gauge metric is filled and kind is right",
// 			fields: _fields,
// 			args: args{
// 				metric: rawMetric{"test", metric.TypeGauge, "123.4"},
// 			},
// 			err: nil,
// 		},
// 		{
// 			name:   "gauge metric is filled and kind is right, value has wrong type",
// 			fields: _fields,
// 			args: args{
// 				metric: rawMetric{"test", metric.TypeGauge, "test"},
// 			},
// 			err: model.ErrConvertMetricFromString,
// 		},
// 		{
// 			name:   "counter metric is filled and kind is right",
// 			fields: _fields,
// 			args: args{
// 				metric: rawMetric{"test", metric.TypeCounter, "144"},
// 			},
// 			err: nil,
// 		},
// 		{
// 			name:   "counter metric is filled and kind is right, value has wrong type",
// 			fields: _fields,
// 			args: args{
// 				metric: rawMetric{"test", metric.TypeCounter, "test"},
// 			},
// 			err: model.ErrConvertMetricFromString,
// 		},
// 		{
// 			name:   "unknown kind of metric",
// 			fields: _fields,
// 			args: args{
// 				metric: rawMetric{"test", "test", "test"},
// 			},
// 			err: service.ErrUnknownKind,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			s := NewMetricService(tt.fields.storage)

// 			m, err := metric.NewMetric(tt.args.metric.Name, tt.args.metric.Value, tt.args.metric.Type)
// 			require.NoError(t, err)

// 			err = s.Save(m)
// 			if tt.err == nil {
// 				assert.NoError(t, err)
// 			} else {
// 				assert.ErrorContains(t, err, tt.err.Error())
// 			}
// 		})
// 	}
// }

// func TestStorageService_checkInput(t *testing.T) {
// 	// type fields struct {
// 	// 	gaugeStorage   repository.MetricRepository[model.Gauge]
// 	// 	counterStorage repository.MetricRepository[model.Counter]
// 	// }

// 	// _fields := fields{
// 	// 	gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
// 	// 	counterStorage: memory.NewMetricCounterMemoryRepository(),
// 	// }

// 	// type args struct {
// 	// 	data metric.RawMetric
// 	// }
// 	// tests := []struct {
// 	// 	name   string
// 	// 	fields fields
// 	// 	args   args
// 	// 	err    error
// 	// }{
// 	// 	{
// 	// 		name:   "all is filled",
// 	// 		fields: _fields,
// 	// 		args: args{
// 	// 			data: metric.NewRawMetric("test", metric.TypeGauge, "12.45"),
// 	// 		},
// 	// 		err: nil,
// 	// 	},
// 	// 	{
// 	// 		name:   "empty name",
// 	// 		fields: _fields,
// 	// 		args: args{
// 	// 			data: metric.NewRawMetric("", metric.TypeGauge, "12.45"),
// 	// 		},
// 	// 		err: service.ErrEmptyName,
// 	// 	},
// 	// 	{
// 	// 		name:   "empty kind",
// 	// 		fields: _fields,
// 	// 		args: args{
// 	// 			data: metric.NewRawMetric("test", "", "12.45"),
// 	// 		},
// 	// 		err: service.ErrEmptyKind,
// 	// 	},
// 	// 	{
// 	// 		name:   "empty value",
// 	// 		fields: _fields,
// 	// 		args: args{
// 	// 			data: metric.NewRawMetric("test", metric.TypeGauge, ""),
// 	// 		},
// 	// 		err: service.ErrEmptyValue,
// 	// 	},
// 	// }

// 	// for _, tt := range tests {
// 	// 	t.Run(tt.name, func(t *testing.T) {
// 	// 		s := StorageService{
// 	// 			gaugeStorage:   tt.fields.gaugeStorage,
// 	// 			counterStorage: tt.fields.counterStorage,
// 	// 		}
// 	// 		err := s.checkInput(tt.args.data)

// 	// 		if tt.err == nil {
// 	// 			assert.NoError(t, err)
// 	// 		} else {
// 	// 			assert.ErrorContains(t, err, tt.err.Error())
// 	// 		}
// 	// 	})
// 	// }
// }

// func TestStorageService_getSaverByType(t *testing.T) {
// 	// type fields struct {
// 	// 	gaugeStorage   repository.MetricRepository[model.Gauge]
// 	// 	counterStorage repository.MetricRepository[model.Counter]
// 	// }

// 	// _fields := fields{
// 	// 	gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
// 	// 	counterStorage: memory.NewMetricCounterMemoryRepository(),
// 	// }

// 	// type args struct {
// 	// 	data metric.RawMetric
// 	// }
// 	// tests := []struct {
// 	// 	name   string
// 	// 	fields fields
// 	// 	args   args
// 	// 	want   metricSaver
// 	// }{
// 	// 	{
// 	// 		name:   "gauge",
// 	// 		fields: _fields,
// 	// 		args: args{
// 	// 			data: metric.NewRawMetric("test", metric.TypeGauge, "12.45"),
// 	// 		},
// 	// 		want: metric.NewGaugeSaver(
// 	// 			metric.NewRawMetric("test", metric.TypeGauge, "12.45"),
// 	// 			_fields.gaugeStorage),
// 	// 	},
// 	// 	{
// 	// 		name:   "counter",
// 	// 		fields: _fields,
// 	// 		args: args{
// 	// 			data: metric.NewRawMetric("test", metric.TypeCounter, "12"),
// 	// 		},
// 	// 		want: metric.NewCounterSaver(
// 	// 			metric.NewRawMetric("test", metric.TypeCounter, "12"),
// 	// 			_fields.counterStorage),
// 	// 	},
// 	// 	{
// 	// 		name:   "unknown saver",
// 	// 		fields: _fields,
// 	// 		args: args{
// 	// 			data: metric.NewRawMetric("test", "test", "12"),
// 	// 		},
// 	// 		want: unknownSaver{
// 	// 			kind: "test",
// 	// 		},
// 	// 	},
// 	// }

// 	// for _, tt := range tests {
// 	// 	t.Run(tt.name, func(t *testing.T) {
// 	// 		s := StorageService{
// 	// 			gaugeStorage:   tt.fields.gaugeStorage,
// 	// 			counterStorage: tt.fields.counterStorage,
// 	// 		}
// 	// 		got := s.getSaverByType(tt.args.data)

// 	// 		if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
// 	// 			t.Errorf("StorageService.getSaverByType() = %v, want %v", got, tt.want)
// 	// 		}
// 	// 	})
// 	// }
// }

// func TestStorageService_AllMetrics(t *testing.T) {
// 	// type storage struct {
// 	// 	gaugeStorage   repository.MetricRepository[model.Gauge]
// 	// 	counterStorage repository.MetricRepository[model.Counter]
// 	// }

// 	// _storage := storage{
// 	// 	gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
// 	// 	counterStorage: memory.NewMetricCounterMemoryRepository(),
// 	// }

// 	// _storage.counterStorage.Save("counter1", 15)
// 	// _storage.counterStorage.Save("counter2", 55)
// 	// _storage.counterStorage.Save("counter3", 75)
// 	// _storage.counterStorage.Save("counter4", 15)
// 	// _storage.counterStorage.Save("counter5", 145325)
// 	// _storage.counterStorage.Save("counter6", 43243)

// 	// _storage.gaugeStorage.Save("gauge1", 155.41)
// 	// _storage.gaugeStorage.Save("gauge2", 535.123)
// 	// _storage.gaugeStorage.Save("gauge3", 75.5344213)
// 	// _storage.gaugeStorage.Save("gauge4", 12315.123)
// 	// _storage.gaugeStorage.Save("gauge5", 1554.131)

// 	// testCases := []struct {
// 	// 	name    string
// 	// 	args    storage
// 	// 	wantLen int
// 	// }{
// 	// 	{
// 	// 		name: "Len must equal 11",
// 	// 		args: storage{
// 	// 			gaugeStorage:   _storage.gaugeStorage,
// 	// 			counterStorage: _storage.counterStorage,
// 	// 		},
// 	// 		wantLen: 11,
// 	// 	},
// 	// 	{
// 	// 		name: "Len must equal 5",
// 	// 		args: storage{
// 	// 			gaugeStorage:   _storage.gaugeStorage,
// 	// 			counterStorage: memory.NewMetricCounterMemoryRepository(),
// 	// 		},
// 	// 		wantLen: 5,
// 	// 	},
// 	// 	{
// 	// 		name: "Len must equal 6",
// 	// 		args: storage{
// 	// 			gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
// 	// 			counterStorage: _storage.counterStorage,
// 	// 		},
// 	// 		wantLen: 6,
// 	// 	},
// 	// 	{
// 	// 		name: "Len must equal 0",
// 	// 		args: storage{
// 	// 			gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
// 	// 			counterStorage: memory.NewMetricCounterMemoryRepository(),
// 	// 		},
// 	// 		wantLen: 0,
// 	// 	},
// 	// }

// 	// for _, tt := range testCases {
// 	// 	t.Run(tt.name, func(t *testing.T) {
// 	// 		s := NewStorageService(tt.args.gaugeStorage, tt.args.counterStorage)

// 	// 		all, err := s.AllMetrics()
// 	// 		require.NoError(t, err)
// 	// 		assert.Len(t, all, tt.wantLen)
// 	// 	})
// 	// }
// }

// func TestStorageService_Get(t *testing.T) {
// 	// type storage struct {
// 	// 	gaugeStorage   repository.MetricRepository[model.Gauge]
// 	// 	counterStorage repository.MetricRepository[model.Counter]
// 	// }

// 	// _storage := storage{
// 	// 	gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
// 	// 	counterStorage: memory.NewMetricCounterMemoryRepository(),
// 	// }

// 	// _storage.counterStorage.Save("counter1", 15)
// 	// _storage.counterStorage.Save("counter2", 55)
// 	// _storage.counterStorage.Save("counter3", 75)
// 	// _storage.counterStorage.Save("counter4", 15)
// 	// _storage.counterStorage.Save("counter5", 145325)
// 	// _storage.counterStorage.Save("counter6", 43243)

// 	// _storage.gaugeStorage.Save("gauge1", 155.41)
// 	// _storage.gaugeStorage.Save("gauge2", 535.123)
// 	// _storage.gaugeStorage.Save("gauge3", 75.5344213)
// 	// _storage.gaugeStorage.Save("gauge4", 12315.123)
// 	// _storage.gaugeStorage.Save("gauge5", 1554.131)

// 	// testCases := []struct {
// 	// 	name    string
// 	// 	args    storage
// 	// 	key     string
// 	// 	kind    string
// 	// 	wantErr bool
// 	// 	value   metric.Metric
// 	// }{
// 	// 	{
// 	// 		name: "get existed counter",
// 	// 		args: storage{
// 	// 			gaugeStorage:   _storage.gaugeStorage,
// 	// 			counterStorage: _storage.counterStorage,
// 	// 		},
// 	// 		key:     "counter1",
// 	// 		kind:    metric.TypeCounter,
// 	// 		wantErr: false,
// 	// 		value:   metric.NewCounterMetric("counter1", 15),
// 	// 	},
// 	// 	{
// 	// 		name: "get not existed counter",
// 	// 		args: storage{
// 	// 			gaugeStorage:   _storage.gaugeStorage,
// 	// 			counterStorage: _storage.counterStorage,
// 	// 		},
// 	// 		key:     "counter543",
// 	// 		kind:    metric.TypeCounter,
// 	// 		wantErr: true,
// 	// 		value:   nil,
// 	// 	},
// 	// 	{
// 	// 		name: "get existed gauge",
// 	// 		args: storage{
// 	// 			gaugeStorage:   _storage.gaugeStorage,
// 	// 			counterStorage: _storage.counterStorage,
// 	// 		},
// 	// 		key:     "gauge3",
// 	// 		kind:    metric.TypeGauge,
// 	// 		wantErr: false,
// 	// 		value:   metric.NewGaugeMetric("gauge3", 75.5344213),
// 	// 	},
// 	// 	{
// 	// 		name: "get not existed gauge",
// 	// 		args: storage{
// 	// 			gaugeStorage:   _storage.gaugeStorage,
// 	// 			counterStorage: _storage.counterStorage,
// 	// 		},
// 	// 		key:     "gauge543",
// 	// 		kind:    metric.TypeGauge,
// 	// 		wantErr: true,
// 	// 		value:   nil,
// 	// 	},
// 	// 	{
// 	// 		name: "use unknown kind of metrics",
// 	// 		args: storage{
// 	// 			gaugeStorage:   _storage.gaugeStorage,
// 	// 			counterStorage: _storage.counterStorage,
// 	// 		},
// 	// 		key:     "test543",
// 	// 		kind:    "someMetric",
// 	// 		wantErr: true,
// 	// 		value:   nil,
// 	// 	},
// 	// 	{
// 	// 		name: "empty counter storage",
// 	// 		args: storage{
// 	// 			gaugeStorage:   _storage.gaugeStorage,
// 	// 			counterStorage: memory.NewMetricCounterMemoryRepository(),
// 	// 		},
// 	// 		key:     "test543",
// 	// 		kind:    metric.TypeCounter,
// 	// 		wantErr: true,
// 	// 		value:   nil,
// 	// 	},
// 	// 	{
// 	// 		name: "empty gauge storage",
// 	// 		args: storage{
// 	// 			gaugeStorage:   memory.NewMetricGaugeMemoryRepository(),
// 	// 			counterStorage: _storage.counterStorage,
// 	// 		},
// 	// 		key:     "test543",
// 	// 		kind:    metric.TypeGauge,
// 	// 		wantErr: true,
// 	// 		value:   nil,
// 	// 	},
// 	// }

// 	// for _, tt := range testCases {
// 	// 	t.Run(tt.name, func(t *testing.T) {
// 	// 		s := NewStorageService(tt.args.gaugeStorage, tt.args.counterStorage)

// 	// 		v, err := s.Get(tt.key, tt.kind)
// 	// 		if tt.wantErr {
// 	// 			require.Error(t, err)
// 	// 			return
// 	// 		}

// 	// 		require.NoError(t, err)
// 	// 		require.NotNil(t, v)
// 	// 		assert.Equal(t, tt.value, v)
// 	// 	})
// 	// }
// }
