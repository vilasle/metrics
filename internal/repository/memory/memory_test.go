package memory

// import (
// 	"sync"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/vilasle/metrics/internal/model"
// )

// func Test_NewMetricGaugeMemoryRepository(t *testing.T) {
// 	_ = NewMetricGaugeMemoryRepository()
// }

// func Test_NewMetricCounterMemoryRepository(t *testing.T) {
// 	_ = NewMetricCounterMemoryRepository()
// }

// func TestGaugeStorage_Save(t *testing.T) {
// 	testCases := []struct {
// 		name  string
// 		key   string
// 		value []float64
// 		want  model.Gauge
// 	}{
// 		{
// 			name:  "save one metric",
// 			key:   "test",
// 			value: []float64{10},
// 			want:  model.Gauge(10),
// 		},
// 		{
// 			name:  "save several metric",
// 			key:   "test",
// 			value: []float64{10.01, 15.14, -1.0146, -0, 17.05},
// 			want:  model.Gauge(17.05),
// 		},
// 	}
// 	for _, tt := range testCases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := NewMetricGaugeMemoryRepository()
// 			for _, v := range tt.value {
// 				r.Save(tt.key, model.Gauge(v))
// 			}
// 			got := r.metrics[tt.key]
// 			assert.Equal(t, tt.want, got)
// 		})
// 	}
// 	r := NewMetricGaugeMemoryRepository()
// 	r.Save("test", 10)
// }

// func TestCounterStorage_Save(t *testing.T) {
// 	testCases := []struct {
// 		name  string
// 		key   string
// 		value []model.Counter
// 	}{
// 		{
// 			name:  "save one metric",
// 			key:   "test",
// 			value: []model.Counter{10},
// 		},
// 		{
// 			name:  "save several metric",
// 			key:   "test",
// 			value: []model.Counter{10, 15, -1, 0, 17},
// 		},
// 	}
// 	for _, tt := range testCases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := NewMetricCounterMemoryRepository()
// 			for _, v := range tt.value {
// 				r.Save(tt.key, v)
// 			}
// 			assert.ElementsMatch(t, tt.value, r.metrics[tt.key])
// 		})
// 	}
// }

// func TestCounterAll(t *testing.T) {
// 	testCases := []struct {
// 		name string
// 		key  string
// 		data map[string][]model.Counter
// 		want model.Counter
// 	}{
// 		{
// 			name: "value have to 15",
// 			key:  "test",
// 			data: map[string][]model.Counter{
// 				"test": {1, 2, 3, 4, 5},
// 			},
// 			want: 15,
// 		},
// 		{
// 			name: "value have to 60",
// 			key:  "test",
// 			data: map[string][]model.Counter{
// 				"test": {1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5, 1, 2, 3, 4, 5},
// 			},
// 			want: 60,
// 		},
// 		{
// 			name: "value have to 0",
// 			key:  "test",
// 			data: map[string][]model.Counter{},
// 			want: 0,
// 		},
// 	}

// 	for _, tt := range testCases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := MetricCounterMemoryRepository[model.Counter]{
// 				metrics: tt.data,
// 				mx:      &sync.RWMutex{},
// 			}
// 			//MetricCounterMemoryRepository does not return error
// 			result, _ := r.All()
// 			assert.Equal(t, tt.want, result[tt.key])
// 		})
// 	}
// }

// func TestCounterGet(t *testing.T) {
// 	testCases := []struct {
// 		name string
// 		key  string
// 		data map[string][]model.Counter
// 		want model.Counter
// 	}{
// 		{
// 			name: "value have to 15",
// 			key:  "test",
// 			data: map[string][]model.Counter{
// 				"test":  {1, 2, 3, 4, 5},
// 				"test1": {1, 2, 3, 4, 5, 10},
// 				"test2": {1, 2, 3, 4, 5, 1, 5},
// 				"test3": {1, 2, 3, 4, 5, 7, 54},
// 				"test4": {1, 2, 3, 4, 5, 1, 5},
// 				"test5": {1, 2, 3, 4, 5, 5, 6},
// 			},
// 			want: 15,
// 		},
// 		{
// 			name: "value have to 0",
// 			key:  "test145",
// 			data: map[string][]model.Counter{
// 				"test":  {1, 2, 3, 4, 5},
// 				"test1": {1, 2, 3, 4, 5, 10},
// 				"test2": {1, 2, 3, 4, 5, 1, 5},
// 				"test3": {1, 2, 3, 4, 5, 7, 54},
// 				"test4": {1, 2, 3, 4, 5, 1, 5},
// 				"test5": {1, 2, 3, 4, 5, 5, 6},
// 			},
// 			want: 0,
// 		},
// 	}

// 	for _, tt := range testCases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := MetricCounterMemoryRepository[model.Counter]{
// 				metrics: tt.data,
// 				mx:      &sync.RWMutex{},
// 			}
// 			//MetricCounterMemoryRepository does not return error
// 			result, _ := r.Get(tt.key)
// 			assert.Equal(t, tt.want, result)
// 		})
// 	}
// }

// func TestGaugeAll(t *testing.T) {
// 	testCases := []struct {
// 		name string
// 		data map[string]model.Gauge
// 		want map[string]model.Gauge
// 	}{
// 		{
// 			name: "with filled metrics",
// 			data: map[string]model.Gauge{
// 				"test1": 3,
// 				"test2": 3123,
// 				"test3": 3245,
// 				"test4": 33.41,
// 				"test5": 32.21,
// 				"test6": 33.33,
// 			},
// 			want: map[string]model.Gauge{
// 				"test1": 3,
// 				"test2": 3123,
// 				"test3": 3245,
// 				"test4": 33.41,
// 				"test5": 32.21,
// 				"test6": 33.33,
// 			},
// 		},
		
// 		{
// 			name: "with empty metrics",
// 			data: map[string]model.Gauge{},
// 			want: map[string]model.Gauge{},
// 		},
// 	}

// 	for _, tt := range testCases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := MetricGaugeMemoryRepository[model.Gauge]{
// 				metrics: tt.data,
// 				mx:      &sync.RWMutex{},
// 			}
// 			//MetricGaugeMemoryRepository does not return error
// 			result, _ := r.All()
// 			assert.Equal(t, tt.want, result)
// 		})
// 	}
// }

// func TestGaugeGet(t *testing.T) {
// 	testCases := []struct {
// 		name string
// 		data map[string]model.Gauge
// 		key string
// 		want model.Gauge
// 	}{
// 		{
// 			name: "with filled metrics",
// 			data: map[string]model.Gauge{
// 				"test1": 3,
// 				"test2": 3123,
// 				"test3": 3245,
// 				"test4": 33.41,
// 				"test5": 32.21,
// 				"test6": 33.33,
// 			},
// 			key: "test1",
// 			want: 3,
// 		},
// 		{
// 			name: "with empty metrics",
// 			data: map[string]model.Gauge{},
// 			key: "test1",
// 			want: 0,
// 		},
// 	}

// 	for _, tt := range testCases {
// 		t.Run(tt.name, func(t *testing.T) {
// 			r := MetricGaugeMemoryRepository[model.Gauge]{
// 				metrics: tt.data,
// 				mx:      &sync.RWMutex{},
// 			}
// 			//MetricGaugeMemoryRepository does not return error
// 			result, _ := r.Get(tt.key)
// 			assert.Equal(t, tt.want, result)
// 		})
// 	}
// }
