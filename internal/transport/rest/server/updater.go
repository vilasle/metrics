package rest

import (
	"io"
	"net/http"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
	"go.uber.org/zap"
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
func updateMetric(svc service.StorageService, r *http.Request, logger *zap.SugaredLogger) Response {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return handleUpdateAsTextJSON(svc, r, logger)
	default:
		return handleUpdateAsTextPlain(svc, r, logger)
	}
}

func handleUpdateAsTextPlain(svc service.StorageService, r *http.Request, logger *zap.SugaredLogger) Response {
	raw := getRawDataFromContext(r.Context())
	logger.Debugw("raw data from url", "raw", raw, "url", r.URL.String())
	err := svc.Save(
		metric.NewRawMetric(raw.Name, raw.Kind, raw.Value),
	)
	return NewTextResponse(emptyBody(), err)
}

func handleUpdateAsTextJSON(svc service.StorageService, r *http.Request, logger *zap.SugaredLogger) Response {
	defer r.Body.Close()
	if r.Body == http.NoBody {
		return NewTextResponse(emptyBody(), ErrEmptyRequestBody)
	}
	content, err := io.ReadAll(r.Body)

	if err != nil {
		return NewTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	logger.Debugw("request body", "url", r.URL.String(), "body", string(content))

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

	logger.Debugw("updated metric", "metric", updMetric)

	updContent, err := updMetric.ToJSON()
	return NewJSONResponse(updContent, err)
}
