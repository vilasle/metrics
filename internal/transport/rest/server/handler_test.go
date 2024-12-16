package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/vilasle/metrics/internal/repository/memory"
	"github.com/vilasle/metrics/internal/service/server"
)

// FIXME if use chi router with url params when test will not work, because chi keep params on request context
func TestUpdateMetricAsTextPlain(t *testing.T) {
	svc := server.NewStorageService(
		memory.NewMetricGaugeMemoryRepository(),
		memory.NewMetricCounterMemoryRepository(),
	)

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
