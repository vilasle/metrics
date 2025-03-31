package metric

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_gauge_Name(t *testing.T) {
	type fields struct {
		name  string
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "getting gauge name",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gauge{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			if got := c.Name(); got != tt.want {
				t.Errorf("gauge.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gauge_Value(t *testing.T) {
	type fields struct {
		name  string
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "getting gauge value (1.0)",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			want: "1",
		},
		{
			name: "getting gauge value (-3121.0)",
			fields: fields{
				name:  "test",
				value: -3121.0,
			},
			want: "-3121",
		},
		{
			name: "getting gauge value (3121.123123)",
			fields: fields{
				name:  "test",
				value: 3121.123123,
			},
			want: "3121.123123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gauge{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			if got := c.Value(); got != tt.want {
				t.Errorf("gauge.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gauge_Type(t *testing.T) {
	type fields struct {
		name  string
		value float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "getting gauge type",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			want: "gauge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gauge{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			if got := c.Type(); got != tt.want {
				t.Errorf("gauge.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gauge_AddValue(t *testing.T) {
	type fields struct {
		name  string
		value float64
	}
	type args struct {
		val any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		{
			name: "add value to gauge (1.0)",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			args: args{
				val: 2.0,
			},
			want: 3.0,
		},
		{
			name: "add value to gauge: incorrect value",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			args: args{
				val: int64(2),
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gauge{
				name:  tt.fields.name,
				value: tt.fields.value,
			}

			err := c.AddValue(tt.args.val)

			if tt.wantErr {
				assert.Error(t, err)
			} else if !reflect.DeepEqual(c.value, tt.want) {
				t.Errorf("gauge.AddValue() = %v, want %v", c.value, tt.want)
			}
		})
	}
}

func Test_gauge_SetValue(t *testing.T) {
	type fields struct {
		name  string
		value float64
	}
	type args struct {
		val any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    float64
	}{
		{
			name: "set gauge value(2.0)",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			args: args{
				val: 2.0,
			},
			want: 2.0,
		},
		{
			name: "set gauge value(234.324234)",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			args: args{
				val: 234.324234,
			},
			want: 234.324234,
		},
		{
			name: "set gauge value(invalid value)",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			args: args{
				val: int64(234),
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gauge{
				name:  tt.fields.name,
				value: tt.fields.value,
			}

			err := c.SetValue(tt.args.val)
			if tt.wantErr {
				assert.Error(t, err)
			} else if !reflect.DeepEqual(c.value, tt.want) {
				t.Errorf("gauge.AddValue() = %v, want %v", c.value, tt.want)
			}
		})
	}
}

func Test_gauge_ToJSON(t *testing.T) {
	type fields struct {
		name  string
		value float64
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "serializing gauge",
			fields: fields{
				name:  "test",
				value: 123456,
			},
			want:    []byte(`{"id":"test","type":"gauge","value":123456}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gauge{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			got, err := json.Marshal(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("gauge.ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("gauge.ToJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_gauge_String(t *testing.T) {
	tests := []struct {
		name    string
		metrics []gauge
		wants   []string
	}{
		{
			name: "gauge string format",
			metrics: []gauge{
				{
					name:  "test1",
					value: 123456,
				},
				{
					name:  "test2",
					value: 123.456,
				},
				{
					name:  "test3",
					value: 1.23456,
				},
			},
			wants: []string{
				"{type: gauge; name: test1; value: 123456.000000}",
				"{type: gauge; name: test2; value: 123.456000}",
				"{type: gauge; name: test3; value: 1.234560}",
			},
		},
	}

	for _, tt := range tests {
		for i, m := range tt.metrics {
			got := m.String()
			assert.Equal(t, tt.wants[i], got)
		}
	}
}

func Test_gauge_Float64(t *testing.T){
	tests := []struct {
		name    string
		metrics []gauge
		wants   []float64
	}{
		{
			name: "gauge get float64 value",
			metrics: []gauge{
				{
					name:  "test1",
					value: 123456,
				},
				{
					name:  "test2",
					value: 123.456,
				},
				{
					name:  "test3",
					value: 1.23456,
				},
			},
			wants: []float64{123456,123.456,1.23456},
		},
	}

	for _, tt := range tests {
		for i, m := range tt.metrics {
			assert.Equal(t, tt.wants[i], m.Float64())
		}
	}
}


func Test_gauge_Int64(t *testing.T){
	tests := []struct {
		name    string
		metrics []gauge
		wants   []int64
	}{
		{
			name: "gauge get int64 value",
			metrics: []gauge{
				{
					name:  "test1",
					value: 123456,
				},
				{
					name:  "test2",
					value: 123.456,
				},
				{
					name:  "test3",
					value: 1.23456,
				},
			},
			wants: []int64{123456,123,1},
		},
	}

	for _, tt := range tests {
		for i, m := range tt.metrics {
			assert.Equal(t, tt.wants[i], m.Int64())
		}
	}
}
