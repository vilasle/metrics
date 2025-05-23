package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

type rawData struct {
	Name  string
	Type  string
	Value string
}

/*
auto-tests use filled Content-Type header only for iter1
that's why handle any Content-Type as text/plain with exception of application/json
*/
func updateMetric(svc service.MetricService, r *http.Request) Response {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return handleUpdateAsTextJSON(svc, r)
	default:
		return handleUpdateAsTextPlain(svc, r)
	}
}

func handleUpdateAsTextPlain(svc service.MetricService, r *http.Request) Response {
	raw := getRawDataFromContext(r.Context())
	logger.Debugw("raw data from url", "raw", raw, "url", r.URL.String())

	if m, err := metric.ParseMetric(raw.Name, raw.Value, raw.Type); err == nil {
		err := svc.Save(r.Context(), m)
		return newTextResponse(emptyBody(), err)
	} else {
		return newTextResponse(emptyBody(), err)
	}
}

func handleUpdateAsTextJSON(svc service.MetricService, r *http.Request) Response {
	defer r.Body.Close()
	content, err := io.ReadAll(r.Body)
	if err != nil || len(content) == 0 {
		return newTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	logger.Debugw("request body", "url", r.URL.String(), "body", string(content))

	m, err := metric.FromJSON(content)
	if err != nil {
		return newTextResponse(emptyBody(), err)
	}

	if err = svc.Save(r.Context(), m); err != nil {
		return newTextResponse(emptyBody(), err)
	}

	logger.Debugw("updated metric", "metric", m)

	updContent, err := json.Marshal(m)
	return newJSONResponse(updContent, err)
}

func updateMetrics(svc service.MetricService, r *http.Request) Response {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return handleUpdateMetricsAsBatch(svc, r)
	default:
		return newTextResponse(emptyBody(), ErrUnknownContentType)
	}
}

func handleUpdateMetricsAsBatch(svc service.MetricService, r *http.Request) Response {
	defer r.Body.Close()
	content, err := io.ReadAll(r.Body)
	if err != nil || len(content) == 0 {
		return newTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	logger.Debugw("request body", "url", r.URL.String(), "body", string(content))

	ms, err := metric.FromJSONArray(content)
	if err != nil {
		return newTextResponse(emptyBody(), err)
	}

	if err = svc.Save(r.Context(), ms...); err != nil {
		return newTextResponse(emptyBody(), err)
	}

	logger.Debugw("updated metrics", "metric", ms)

	return newTextResponse(emptyBody(), err)
}
