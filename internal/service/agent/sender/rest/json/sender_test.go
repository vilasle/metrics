package json

import (
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
)

func TestHTTPSender_Send(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.Handler
		wantError      bool
		metrics        metric.Metric
		hashSumKey     string
		useCompression bool
	}{
		{
			name: "success",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			wantError:      false,
			metrics:        metric.NewGaugeMetric("test", 134.5),
			hashSumKey:     "r312313gfdg32123",
			useCompression: true,
		},
		{
			name: "failure not found",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			wantError: true,
			metrics:   metric.NewGaugeMetric("", 134.5),
		},
		{
			name: "failure bad request",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}),
			wantError: true,
			metrics:   metric.NewGaugeMetric("", 134.5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			u, err := url.Parse(server.URL)
			require.NoError(t, err)

			sender := HTTPJsonSender{
				URL:        u,
				httpClient: newClient(tt.useCompression),
				hashSumKey: tt.hashSumKey,
				req:        make(chan metric.Metric, 1),
				resp:       make(chan error, 1),
				rateLimit:  1,
			}

			err = sender.Send(tt.metrics)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func TestHTTPSender_SendBatch(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.Handler
		wantError  bool
		metrics    []metric.Metric
		hashSumKey string
	}{
		{
			name: "success",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			wantError:  false,
			metrics:    []metric.Metric{metric.NewGaugeMetric("test", 134.5)},
			hashSumKey: "r312313gfdg32123",
		},
		{
			name: "failure not found",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}),
			wantError: true,
			metrics:   []metric.Metric{metric.NewGaugeMetric("", 134.5)},
		},
		{
			name: "failure bad request",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}),
			wantError: true,
			metrics:   []metric.Metric{metric.NewGaugeMetric("", 134.5)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			u, err := url.Parse(server.URL)
			require.NoError(t, err)

			sender := HTTPJsonSender{
				URL:        u,
				httpClient: newClient(false),
				hashSumKey: tt.hashSumKey,
				req:        make(chan metric.Metric, 1),
				resp:       make(chan error, 1),
				rateLimit:  1,
			}

			err = sender.SendBatch(tt.metrics...)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func BenchmarkHTTPSender_prepareBodyForReport(b *testing.B) {
	m := metric.NewGaugeMetric("gauge1", 12.4523)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := prepareBodyForReport(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHTTPSender_prepareBatchBodyForReport(b *testing.B) {
	metrics := make([]metric.Metric, 1000)
	for i := 0; i < 1000; i++ {
		metrics[i] = metric.NewGaugeMetric("gauge1", rand.Float64())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := prepareBatchBodyForReport(metrics...); err != nil {
			b.Fatal(err)
		}
	}
}
