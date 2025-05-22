package http

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
)

func Test_TextRequestMaker(t *testing.T) {
	maker, err := NewTextRequestMaker("http://localhost:8080")
	require.NoError(t, err)

	metric := metric.NewCounterMetric("test", 1)
	req, err := maker.Make(metric)
	require.NoError(t, err)

	require.Equal(t, "http://localhost:8080/counter/test/1", req.URL.String())

}

func Test_JSONRequestMaker(t *testing.T) {

	testCases := []struct {
		name     string
		metric   []metric.Metric
		expected string
	}{
		{
			name:     "counter",
			metric:   []metric.Metric{metric.NewCounterMetric("test", 1)},
			expected: `{"id":"test","type":"counter","delta":1}`,
		},
		{
			name:     "gauge",
			metric:   []metric.Metric{metric.NewGaugeMetric("test", 1.123)},
			expected: `{"id":"test","type":"gauge","value":1.123}`,
		},
		{
			name: "several metrics",
			metric: []metric.Metric{
				metric.NewCounterMetric("test", 1),
				metric.NewGaugeMetric("test", 1.123),
			},
			expected: `[{"id":"test","type":"counter","delta":1},{"id":"test","type":"gauge","value":1.123}]`,
		},
	}

	for _, tt := range testCases {
		jw := NewJSONWriter()

		maker, err := NewJSONRequestMaker("http://localhost:8080", jw)
		require.NoError(t, err)

		req, err := maker.Make(tt.metric...)
		require.NoError(t, err)

		content, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		require.Equal(t, tt.expected, string(content))

	}

}
