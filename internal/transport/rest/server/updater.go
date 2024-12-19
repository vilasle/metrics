package rest

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

type rawData struct {
	Name  string
	Kind  string
	Value string
}

func updateMetric(svc service.StorageService, r *http.Request) Response {
	contentType := r.Header.Get("Content-Type")

	switch contentType {
	case "text/plain":
		return handleUpdateAsTextPlain(svc, r)
	case "application/json":
		return handleUpdateAsTextJson(svc, r)
	default:
		return NewTextResponse(emptyBody(), ErrUnknownContentType)
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

func getRawDataFromContext(ctx context.Context) rawData {
	return rawData{
		Kind:  chi.URLParamFromCtx(ctx, "type"),
		Name:  chi.URLParamFromCtx(ctx, "name"),
		Value: chi.URLParamFromCtx(ctx, "value"),
	}
}
