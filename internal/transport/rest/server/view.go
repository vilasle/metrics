package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

func showAllMetrics(svc service.MetricService, r *http.Request) Response {
	// handler catch all unregistered endpoints and block they
	if r.RequestURI != "/" {
		return newTextResponse(emptyBody(), ErrForbiddenResource)
	}

	metrics, err := svc.All(r.Context())
	if err != nil {
		return newTextResponse(emptyBody(), err)
	}

	content, err := generateViewOfAllMetrics(metrics)
	return newHTMLResponse(content, err)
}

func generateViewOfAllMetrics(metrics []metric.Metric) ([]byte, error) {
	data := struct {
		Metrics []struct {
			Name string
			Link string
		}
	}{}

	for _, m := range metrics {
		data.Metrics = append(data.Metrics, struct {
			Name string
			Link string
		}{
			Name: m.Name(),
			Link: fmt.Sprintf("/value/%s/%s", m.Type(), m.Name()),
		})
	}

	buf := &bytes.Buffer{}
	view, err := template.New("metrics").Parse(allMetricsTemplate())
	if err != nil {
		return emptyBody(), err
	}

	if err := view.Execute(buf, data); err == nil {
		return buf.Bytes(), nil
	} else {
		return emptyBody(), err
	}
}

/*
auto-tests use filled Content-Type header only for iter1
that's why handle any Content-Type as text/plain with exception of application/json
*/
func showSpecificMetric(svc service.MetricService, r *http.Request) Response {
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":
		return handleDisplayMetricAsTextJSON(svc, r)
	default:
		return handleDisplayMetricAsTextPlain(svc, r)
	}
}

func handleDisplayMetricAsTextPlain(svc service.MetricService, r *http.Request) Response {
	raw := getRawDataFromContext(r.Context())
	logger.Debugw("raw data from url", "url", r.URL.String(), "raw", raw)
	if notFilled(raw.Name, raw.Type) {
		return newTextResponse(emptyBody(), ErrEmptyRequiredFields)
	}

	metric, err := svc.Get(r.Context(), raw.Type, raw.Name)
	if err != nil {
		return newTextResponse(emptyBody(), err)
	}
	return newTextResponse([]byte(metric.Value()), nil)
}

func handleDisplayMetricAsTextJSON(svc service.MetricService, r *http.Request) Response {
	content, err := io.ReadAll(r.Body)

	r.Body.Close()

	if err != nil || len(content) == 0 {
		return newTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	logger.Debugw("request body", "url", r.URL.String(), "body", string(content))

	m, err := metric.FromJSON(content)
	if err != nil && !errors.Is(err, metric.ErrEmptyValue) {
		return newTextResponse(emptyBody(), err)
	}

	metric, err := svc.Get(r.Context(), m.Type(), m.Name())
	if err != nil {
		return newTextResponse(emptyBody(), err)
	}
	logger.Debugw("display metric", "metric", metric)

	metricContent, err := json.Marshal(metric)
	return newJSONResponse(metricContent, err)
}

func allMetricsTemplate() string {
	return `
	<html>
		<head>
			<title>Metrics</title>
		</head>
		<body>
			<ul style="list-style: none;">
			{{range .Metrics}}
				<li> <a href="{{ .Link }}">{{.Name}}</li>
			{{end}}
			</ul>
		</body>
	</html>`
}
