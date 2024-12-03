package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vilasle/metrics/internal/model"
)

func Test_NewMetricGaugeMemoryRepository(t *testing.T) {
	_ = NewMetricGaugeMemoryRepository()
}

func Test_NewMetricCounterMemoryRepository(t *testing.T) {
	_ = NewMetricCounterMemoryRepository()
}

func TestGaugeStorage_Save(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value []float64
		want  model.Gauge
	}{
		{
			name:  "save one metric",
			key:   "test",
			value: []float64{10},
			want:  model.Gauge(10),
		},
		{
			name:  "save several metric",
			key:   "test",
			value: []float64{10.01, 15.14, -1.0146, -0, 17.05},
			want:  model.Gauge(17.05),
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMetricGaugeMemoryRepository()
			for _, v := range tt.value {
				r.Save(tt.key, model.Gauge(v))
			}
			got := r.metrics[tt.key]
			assert.Equal(t, tt.want, got)
		})
	}
	r := NewMetricGaugeMemoryRepository()
	r.Save("test", 10)

}

func TestCounterStorage_Save(t *testing.T) {
	testCases := []struct {
		name  string
		key   string
		value []model.Counter
	}{
		{
			name:  "save one metric",
			key:   "test",
			value: []model.Counter{10},
		},
		{
			name:  "save several metric",
			key:   "test",
			value: []model.Counter{10, 15, -1, 0, 17},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r := NewMetricCounterMemoryRepository()
			for _, v := range tt.value {
				r.Save(tt.key, v)
			}
			assert.ElementsMatch(t, tt.value, r.metrics[tt.key])
		})
	}
}
