package metric

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_counter_Name(t *testing.T) {
	type fields struct {
		name  string
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "getting counter name",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := counter{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			if got := c.Name(); got != tt.want {
				t.Errorf("counter.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_counter_Value(t *testing.T) {
	type fields struct {
		name  string
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "getting counter value (1.0)",
			fields: fields{
				name:  "test",
				value: 1,
			},
			want: "1",
		},
		{
			name: "getting counter value (-3121.0)",
			fields: fields{
				name:  "test",
				value: -3121,
			},
			want: "-3121",
		},
		{
			name: "getting counter value (3121)",
			fields: fields{
				name:  "test",
				value: 3121,
			},
			want: "3121",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := counter{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			if got := c.Value(); got != tt.want {
				t.Errorf("counter.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_counter_Type(t *testing.T) {
	type fields struct {
		name  string
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "getting counter type",
			fields: fields{
				name:  "test",
				value: 1.0,
			},
			want: "counter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := counter{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			if got := c.Type(); got != tt.want {
				t.Errorf("counter.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_counter_AddValue(t *testing.T) {
	type fields struct {
		name  string
		value int64
	}
	type args struct {
		val any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "add value to counter (1)",
			fields: fields{
				name:  "test",
				value: 1,
			},
			args: args{
				val: int64(2),
			},
			want: 3,
		},
		{
			name: "add value to counter: incorrect value",
			fields: fields{
				name:  "test",
				value: 1,
			},
			args: args{
				val: float64(2),
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &counter{
				name:  tt.fields.name,
				value: tt.fields.value,
			}

			err := c.AddValue(tt.args.val)

			if tt.wantErr {
				assert.Error(t, err)
			} else if !reflect.DeepEqual(c.value, tt.want) {
				t.Errorf("counter.AddValue() = %v, want %v", c.value, tt.want)
			}
		})
	}
}

func Test_counter_SetValue(t *testing.T) {
	type fields struct {
		name  string
		value int64
	}
	type args struct {
		val any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    int64
	}{
		{
			name: "set counter value(2)",
			fields: fields{
				name:  "test",
				value: 1,
			},
			args: args{
				val: int64(2),
			},
			want: 2,
		},
		{
			name: "set counter value(234)",
			fields: fields{
				name:  "test",
				value: 1,
			},
			args: args{
				val: int64(234),
			},
			want: 234,
		},
		{
			name: "set counter value(invalid value)",
			fields: fields{
				name:  "test",
				value: 1,
			},
			args: args{
				val: float64(234),
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &counter{
				name:  tt.fields.name,
				value: tt.fields.value,
			}

			err := c.SetValue(tt.args.val)
			if tt.wantErr {
				assert.Error(t, err)
			} else if !reflect.DeepEqual(c.value, tt.want) {
				t.Errorf("counter.AddValue() = %v, want %v", c.value, tt.want)
			}
		})
	}
}

func Test_counter_ToJSON(t *testing.T) {
	type fields struct {
		name  string
		value int64
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "serializing counter",
			fields: fields{
				name:  "test",
				value: 123456,
			},
			want:    []byte(`{"id":"test","type":"counter","delta":123456}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := counter{
				name:  tt.fields.name,
				value: tt.fields.value,
			}
			got, err := c.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("counter.ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("counter.ToJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
