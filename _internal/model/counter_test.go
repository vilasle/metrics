package model

import (
	"testing"
)

func TestCounter_Type(t *testing.T) {
	tests := []struct {
		name string
		c    Counter
		want string
	}{
		{
			name: "counter type",
			c:    45,
			want: "counter",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Type(); got != tt.want {
				t.Errorf("Counter.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCounter_Value(t *testing.T) {
	tests := []struct {
		name string
		c    Counter
		want string
	}{
		{
			name: "counter 156",
			c:    156,
			want: "156",
		},
		{
			name: "counter 265",
			c:    265,
			want: "265",
		},
		{
			name: "counter 1",
			c:    1,
			want: "1",
		},
		{
			name: "counter 1560",
			c:    1560,
			want: "1560",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Value(); got != tt.want {
				t.Errorf("Counter.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCounterFromString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Counter
		wantErr bool
	}{
		{
			name:    "valid counter",
			args:    args{"1465645"},
			want:    Counter(1465645),
			wantErr: false,
		},
		{
			name:    "invalid counter",
			args:    args{"14gd65g64d5g"},
			want:    Counter(0),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CounterFromString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("CounterFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CounterFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}
