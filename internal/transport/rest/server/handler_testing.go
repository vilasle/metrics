package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vilasle/metrics/internal/repository/memory"
	"github.com/vilasle/metrics/internal/service/server"
)

//FIXME if use chi router with url params when test will not work, because chi keep params on request context
func TestUpdateMetric(t *testing.T) {
	svc := server.NewStorageService(
		memory.NewMetricGaugeMemoryRepository(),
		memory.NewMetricCounterMemoryRepository(),
	)

	cases := []struct {
		name        string
		path        []string
		method      string
		contentType []string
		statusCode  int
	}{
		{
			name:   "send normal gauge",
			method: http.MethodPost,
			path: []string{
				"/update/gauge/test/1.05",
				"/update/gauge/test1/1.033",
				"/update/gauge/test/140.10",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send normal gauge with big value",
			method: http.MethodPost,
			path: []string{
				"/update/gauge/test/43435345435345.343424205",
				"/update/gauge/test1/4343534234634342.033",
				"/update/gauge/test/14000000000000.10",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send a gauge with negative value",
			method: http.MethodPost,
			path: []string{
				"/update/gauge/test/-43435345435345.343424205",
				"/update/gauge/test1/-4343534234634342.033",
				"/update/gauge/test/-14000000000000.10",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send gauge with string value",
			method: http.MethodPost,
			path: []string{
				"/update/gauge/test/string_value",
			},

			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send gauge with empty value",
			method: http.MethodPost,
			path: []string{
				"/update/gauge/test/",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send normal counter",
			method: http.MethodPost,
			path: []string{
				"/update/counter/test/124",
				"/update/counter/test/12452",
				"/update/counter/test/213124",
				"/update/counter/test/0",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send normal counter with big value",
			method: http.MethodPost,
			path: []string{
				"/update/counter/test/12400000000000",
				"/update/counter/test/1245200000000000000",
				"/update/counter/test/213124000000000000",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send a counter with negative value",
			method: http.MethodPost,
			path: []string{
				"/update/counter/test/-124",
				"/update/counter/test/-12452",
				"/update/counter/test/-213124",
				"/update/counter/test/-0",
			},

			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusOK,
		},
		{
			name:   "send a counter with string value",
			method: http.MethodPost,
			path: []string{
				"/update/counter/test/sting_value",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send counter with empty value",
			method: http.MethodPost,
			path: []string{
				"/update/counter/test/",
				"/update/counter/test",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send unsupported type of metric",
			method: http.MethodPost,
			path: []string{
				"/update/another_metrics/test/",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name:   "send metric without name",
			method: http.MethodPost,
			path: []string{
				"/update/counter/",
				"/update/counter",
				"/update/gauge/",
				"/update/gauge",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusNotFound,
		},
		{
			name:   "send without kind of metric",
			method: http.MethodPost,
			path: []string{
				"/update/",
			},
			contentType: []string{
				"text/plain",
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			for _, p := range tt.path {
				req, err := http.NewRequest(tt.method, p, nil)

				if err != nil {
					t.Fatal(err)
				}
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
