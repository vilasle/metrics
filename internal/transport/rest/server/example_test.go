package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository/memory"
	"github.com/vilasle/metrics/internal/service/server"
)

func ExampleUpdateMetric() {
	// test service
	svc := server.NewMetricService(memory.NewMetricRepository())
	handler := UpdateMetric(svc)

	// Update gauge metric as text/plain
	{
		// metric data
		id, mtype, value := "test", "gauge", "143.213"

		// url have to be like /update/{type}/{name}/{value}
		reqAddr, err := url.JoinPath("/update", mtype, id, value)
		if err != nil {
			fmt.Printf("can not join url: %v\n", err)
			return
		}
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("type", mtype)
		ctx.URLParams.Add("name", id)
		ctx.URLParams.Add("value", value)

		req := httptest.NewRequest(http.MethodPost, reqAddr, nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

		req.Header.Add("Content-Type", "text/plain")

		wrt := httptest.NewRecorder()

		handler.ServeHTTP(wrt, req)
		if wrt.Code != http.StatusOK {
			fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
		}
		fmt.Printf("gauge text/plain code: %d\n", wrt.Code)
	}

	// Update counter metric as text/plain
	{
		// metric data
		id, mtype, value := "test", "counter", "213"

		// url have to be like /update/{type}/{name}/{value}
		reqAddr, err := url.JoinPath("/update", mtype, id, value)
		if err != nil {
			fmt.Printf("can not join url: %v\n", err)
			return
		}
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("type", mtype)
		ctx.URLParams.Add("name", id)
		ctx.URLParams.Add("value", value)

		req := httptest.NewRequest(http.MethodPost, reqAddr, nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

		req.Header.Add("Content-Type", "text/plain")

		wrt := httptest.NewRecorder()

		handler.ServeHTTP(wrt, req)
		if wrt.Code != http.StatusOK {
			fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
		}
		fmt.Printf("counter text/plain code: %d\n", wrt.Code)
	}

	// Update gauge metric as application/json
	{
		// metric data
		rd := strings.NewReader(`{"id":"test","type":"gauge","value":143.213}`)
		reqAddr := "/update"

		req := httptest.NewRequest(http.MethodPost, reqAddr, rd)

		req.Header.Add("Content-Type", "application/json")

		wrt := httptest.NewRecorder()

		handler.ServeHTTP(wrt, req)
		if wrt.Code != http.StatusOK {
			fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
		}

		responseBody, err := io.ReadAll(wrt.Body)
		if err != nil {
			fmt.Printf("can not read response body by reason %v\n", err)
			return
		}

		fmt.Printf("gauge application/json code: %d\n", wrt.Code)
		fmt.Printf("gauge application/json body: %s\n", string(responseBody))
	}

	// Update counter metric as application/json
	{
		// metric data
		rd := strings.NewReader(`{"id":"test","type":"counter","delta":213}`)

		reqAddr := "/update"

		req := httptest.NewRequest(http.MethodPost, reqAddr, rd)

		req.Header.Add("Content-Type", "application/json")

		wrt := httptest.NewRecorder()

		handler.ServeHTTP(wrt, req)
		if wrt.Code != http.StatusOK {
			fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
		}

		responseBody, err := io.ReadAll(wrt.Body)
		if err != nil {
			fmt.Printf("can not read response body by reason %v\n", err)
			return
		}

		fmt.Printf("counter application/json code: %d\n", wrt.Code)
		fmt.Printf("counter application/json body: %s\n", string(responseBody))
	}

	// Output:
	// gauge text/plain code: 200
	// counter text/plain code: 200
	// gauge application/json code: 200
	// gauge application/json body: {"id":"test","type":"gauge","value":143.213}
	// counter application/json code: 200
	// counter application/json body: {"id":"test","type":"counter","delta":213}
}

func ExampleBatchUpdate() {
	// test service
	svc := server.NewMetricService(memory.NewMetricRepository())
	handler := BatchUpdate(svc)

	//  body
	rd := strings.NewReader(`[{"id":"test","type":"counter","delta":213},{"id":"test","type":"gauge","value":143.213}]`)
	reqAddr := "/updates/"

	req := httptest.NewRequest(http.MethodPost, reqAddr, rd)
	req.Header.Add("Content-Type", "application/json")

	wrt := httptest.NewRecorder()

	handler.ServeHTTP(wrt, req)
	if wrt.Code != http.StatusOK {
		fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
	}
	fmt.Printf("batch update code: %d\n", wrt.Code)

	// Output:
	// batch update code: 200
}

func ExampleDisplayAllMetrics() {
	storage := memory.NewMetricRepository()
	storage.Save(context.Background(), metric.NewCounterMetric("counter1", 123))
	storage.Save(context.Background(), metric.NewCounterMetric("counter2", 231))
	storage.Save(context.Background(), metric.NewCounterMetric("counter3", 321))
	storage.Save(context.Background(), metric.NewGaugeMetric("gauge1", 123.000123))
	storage.Save(context.Background(), metric.NewGaugeMetric("gauge2", 231.000123))
	storage.Save(context.Background(), metric.NewGaugeMetric("gauge3", 321.000123))
	svc := server.NewMetricService(storage)

	handler := DisplayAllMetrics(svc)

	reqAddr := "/"

	req := httptest.NewRequest(http.MethodGet, reqAddr, nil)
	req.Header.Add("Content-Type", "text/plain")

	wrt := httptest.NewRecorder()

	handler.ServeHTTP(wrt, req)
	if wrt.Code != http.StatusOK {
		fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
	}
	fmt.Printf("display all metrics code: %d\n", wrt.Code)

	// Output:
	// display all metrics code: 200
}

func ExampleDisplayMetric() {
	storage := memory.NewMetricRepository()
	storage.Save(context.Background(), metric.NewCounterMetric("counter1", 123))
	storage.Save(context.Background(), metric.NewCounterMetric("counter2", 231))
	storage.Save(context.Background(), metric.NewCounterMetric("counter3", 321))
	storage.Save(context.Background(), metric.NewGaugeMetric("gauge1", 123.000123))
	storage.Save(context.Background(), metric.NewGaugeMetric("gauge2", 231.000123))
	storage.Save(context.Background(), metric.NewGaugeMetric("gauge3", 321.000123))
	svc := server.NewMetricService(storage)

	handler := DisplayMetric(svc)

	// display gauge metric text/plain
	{
		// metric data
		id, mtype := "gauge1", "gauge"

		// url have to be like /update/{type}/{name}
		reqAddr, err := url.JoinPath("/value", mtype, id)
		if err != nil {
			fmt.Printf("can not join url: %v\n", err)
			return
		}
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("type", mtype)
		ctx.URLParams.Add("name", id)

		req := httptest.NewRequest(http.MethodGet, reqAddr, nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

		wrt := httptest.NewRecorder()

		handler.ServeHTTP(wrt, req)
		if wrt.Code != http.StatusOK {
			fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
		}
		body, err := io.ReadAll(wrt.Body)
		if err != nil {
			fmt.Printf("can not read body: %v\n", err)
			return
		}

		fmt.Printf("display gauge text/plain code: %d\n", wrt.Code)
		fmt.Printf("display gauge text/plain body: %s\n", string(body))
	}

	// display counter metric text/plain
	{
		// metric data
		id, mtype := "counter1", "counter"

		// url have to be like /update/{type}/{name}
		reqAddr, err := url.JoinPath("/value", mtype, id)
		if err != nil {
			fmt.Printf("can not join url: %v\n", err)
			return
		}
		ctx := chi.NewRouteContext()
		ctx.URLParams.Add("type", mtype)
		ctx.URLParams.Add("name", id)

		req := httptest.NewRequest(http.MethodGet, reqAddr, nil)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

		wrt := httptest.NewRecorder()

		handler.ServeHTTP(wrt, req)
		if wrt.Code != http.StatusOK {
			fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
		}
		body, err := io.ReadAll(wrt.Body)
		if err != nil {
			fmt.Printf("can not read body: %v\n", err)
			return
		}

		fmt.Printf("display counter text/plain code: %d\n", wrt.Code)
		fmt.Printf("display counter text/plain body: %s\n", string(body))
	}

	// display gauge metric application/json
	{
		
		// url have to be like /update/{type}/{name}
		reqAddr := "/value"

		// metric data
		rd := strings.NewReader(`{"id":"gauge1","type":"gauge"}`)

		req := httptest.NewRequest(http.MethodPost, reqAddr, rd)
		req.Header.Add("Content-Type", "application/json")

		wrt := httptest.NewRecorder()

		handler.ServeHTTP(wrt, req)
		if wrt.Code != http.StatusOK {
			fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
		}
		body, err := io.ReadAll(wrt.Body)
		if err != nil {
			fmt.Printf("can not read body: %v\n", err)
			return
		}

		fmt.Printf("display gauge application/json code: %d\n", wrt.Code)
		fmt.Printf("display gauge application/json body: %s\n", string(body))
	}
	// display counter metric application/json
	{
		// url have to be like /update/{type}/{name}
		reqAddr := "/value"

		// metric data
		rd := strings.NewReader(`{"id":"counter1","type":"counter"}`)

		req := httptest.NewRequest(http.MethodPost, reqAddr, rd)
		req.Header.Add("Content-Type", "application/json")

		wrt := httptest.NewRecorder()

		handler.ServeHTTP(wrt, req)
		if wrt.Code != http.StatusOK {
			fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
		}
		body, err := io.ReadAll(wrt.Body)
		if err != nil {
			fmt.Printf("can not read body: %v\n", err)
			return
		}

		fmt.Printf("display counter application/json code: %d\n", wrt.Code)
		fmt.Printf("display counter application/json body: %s\n", string(body))
	}
	// Output:
	// display gauge text/plain code: 200
	// display gauge text/plain body: 123.000123
	// display counter text/plain code: 200
	// display counter text/plain body: 123
	// display gauge application/json code: 200
	// display gauge application/json body: {"id":"gauge1","type":"gauge","value":123.000123}
	// display counter application/json code: 200
	// display counter application/json body: {"id":"counter1","type":"counter","delta":123}
}

func ExamplePing() {
	storage := memory.NewMetricRepository()
	svc := server.NewMetricService(storage)

	handler := Ping(svc)

	reqAddr := "/ping"

	req := httptest.NewRequest(http.MethodGet, reqAddr, nil)
	req.Header.Add("Content-Type", "text/plain")

	wrt := httptest.NewRecorder()

	handler.ServeHTTP(wrt, req)
	if wrt.Code != http.StatusOK {
		fmt.Printf("expected status code %d, got %d\n", http.StatusOK, wrt.Code)
	}
	fmt.Printf("ping code: %d\n", wrt.Code)
}
