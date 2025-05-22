package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
)

func TestNewHTTPSender(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tm, err := NewTextRequestMaker(server.URL)
	require.NoError(t, err)

	sender := NewHTTPSender(tm)

	err = sender.Send(metric.NewCounterMetric("test", 1))
	require.NoError(t, err)
}
