package metric

import (
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
			got, err := NewMetric(tt.args.name, tt.args.value, tt.args.metricType)

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
			want:    nil,
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
			want:    nil,
			wantErr: true,
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
