package rest

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

type rawData struct {
	Name  string
	Kind  string
	Value string
}

/*
auto-tests use filled Content-Type header only for iter1
that's why handle any Content-Type as text/plain with exception of application/json
*/
func updateMetric(svc service.StorageService, r *http.Request) Response {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return handleUpdateAsTextJSON(svc, r)
	default:
		return handleUpdateAsTextPlain(svc, r)
	}
}

func handleUpdateAsTextPlain(svc service.StorageService, r *http.Request) Response {
	raw := getRawDataFromContext(r.Context())
	logger.Debugw("raw data from url", "raw", raw, "url", r.URL.String())
	err := svc.Save(
		metric.NewRawMetric(raw.Name, raw.Kind, raw.Value),
	)
	return NewTextResponse(emptyBody(), err)
}

func handleUpdateAsTextJSON(svc service.StorageService, r *http.Request) Response {
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

	updMetric, err := svc.Get(m.Name, m.Kind)
	if err != nil {
		return NewTextResponse(emptyBody(), err)
	}

	logger.Debugw("updated metric", "metric", updMetric)

	updContent, err := updMetric.ToJSON()
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
