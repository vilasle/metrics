package rest

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

func showAllMetrics(svc service.StorageService, r *http.Request) Response {
	//handler catch all unregistered endpoints and block they
	if r.RequestURI != "/" {
		return NewTextResponse(emptyBody(), ErrForbiddenResource)
	}

	metrics, err := svc.AllMetrics()
	if err != nil {
		return NewTextResponse(emptyBody(), err)
	}

	content, err := generateViewOfAllMetrics(metrics)
	return NewHTMLResponse(content, err)
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
func showSpecificMetric(svc service.StorageService, r *http.Request) Response {
	switch r.Header.Get("Content-Type") {
	case "application/json":
		return handleDisplayMetricAsTextJSON(svc, r)
	default:
		return handleDisplayMetricAsTextPlain(svc, r)
	}
}

func handleDisplayMetricAsTextPlain(svc service.StorageService, r *http.Request) Response {
	raw := getRawDataFromContext(r.Context())

	if notFilled(raw.Name, raw.Kind) {
		return NewTextResponse(emptyBody(), ErrEmptyRequiredFields)
	}

	metric, err := svc.Get(raw.Name, raw.Kind)
	if err != nil {
		return NewTextResponse(emptyBody(), err)
	}
	return NewTextResponse([]byte(metric.Value()), nil)
}

func handleDisplayMetricAsTextJSON(svc service.StorageService, r *http.Request) Response {
	defer r.Body.Close()
	if r.Body == http.NoBody {
		return NewTextResponse(emptyBody(), ErrEmptyRequestBody)
	}
	content, err := io.ReadAll(r.Body)

	if err != nil {
		return NewTextResponse(emptyBody(), ErrReadingRequestBody)
	}

	m, err := metric.FromJSON(content)
	if err != nil && !errors.Is(err, metric.ErrNotFilledValue) {
		return NewTextResponse(emptyBody(), err)
	}

	metric, err := svc.Get(m.Name, m.Kind)
	if err != nil {
		return NewTextResponse(emptyBody(), err)
	}

	metricContent, err := metric.ToJSON()
	return NewJSONResponse(metricContent, err)
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
