package model

import (
	"testing"
)

func TestGauge_Value(t *testing.T) {
	tests := []struct {
		name string
		g    Gauge
		want string
	}{
		{
			name: "gauge value 156.9",
			g: 156.9,
			want: "156.9",
		},
		{
			name: "gauge value 156.00",
			g: 156.00,
			want: "156",
		},
		{
			name: "gauge value 256.00",
			g: 256.00,
			want: "256",
		},
		{
			name: "gauge value 156.00435",
			g: 156.00435,
			want: "156.00435",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.Value(); got != tt.want {
				t.Errorf("Gauge.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGauge_Type(t *testing.T) {
	tests := []struct {
		name string
		g    Gauge
		want string
	}{
		{
			name: "gauge type",
			g: 156.4,
			want: "gauge",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.Type(); got != tt.want {
				t.Errorf("Gauge.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}


func TestGaugeFromString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    Gauge
		wantErr bool
	}{
		{
			name: "valid gauge",
			args: args{
				s: "1234567890.545342",
			},
			want:	Gauge(1234567890.545342),
			wantErr: false,
		},
		{
			name: "invalid gauge",
			args: args{
				s: "1234567890a",
			},
			want:	Gauge(0),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GaugeFromString(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("GaugeFromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GaugeFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

