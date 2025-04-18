package text

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
)

func Test_HTTPTextSender_Send(t *testing.T) {
	tests := []struct {
		name      string
		handler   http.Handler
		wantError bool
		metrics   metric.Metric
	}{
		{
			name: "success",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			wantError: false,
			metrics:   metric.NewGaugeMetric("test", 134.5),
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

			sender, err := NewHTTPTextSender(server.URL)
			require.NoError(t, err)

			err = sender.Send(tt.metrics)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}
}