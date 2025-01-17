package rest

import (
	"bytes"
	"compress/gzip"
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
		err := svc.Save(m)
		return NewTextResponse(emptyBody(), err)
	} else {
		return NewTextResponse(emptyBody(), err)
	}
}

func handleUpdateAsTextJSON(svc service.MetricService, r *http.Request) Response {
	defer r.Body.Close()
	if r.Body == http.NoBody {
		return NewTextResponse(emptyBody(), ErrEmptyRequestBody)
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return NewTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	decompressedContent, err := unpackContent(content, r.Header.Get("Content-Encoding") == "gzip")
	if err != nil {
		return NewTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	logger.Debugw("request body", "url", r.URL.String(), "body", string(decompressedContent))

	m, err := metric.FromJSON(decompressedContent)
	if err != nil {
		return NewTextResponse(emptyBody(), err)
	}

	if err = svc.Save(m); err != nil {
		return NewTextResponse(emptyBody(), err)
	}

	logger.Debugw("updated metric", "metric", m)

	updContent, err := json.Marshal(m)
	return NewJSONResponse(updContent, err)
}

func unpackContent(content []byte, isCompressed bool) ([]byte, error) {
	if !isCompressed {
		return content, nil
	}

	rd := bytes.NewReader(content)
	grd, err := gzip.NewReader(rd)
	if err != nil {
		return nil, err
	}

	defer grd.Close()
	if c, err := io.ReadAll(grd); err == io.EOF {
		return c, nil
	} else {
		return c, err
	}
}

func updateMetrics(svc service.MetricService, r *http.Request) Response {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return handleUpdateMetricsAsBatch(svc, r)
	default:
		return NewTextResponse(emptyBody(), ErrUnknownContentType)
	}
}

func handleUpdateMetricsAsBatch(svc service.MetricService, r *http.Request) Response {
	defer r.Body.Close()
	if r.Body == http.NoBody {
		return NewTextResponse(emptyBody(), ErrEmptyRequestBody)
	}
	content, err := io.ReadAll(r.Body)
	if err != nil {
		return NewTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	decompressedContent, err := unpackContent(content, r.Header.Get("Content-Encoding") == "gzip")
	if err != nil {
		return NewTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	logger.Debugw("request body", "url", r.URL.String(), "body", string(decompressedContent))

	ms, err := metric.FromJSONArray(decompressedContent)
	if err != nil {
		return NewTextResponse(emptyBody(), err)
	}

	if err = svc.Save(ms...); err != nil {
		return NewTextResponse(emptyBody(), err)
	}

	logger.Debugw("updated metrics", "metric", ms)

	return NewTextResponse(emptyBody(), err)
}
