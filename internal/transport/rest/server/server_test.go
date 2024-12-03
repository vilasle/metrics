package rest

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestHttpServer_Register(t *testing.T) {
	type args struct {
		path    string
		handler http.HandlerFunc
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "register",
			args: args{
				path:    "/test",
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewHTTPServer(":8080")
			s.Register(tt.args.path, []string{http.MethodGet}, []string{}, tt.args.handler)
		})
	}
}

func TestHttpServer_StartStop(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "start stop",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewHTTPServer(":8080")
			s.Register("/test", []string{}, []string{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){}))

			go func() {
				if err := s.Start(); (err != nil) != tt.wantErr {
					t.Errorf("HttpServer.Start() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			time.Sleep(1 * time.Second)

			require.Equal(t, true, s.IsRunning(), "server should be running")

			time.Sleep(1 * time.Second)

			if err := s.Stop(); (err != nil) != tt.wantErr {
				t.Errorf("HttpServer.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(1 * time.Second)

			require.Equal(t, false, s.IsRunning(), "server should be stopped")
		})
	}
}

func TestHttpServer_StartForceStop(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "start stop",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewHTTPServer(":8080")
			s.Register("/test", []string{}, []string{}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

			go func() {
				if err := s.Start(); (err != nil) != tt.wantErr {
					t.Errorf("HttpServer.Start() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			time.Sleep(1 * time.Second)

			require.Equal(t, true, s.IsRunning(), "server should be running")

			time.Sleep(1 * time.Second)

			if err := s.ForceStop(); (err != nil) != tt.wantErr {
				t.Errorf("HttpServer.Stop() error = %v, wantErr %v", err, tt.wantErr)
			}
			time.Sleep(1 * time.Second)

			require.Equal(t, false, s.IsRunning(), "server should be stopped")
		})
	}
}

func TestHttpServer_IsRunning(t *testing.T) {
	type fields struct {
		srv     *http.Server
		mux     *chi.Mux
		running bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "running",
			fields: fields{
				running: true,
			},
			want: true,
		},
		{
			name: "not running",
			fields: fields{
				running: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := HTTPServer{
				srv:     tt.fields.srv,
				mux:     tt.fields.mux,
				running: tt.fields.running,
			}
			if got := s.IsRunning(); got != tt.want {
				t.Errorf("HttpServer.IsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}
