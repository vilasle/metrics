package metric

import (
	"fmt"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetric(t *testing.T) {
	type args struct {
		name       string
		value      string
		metricType string
	}

	tests := []struct {
		name    string
		args    args
		want    Metric
		wantErr bool
	}{
		{
			name: "right gauge metric",
			args: args{
				name:       "test",
				value:      "1",
				metricType: TypeGauge,
			},
			want: &gauge{
				name:  "test",
				value: 1,
			},
			wantErr: false,
		},
		{
			name: "right zero gauge metric",
			args: args{
				name:       "test",
				value:      "0",
				metricType: TypeGauge,
			},
			want: &gauge{
				name:  "test",
				value: 0,
			},
			wantErr: false,
		},
		{
			name: "right negative gauge metric",
			args: args{
				name:       "test",
				value:      "-1032",
				metricType: TypeGauge,
			},
			want: &gauge{
				name:  "test",
				value: -1032,
			},
			wantErr: false,
		},
		{
			name: "right a huge gauge metric",
			args: args{
				name:       "test",
				value:      "2340230423.2342342",
				metricType: TypeGauge,
			},
			want: &gauge{
				name:  "test",
				value: 2340230423.2342342,
			},
			wantErr: false,
		},
		{
			name: "invalid gauge metric: value",
			args: args{
				name:       "test",
				value:      "dfsgkr32423",
				metricType: TypeGauge,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid gauge metric: empty name",
			args: args{
				name:       "",
				value:      "12.321",
				metricType: TypeGauge,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid gauge metric: empty value",
			args: args{
				name:       "test",
				value:      "",
				metricType: TypeGauge,
			},
			want:    nil,
			wantErr: true,
		},

		{
			name: "right counter metric",
			args: args{
				name:       "test",
				value:      "1",
				metricType: TypeCounter,
			},
			want: &counter{
				name:  "test",
				value: 1,
			},
			wantErr: false,
		},
		{
			name: "right zero counter metric",
			args: args{
				name:       "test",
				value:      "0",
				metricType: TypeCounter,
			},
			want: &counter{
				name:  "test",
				value: 0,
			},
			wantErr: false,
		},
		{
			name: "right negative counter metric",
			args: args{
				name:       "test",
				value:      "-1032",
				metricType: TypeCounter,
			},
			want: &counter{
				name:  "test",
				value: -1032,
			},
			wantErr: false,
		},
		{
			name: "right a huge counter metric",
			args: args{
				name:       "test",
				value:      "2340230423",
				metricType: TypeGauge,
			},
			want: &gauge{
				name:  "test",
				value: 2340230423,
			},
			wantErr: false,
		},
		{
			name: "invalid counter metric: value",
			args: args{
				name:       "test",
				value:      "dfsgkr32423",
				metricType: TypeCounter,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid counter metric: empty name",
			args: args{
				name:       "",
				value:      "12",
				metricType: TypeCounter,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid counter metric: empty value",
			args: args{
				name:       "test",
				value:      "",
				metricType: TypeCounter,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid counter metric: invalid type",
			args: args{
				name:       "test",
				value:      "123",
				metricType: TypeGauge + TypeCounter,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMetric(tt.args.name, tt.args.value, tt.args.metricType)

			if tt.wantErr {
				assert.Error(t, err)
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateSummedCounter(t *testing.T) {
	type args struct {
		name    string
		metrics []Metric
	}
	tests := []struct {
		name    string
		args    args
		want    Metric
		wantErr bool
	}{
		{
			name: "single value, must be 100",
			args: args{
				name: "test",
				metrics: []Metric{
					&counter{
						name:  "test",
						value: 100,
					},
				},
			},
			want: &counter{
				name:  "test",
				value: 100,
			},
			wantErr: false,
		},
		{
			name: "several values, must be 677",
			args: args{
				name: "test",
				metrics: []Metric{
					&counter{
						name:  "test",
						value: 100,
					},
					&counter{
						name:  "test",
						value: 50,
					},
					&counter{
						name:  "test",
						value: 50,
					},
					&counter{
						name:  "test",
						value: 25,
					},
					&counter{
						name:  "test",
						value: 25,
					},
					&counter{
						name:  "test",
						value: 25,
					},
					&counter{
						name:  "test",
						value: 25,
					},
					&counter{
						name:  "test",
						value: 100,
					},
					&counter{
						name:  "test",
						value: 50,
					},
					&counter{
						name:  "test",
						value: 50,
					},
					&counter{
						name:  "test",
						value: 25,
					},
					&counter{
						name:  "test",
						value: 25,
					},
					&counter{
						name:  "test",
						value: 25,
					},
					&counter{
						name:  "test",
						value: 25,
					},
					&counter{
						name:  "test",
						value: 35,
					},
					&counter{
						name:  "test",
						value: 35,
					},
					&counter{
						name:  "test",
						value: 7,
					},
				},
			},
			want: &counter{
				name:  "test",
				value: 677,
			},
			wantErr: false,
		},
		{
			name: "several values, with uncorrected type of metric",
			args: args{
				name: "test",
				metrics: []Metric{
					&counter{
						name:  "test",
						value: 100,
					},
					&gauge{
						name:  "test",
						value: 100,
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateSummedCounter(tt.args.name, tt.args.metrics)
			if tt.wantErr {
				assert.Error(t, err)
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMetric() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestFromJSON(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Metric
		wantErr bool
	}{
		{
			name: "correct counter metric",
			args: args{
				content: []byte(`{"id":"test1","type":"counter","delta": 1231}`),
			},
			want: &counter{
				name:  "test1",
				value: 1231,
			},
			wantErr: false,
		},
		{
			name: "uncorrected counter metric",
			args: args{
				content: []byte(`{"id":"test1","type":"counter","value": 1231}`),
			},
			want: &counter{
				name:  "test1",
				value: 0,
			},
			wantErr: true,
		},
		{
			name: "correct gauge metric",
			args: args{
				content: []byte(`{"id":"test1","type":"gauge","value": 1231.12312}`),
			},
			want: &gauge{
				name:  "test1",
				value: 1231.12312,
			},
			wantErr: false,
		},
		{
			name: "uncorrected gauge metric",
			args: args{
				content: []byte(`{"id":"test1","type":"gauge","delta": 1231}`),
			},
			want: &gauge{
				name:  "test1",
				value: 0,
			},
			wantErr: false,
		},
		{
			name: "invalid metric: empty name",
			args: args{
				content: []byte(`{"id":"","type":"gauge","value": 1231.213213}`),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid metric: unknown type",
			args: args{
				content: []byte(`{"id":"test1","type":"temp","value": 1231.213213}`),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid body",
			args: args{
				content: []byte(`{"id":"","type":"temp","value": "1231.213213"}`),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromJSON(tt.args.content)
			if tt.wantErr {
				assert.Error(t, err)
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromJSONArray(t *testing.T) {

	tests := []struct {
		name    string
		input   []byte
		want    []Metric
		wantErr bool
	}{
		{
			name: "success parsing",
			input: []byte(`
				[
					{
						"id": "gauge1",
						"type": "gauge",
						"value": 1.15
					},
					{
						"id": "gauge2",
						"type": "gauge",
						"value": 1123.15
					},
					{
						"id": "counter1",
						"type": "counter",
						"delta": 15
					},
					{
						"id": "counter2",
						"type": "counter",
						"delta": 15234
					}
				]`),
			want: []Metric{
				&gauge{
					name:  "gauge1",
					value: 1.15,
				},
				&gauge{
					name:  "gauge2",
					value: 1123.15,
				},
				&counter{
					name:  "counter1",
					value: 15,
				},
				&counter{
					name:  "counter2",
					value: 15234,
				},
			},
		},
		{
			name: "wrong json",
			input: []byte(`
				[
					{
						"id": "gauge1",
						"type": "gauge",
						"value": 1.15
					},
					
						"id": "gauge2",
						"type": "gauge",
						"value": 1123.15
					},
					{
						"id: "counter1",
						"type": "counter",
						"delta": 15
					}
					{
						"id": "counter2",
						"type": "counter",
						"delta": 15234
					}
				]`),
			want:    []Metric{},
			wantErr: true,
		},
		{
			name: "empty names ",
			input: []byte(`
				[
					{
						"id": "",
						"type": "gauge",
						"value": 1.15
					},
					{
						"id": "",
						"type": "gauge",
						"value": 1123.15
					},
					{
						"id": "counter1",
						"type": "counter",
						"delta": 15
					},
					{
						"id": "counter2",
						"type": "counter",
						"delta": 15234
					}
				]`),
			want:    []Metric{},
			wantErr: true,
		},
		{
			name: "wrong value fields",
			input: []byte(`
				[
					{
						"id": "",
						"type": "gauge",
						"delta": 1
					},
					{
						"id": "",
						"type": "gauge",
						"delta": 1123
					},
					{
						"id": "counter1",
						"type": "counter",
						"value": 15
					},
					{
						"id": "counter2",
						"type": "counter",
						"value": 15234
					}
				]`),
			want:    []Metric{},
			wantErr: true,
		},
		{
			name: "unknown type",
			input: []byte(`
				[
					{
						"id": "gauge1",
						"type": "test",
						"value": 1.15
					},
					{
						"id": "gauge2",
						"type": "auge",
						"value": 1123.15
					}
				]`),
			want:    []Metric{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics, err := FromJSONArray([]byte(tt.input))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, metrics)
		})
	}

}

func Benchmark_ParseMetric(b *testing.B) {

	metrics := make([]struct {
		name       string
		value      string
		metricType string
	}, 2000)

	qtyG := 1000
	qtyC := 1000

	for i := 0; i < qtyG; i++ {
		metrics[i].name = fmt.Sprintf("gauge%d", i)
		metrics[i].value = fmt.Sprintf("%f", rand.Float64())
		metrics[i].metricType = TypeGauge
	}

	for i := 0; i < qtyC; i++ {
		metrics[i+qtyG].name = fmt.Sprintf("counter%d", i)
		metrics[i+qtyG].value = fmt.Sprintf("%d", rand.Int64())
		metrics[i+qtyG].metricType = TypeCounter
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, metric := range metrics {
			ParseMetric(metric.name, metric.value, metric.metricType)
		}
	}
}
