package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository/memory"
	"github.com/vilasle/metrics/internal/service/server"
)

func TestUpdateMetricAsPlainText(t *testing.T) {
	svc := server.NewMetricService(memory.NewMetricRepository())

	cases := []struct {
		name        string
		path        []map[string]string
		method      string
		contentType []string
		statusCode  int
	}{
		{
			name:   "send normal gauge",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "gauge", "name": "test", "value": "1.05"},
				{"type": "gauge", "name": "test1", "value": "1.033"},
				{"type": "gauge", "name": "test", "value": "140.10"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send normal gauge with big value",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "gauge", "name": "test", "value": "43435345435345.343424205"},
				{"type": "gauge", "name": "test1", "value": "4343534234634342.033"},
				{"type": "gauge", "name": "test", "value": "14000000000000.10"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send a gauge with negative value",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "gauge", "name": "test", "value": "-43435345435345.343424205"},
				{"type": "gauge", "name": "test1", "value": "-4343534234634342.033"},
				{"type": "gauge", "name": "test", "value": "-14000000000000.10"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send gauge with string value",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "gauge", "name": "test", "value": "string_value"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send gauge with empty value",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "gauge", "name": "test"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send normal counter",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "counter", "name": "test", "value": "124"},
				{"type": "counter", "name": "test1", "value": "12452"},
				{"type": "counter", "name": "test", "value": "213124"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send normal counter with big value",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "counter", "name": "test", "value": "12400000000000"},
				{"type": "counter", "name": "test1", "value": "1245200000000000000"},
				{"type": "counter", "name": "test", "value": "213124000000000000"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send a counter with negative value",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "counter", "name": "test", "value": "-124"},
				{"type": "counter", "name": "test1", "value": "-12452"},
				{"type": "counter", "name": "test", "value": "-213124"},
				{"type": "counter", "name": "test", "value": "0"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send a counter with string value",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "counter", "name": "test", "value": "sting_value"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send counter with empty value",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "counter", "name": "test"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send unsupported type of metric",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "another_metrics", "name": "test", "value": "1.05"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send metric without name",
			method: http.MethodPost,
			path: []map[string]string{
				{"type": "gauge"},
				{"type": "counter"},
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name:   "send without kind of metric",
			method: http.MethodPost,
			path:   []map[string]string{},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			for _, p := range tt.path {
				ctx := chi.NewRouteContext()

				for k, v := range p {
					ctx.URLParams.Add(k, v)
				}
				req, err := http.NewRequest(http.MethodPost, "/update/{type}/{name}/{value}", nil)
				if err != nil {
					t.Fatal(err)
				}
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
				for _, ct := range tt.contentType {
					req.Header.Set("Content-Type", ct)
				}
				rr := httptest.NewRecorder()
				handler := UpdateMetric(svc)
				handler.ServeHTTP(rr, req)
				if status := rr.Code; status != tt.statusCode {
					t.Errorf("handler returned wrong status code: got %v want %v",
						status, tt.statusCode)
				}
			}
		})
	}
}

func TestDisplayAllMetricsAsHtml(t *testing.T) {
	storage := memory.NewMetricRepository()

	for _, v := range []struct {
		n string
		v string
		t string
	}{
		{"gauge1", "1.05", metric.TypeGauge},
		{"gauge2", "1.15", metric.TypeGauge},
		{"counter1", "2", metric.TypeCounter},
		{"counter2", "3", metric.TypeGauge},
	} {
		m, err := metric.ParseMetric(v.n, v.v, v.t)
		if err != nil {
			t.Fatal(err)
		}

		if err := storage.Save(context.TODO(), m); err != nil {
			t.Fatal(err)
		}
	}

	testCases := []struct {
		name       string
		path       string
		statusCode int
		svc        *server.MetricService
		contents   bool
		exp        string
	}{
		{
			name:       "get several metrics",
			statusCode: http.StatusOK,
			path:       "/",
			svc:        server.NewMetricService(storage),
			contents:   true,
			exp:        `<li>.+<\/li>`,
		},
		{
			name:       "empty storage",
			statusCode: http.StatusOK,
			path:       "/",
			svc:        server.NewMetricService(memory.NewMetricRepository()),
			contents:   false,
			exp:        `<li>.+<\/li>`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.RequestURI = tt.path

			ctx := chi.NewRouteContext()

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			rr := httptest.NewRecorder()
			handler := DisplayAllMetrics(tt.svc)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)

			if rr.Code == http.StatusOK {
				body := rr.Body.String()
				exp := regexp.MustCompile(tt.exp)
				if tt.contents {
					assert.Regexp(t, exp, body)
				} else {
					assert.NotRegexp(t, exp, body)
				}
			}
		})
	}

}

func TestDisplayMetricAsPlainText(t *testing.T) {
	storage := memory.NewMetricRepository()

	for _, v := range []struct {
		n string
		v string
		t string
	}{
		{"gauge1", "1.05", metric.TypeGauge},
		{"gauge2", "1.15", metric.TypeGauge},
		{"counter1", "2", metric.TypeCounter},
		{"counter2", "3", metric.TypeCounter},
	} {
		m, err := metric.ParseMetric(v.n, v.v, v.t)
		if err != nil {
			t.Fatal(err)
		}

		if err := storage.Save(context.TODO(), m); err != nil {
			t.Fatal(err)
		}
	}
	svc := server.NewMetricService(storage)

	testCases := []struct {
		name        string
		method      string
		path        map[string]string
		contentType []string
		statusCode  int
		want        string
	}{
		{
			name:   "get gauge1, expect 1.05",
			method: http.MethodGet,
			path: map[string]string{
				"type": "gauge", "name": "gauge1",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
			want:       "1.05",
		},
		{
			name:   "get gauge2, expect 1.15",
			method: http.MethodGet,
			path: map[string]string{
				"type": "gauge", "name": "gauge2",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
			want:       "1.15",
		},
		{
			name:   "get counter1, expect 2",
			method: http.MethodGet,
			path: map[string]string{
				"type": "counter", "name": "counter1",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
			want:       "2",
		},
		{
			name:   "get counter2, expect 3",
			method: http.MethodGet,
			path: map[string]string{
				"type": "counter", "name": "counter2",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
			want:       "3",
		},
		{
			name:   "get counter, empty name, expect NotFound",
			method: http.MethodGet,
			path: map[string]string{
				"type": "counter",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusNotFound,
			want:       "",
		},
		{
			name:   "get counter, empty type, expect NotFound",
			method: http.MethodGet,
			path:   map[string]string{},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusNotFound,
			want:       "",
		},
		{
			name:   "get not existed counter, expect NotFound",
			method: http.MethodGet,
			path: map[string]string{
				"type": "counter", "name": "counter3",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusNotFound,
			want:       "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := chi.NewRouteContext()

			for k, v := range tt.path {
				ctx.URLParams.Add(k, v)
			}
			req, err := http.NewRequest(tt.method, "/value/{type}/{name}", nil)
			if err != nil {
				t.Fatal(err)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

			rr := httptest.NewRecorder()
			handler := DisplayMetric(svc)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
			if tt.want != "" {
				assert.Equal(t, tt.want, rr.Body.String())
			}
		})
	}
}

func TestDisplayMetricAsJSON(t *testing.T) {
	storage := memory.NewMetricRepository()

	for _, v := range []struct {
		n string
		v string
		t string
	}{
		{"gauge1", "1.05", metric.TypeGauge},
		{"gauge2", "1.15", metric.TypeGauge},
		{"counter1", "2", metric.TypeCounter},
		{"counter2", "3", metric.TypeCounter},
	} {
		m, err := metric.ParseMetric(v.n, v.v, v.t)
		if err != nil {
			t.Fatal(err)
		}

		if err := storage.Save(context.TODO(), m); err != nil {
			t.Fatal(err)
		}
	}
	svc := server.NewMetricService(storage)

	testCases := []struct {
		name       string
		method     string
		body       io.Reader
		statusCode int
		want       metric.Metric
	}{
		{
			name:       "get gauge1, expect 1.05",
			method:     http.MethodPost,
			body:       strings.NewReader(`{"id": "gauge1","type": "gauge"}`),
			statusCode: http.StatusOK,
			want:       metric.NewGaugeMetric("gauge1", 1.05),
		},
		{
			name:       "get gauge2, expect 1.05",
			method:     http.MethodPost,
			body:       strings.NewReader(`{"id": "gauge2","type": "gauge"}`),
			statusCode: http.StatusOK,
			want:       metric.NewGaugeMetric("gauge2", 1.15),
		},
		{
			name:       "get counter1, expect 2",
			method:     http.MethodPost,
			body:       strings.NewReader(`{"id": "counter1","type": "counter"}`),
			statusCode: http.StatusOK,
			want:       metric.NewCounterMetric("counter1", 2),
		},
		{
			name:       "get counter2, expect 3",
			method:     http.MethodPost,
			body:       strings.NewReader(`{"id": "counter2","type": "counter"}`),
			statusCode: http.StatusOK,
			want:       metric.NewCounterMetric("counter2", 3),
		},
		{
			name:       "get counter, empty name, expect NotFound",
			method:     http.MethodPost,
			body:       strings.NewReader(`{"id": "","type": "counter"}`),
			statusCode: http.StatusInternalServerError,
			want:       nil,
		},
		{
			name:       "get counter, empty type, expect NotFound",
			method:     http.MethodPost,
			body:       strings.NewReader(`{"id": "counter1","type": ""}`),
			statusCode: http.StatusBadRequest,
			want:       nil,
		},
		{
			name:       "get not existed counter, expect NotFound",
			method:     http.MethodPost,
			body:       strings.NewReader(`{"id": "counter3","type": "counter"}`),
			statusCode: http.StatusNotFound,
			want:       nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := chi.NewRouteContext()

			req, err := http.NewRequest(tt.method, "/value/", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))
			req.Header.Add("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler := DisplayMetric(svc)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.statusCode, rr.Code)

			if tt.want == nil {
				return
			}

			object := struct {
				ID    string   `json:"id"`
				MType string   `json:"type"`
				Delta *int64   `json:"delta,omitempty"`
				Value *float64 `json:"value,omitempty"`
			}{}

			err = json.Unmarshal(rr.Body.Bytes(), &object)
			require.NoError(t, err)

			assert.Equal(t, tt.want.Name(), object.ID)
			assert.Equal(t, tt.want.Type(), object.MType)

			if tt.want.Type() == metric.TypeGauge {
				assert.Equal(t, tt.want.Float64(), *object.Value)
			} else {
				assert.Equal(t, tt.want.Int64(), *object.Delta)
			}

		})
	}
}
