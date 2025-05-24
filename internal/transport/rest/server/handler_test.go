package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
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
	setup := func(mms *MockMetricService, ctx context.Context, result []metric.Metric, err error) {
		mms.EXPECT().All(ctx).Return(result, err)
	}

	type mockArgs struct {
		result []metric.Metric
		err    error
	}

	testCases := []struct {
		name       string
		path       string
		statusCode int
		contents   bool
		exp        string
		*mockArgs
	}{
		{
			name:       "get several metrics",
			statusCode: http.StatusOK,
			path:       "/",
			contents:   true,
			exp:        `<li>.+<\/li>`,
			mockArgs: &mockArgs{
				result: []metric.Metric{
					metric.NewGaugeMetric("test", 1.0),
					metric.NewCounterMetric("test2", 2.0),
				},
				err: nil,
			},
		},
		{
			name:       "empty storage",
			statusCode: http.StatusOK,
			path:       "/",
			contents:   false,
			exp:        `<li>.+<\/li>`,
			mockArgs: &mockArgs{
				result: []metric.Metric{},
				err:    nil,
			},
		},
		{
			name:       "wrong path",
			statusCode: http.StatusNotFound,
			path:       "/show/metrics",
			contents:   false,
			exp:        `<li>.+<\/li>`,
		},
		{
			name:       "storage error",
			statusCode: http.StatusInternalServerError,
			path:       "/",
			contents:   false,
			exp:        `<li>.+<\/li>`,
			mockArgs: &mockArgs{
				result: []metric.Metric{},
				err:    fmt.Errorf("error storage"),
			},
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

			//setup mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMockMetricService(ctrl)
			if tt.mockArgs != nil {
				setup(svc, req.Context(), tt.mockArgs.result, tt.mockArgs.err)
			}

			rr := httptest.NewRecorder()
			handler := DisplayAllMetrics(svc)
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
		{
			name:       "empty body",
			method:     http.MethodPost,
			body:       http.NoBody,
			statusCode: http.StatusInternalServerError,
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

func BenchmarkUpdateMetricAsPlainText(b *testing.B) {
	ctxGauge := chi.NewRouteContext()
	ctxCounter := chi.NewRouteContext()

	for k, v := range map[string]string{"type": "gauge", "name": "gauge1", "value": "12312.3123123"} {
		ctxGauge.URLParams.Add(k, v)
	}

	for k, v := range map[string]string{"type": "counter", "name": "counter1", "value": "3434312"} {
		ctxCounter.URLParams.Add(k, v)
	}

	storage := memory.NewMetricRepository()

	svc := server.NewMetricService(storage)

	reqG, err := http.NewRequest(http.MethodPost, "/update/{type}/{name}/{value}", nil)
	if err != nil {
		b.Fatal(err)
	}
	reqG = reqG.WithContext(context.WithValue(reqG.Context(), chi.RouteCtxKey, ctxGauge))
	reqG.Header.Set("Content-Type", "text/plain")

	reqC, err := http.NewRequest(http.MethodPost, "/update/{type}/{name}/{value}", nil)
	if err != nil {
		b.Fatal(err)
	}
	reqC = reqC.WithContext(context.WithValue(reqC.Context(), chi.RouteCtxKey, ctxCounter))
	reqC.Header.Set("Content-Type", "text/plain")

	b.ResetTimer()

	b.Run("update gauge", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rr := httptest.NewRecorder()
			UpdateMetric(svc).ServeHTTP(rr, reqG)
			if rr.Code != http.StatusOK {
				b.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Code)
			}
		}
	})

	b.Run("update counter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rr := httptest.NewRecorder()
			UpdateMetric(svc).ServeHTTP(rr, reqG)
			if rr.Code != http.StatusOK {
				b.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Code)
			}
		}
	})
}

func BenchmarkUpdateMetricAsJSON(b *testing.B) {
	storage := memory.NewMetricRepository()

	svc := server.NewMetricService(storage)

	gaugeRd := bytes.NewReader([]byte(`{"type": "gauge", "id": "gauge1", "value": 12312.3123123}`))
	counterRd := bytes.NewReader([]byte(`{"type": "counter", "id": "counter1", "delta": 12312}`))

	reqG, err := http.NewRequest(http.MethodPost, "/update/", gaugeRd)
	if err != nil {
		b.Fatal(err)
	}
	reqG.Header.Set("Content-Type", "application/json")

	reqC, err := http.NewRequest(http.MethodPost, "/update/", counterRd)
	if err != nil {
		b.Fatal(err)
	}
	reqC.Header.Set("Content-Type", "application/json")

	b.ResetTimer()

	b.Run("update gauge", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rr := httptest.NewRecorder()
			h := UpdateMetric(svc)
			h.ServeHTTP(rr, reqG)
		}
	})

	b.Run("update counter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			rr := httptest.NewRecorder()
			h := UpdateMetric(svc)
			h.ServeHTTP(rr, reqC)
		}
	})
}

