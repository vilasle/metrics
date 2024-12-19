package rest

import (
	"io"
	"net/http"

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
		return handleUpdateAsTextJson(svc, r)
	default:
		return handleUpdateAsTextPlain(svc, r)
	}
}

func handleUpdateAsTextPlain(svc service.StorageService, r *http.Request) Response {
	raw := getRawDataFromContext(r.Context())
	err := svc.Save(
		metric.NewRawMetric(raw.Name, raw.Kind, raw.Value),
	)
	return NewTextResponse(emptyBody(), err)
}

func handleUpdateAsTextJson(svc service.StorageService, r *http.Request) Response {
	defer r.Body.Close()
	if r.Body == http.NoBody {
		return NewTextResponse(emptyBody(), ErrEmptyRequestBody)
	}
	content, err := io.ReadAll(r.Body)

	if err != nil {
		return NewTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	m, err := metric.FromJSON(content)
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

	updContent, err := updMetric.ToJson()
	return NewJsonResponse(updContent, err)
}
