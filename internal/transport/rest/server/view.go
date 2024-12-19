package rest

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

func showAllMetrics(svc service.StorageService, r *http.Request) Response {
	//handler catch all unregistered handlers and block for they
	if r.RequestURI != "/" {
		return NewTextResponse(emptyBody(), ErrForbiddenResource)
	}

	metrics, err := svc.AllMetrics()
	if err != nil {
		return NewTextResponse(emptyBody(), err)
	}

	content, err := generateViewOfAllMetrics(metrics)
	return NewHtmlResponse(content, err)
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

func showSpecificMetric(svc service.StorageService, r *http.Request) Response {
	contentType := r.Header.Get("Content-Type")

	switch contentType {
	case "text/plain":
		return handleDisplayMetricAsTextPlain(svc, r)
	case "application/json":
		return handleUpdateAsTextJson(svc, r)
	default:
		return NewTextResponse(emptyBody(), ErrUnknownContentType)
	}
}

func handleDisplayMetricAsTextPlain(svc service.StorageService, r *http.Request) Response {
	//TODO implement it
	panic("not implemented")
	// raw := getRawDataFromContext(r.Context())

	// if notFilled(raw.Name, raw.Kind) {
	// 	http.NotFound(w, nil)
	// 	return
	// }
	// //TODO error handling. define response by error
	// metric, err := svc.Get(raw.Name, raw.Kind)
	// if err != nil && (errors.Is(err, service.ErrMetricIsNotExist) || errors.Is(err, service.ErrUnknownKind)) {
	// 	http.NotFound(w, nil)
	// 	return
	// } else if err != nil {
	// 	http.Error(w, "", http.StatusInternalServerError)
	// 	return
	// }

	// w.Write([]byte(metric.Value()))

	// w.Header().Add("Content-Type", "text/plain]; charset=utf-8")
	// w.WriteHeader(http.StatusOK)
}

func handleDisplayMetricAsTextJson(svc service.StorageService, r *http.Request) Response {
	//TODO implement it
	panic("not implemented")
	// //TODO now only check logic. need to pass tests
	// content, err := io.ReadAll(r.Body)
	// r.Body.Close()

	// if err != nil {
	// 	http.Error(w, "", http.StatusBadRequest)
	// 	return
	// }

	// input := struct {
	// 	Id      string  `json:"id"`
	// 	Type    string  `json:"type"`
	// 	Gauge   float64 `json:"value,omitempty"`
	// 	Counter int64   `json:"delta,omitempty"`
	// }{}

	// err = json.Unmarshal(content, &input)
	// if err != nil {
	// 	http.Error(w, "", http.StatusBadRequest)
	// 	return
	// }

	// metric, err := svc.Get(input.Id, input.Type)
	// if err != nil && (errors.Is(err, service.ErrMetricIsNotExist) || errors.Is(err, service.ErrUnknownKind)) {
	// 	http.NotFound(w, nil)
	// 	return
	// } else if err != nil {
	// 	http.Error(w, "", http.StatusInternalServerError)
	// 	return
	// }

	// result, err := metric.ToJson()
	// if err != nil {
	// 	http.Error(w, "", http.StatusInternalServerError)
	// 	return
	// }

	// content, err = json.Marshal(result)
	// if err != nil {
	// 	http.Error(w, "", http.StatusInternalServerError)
	// 	return
	// }

	// w.Write(content)
	// w.Header().Add("Content-Type", "application/json")
	// w.WriteHeader(200)

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