func BenchmarkUpdateMetricAsBatch(b *testing.B) {
	storage := memory.NewMetricRepository()

	batch := make([]metric.Metric, 0, 2000)

	qtyG := 1000
	qtyC := 1000

	for i := 0; i < qtyG; i++ {
		batch = append(batch, metric.NewGaugeMetric(fmt.Sprintf("gauge%d", i), rand.Float64()))
	}

	for i := 0; i < qtyC; i++ {
		batch = append(batch, metric.NewCounterMetric(fmt.Sprintf("counter%d", i), rand.Int64()))
	}

	content, err := json.Marshal(batch)
	if err != nil {
		b.Fatal(err)
	}

	rd := bytes.NewReader(content)

	svc := server.NewMetricService(storage)

	req, err := http.NewRequest(http.MethodPost, "/updates/", rd)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		h := BatchUpdate(svc)
		h.ServeHTTP(rr, req)
	}
}

func BenchmarkDisplayAllMetricsAsHtml(b *testing.B) {
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		b.Fatal(err)
	}
	ctx := chi.NewRouteContext()

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	storage := memory.NewMetricRepository()
	svc := server.NewMetricService(storage)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler := DisplayAllMetrics(svc)
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkDisplayMetricAsTextPlain(b *testing.B) {
	storage := memory.NewMetricRepository()
	svc := server.NewMetricService(storage)

	qtyG := 10
	qtyC := 10
	for i := 0; i < qtyG; i++ {
		m := metric.NewGaugeMetric(fmt.Sprintf("gauge%d", i), rand.Float64())

		if err := storage.Save(context.TODO(), m); err != nil {
			b.Fatal(err)
		}
	}

	for i := 0; i < qtyC; i++ {
		m := metric.NewCounterMetric(fmt.Sprintf("counter%d", i), rand.Int64())
		if err := storage.Save(context.TODO(), m); err != nil {
			b.Fatal(err)
		}
	}

	ctxG := chi.NewRouteContext()

	for k, v := range map[string]string{"type": "gauge", "name": "gauge1"} {
		ctxG.URLParams.Add(k, v)
	}
	reqG, err := http.NewRequest(http.MethodGet, "/value/{type}/{name}", nil)
	if err != nil {
		b.Fatal(err)
	}
	reqG = reqG.WithContext(context.WithValue(reqG.Context(), chi.RouteCtxKey, ctxG))

	ctxC := chi.NewRouteContext()

	for k, v := range map[string]string{"type": "gauge", "name": "gauge1"} {
		ctxC.URLParams.Add(k, v)
	}
	reqC, err := http.NewRequest(http.MethodGet, "/value/{type}/{name}", nil)
	if err != nil {
		b.Fatal(err)
	}
	reqC = reqC.WithContext(context.WithValue(reqC.Context(), chi.RouteCtxKey, ctxC))

	b.ResetTimer()

	b.Run("get gauge metric", func(b *testing.B) {
		rr := httptest.NewRecorder()
		handler := DisplayMetric(svc)
		handler.ServeHTTP(rr, reqG)
	})

	b.Run("get counter metric", func(b *testing.B) {
		rr := httptest.NewRecorder()
		handler := DisplayMetric(svc)
		handler.ServeHTTP(rr, reqC)
	})
}

func BenchmarkDisplayMetricAsJSON(b *testing.B) {
	storage := memory.NewMetricRepository()
	svc := server.NewMetricService(storage)

	qtyG := 10
	qtyC := 10
	for i := 0; i < qtyG; i++ {
		m := metric.NewGaugeMetric(fmt.Sprintf("gauge%d", i), rand.Float64())

		if err := storage.Save(context.TODO(), m); err != nil {
			b.Fatal(err)
		}
	}

	for i := 0; i < qtyC; i++ {
		m := metric.NewCounterMetric(fmt.Sprintf("counter%d", i), rand.Int64())
		if err := storage.Save(context.TODO(), m); err != nil {
			b.Fatal(err)
		}
	}

	reqG, err := http.NewRequest(http.MethodGet, "/value/", bytes.NewReader([]byte(`{"type": "gauge", "id": "gauge1"}`)))
	if err != nil {
		b.Fatal(err)
	}

	reqC, err := http.NewRequest(http.MethodGet, "/value/", bytes.NewReader([]byte(`{"type": "counter", "id": "counter1"}`)))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	b.Run("get gauge metric as json", func(b *testing.B) {
		rr := httptest.NewRecorder()
		handler := DisplayMetric(svc)
		handler.ServeHTTP(rr, reqG)
	})

	b.Run("get counter metric as json", func(b *testing.B) {
		rr := httptest.NewRecorder()
		handler := DisplayMetric(svc)
		handler.ServeHTTP(rr, reqC)
	})
}
