package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

type viewData struct {
	Metrics []struct {
		Name string
		Link string
	}
}

type RawData struct {
	Name  string
	Kind  string
	Value string
}

func UpdateMetric(svc service.StorageService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := r.Header.Get("Content-Type")

		switch content {
		case "text/plain":
			handleUpdateAsTextPlain(svc, w, r)
		case "application/json":
			handleUpdateAsTextJson(svc, w, r)
		default:
			http.Error(w, "", http.StatusBadRequest)
			return
		}
	})
}

func handleUpdateAsTextPlain(svc service.StorageService, w http.ResponseWriter, r *http.Request) {
	raw := getRawDataFromContext(r.Context())
	err := svc.Save(
		metric.NewRawMetric(raw.Name, raw.Kind, raw.Value),
	)

	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(getStatusCode(err))
}

func handleUpdateAsTextJson(svc service.StorageService, w http.ResponseWriter, r *http.Request) {
	//TODO now only check logic. need to pass tests
	content, err := io.ReadAll(r.Body)
	r.Body.Close()

	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	input := struct {
		Id      string  `json:"id"`
		Type    string  `json:"type"`
		Gauge   float64 `json:"value"`
		Counter int64   `json:"delta"`
	}{}

	err = json.Unmarshal(content, &input)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	raw := RawData{
		Name: input.Id,
		Kind: input.Type,
	}

	if raw.Kind == "gauge" {
		raw.Value = fmt.Sprintf("%f", input.Gauge)
	} else {
		raw.Value = fmt.Sprintf("%d", input.Counter)
	}

	err = svc.Save(
		metric.NewRawMetric(raw.Name, raw.Kind, raw.Value),
	)

	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}

	w.Write(content)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
}

func DisplayAllMetrics(svc service.StorageService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//handler catch all unregistered handlers and block for they
		if r.RequestURI != "/" {
			http.NotFound(w, nil)
			return
		}

		metrics, err := svc.AllMetrics()
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		data := viewData{
			Metrics: []struct {
				Name string
				Link string
			}{},
		}
		for _, m := range metrics {
			data.Metrics = append(data.Metrics, struct {
				Name string
				Link string
			}{
				Name: m.Name(),
				Link: fmt.Sprintf("/value/%s/%s", m.Type(), m.Name()),
			})
		}

		if view, err := template.New("metrics").Parse(allMetricsTemplate()); err == nil {
			view.Execute(w, data)
		} else {
			http.Error(w, "", http.StatusInternalServerError)
			return

		}
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	})
}

func DisplayMetric(svc service.StorageService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		content := r.Header.Get("Content-Type")

		switch content {
		case "text/plain":
			handleDisplayMetricAsTextPlain(svc, w, r)
		case "application/json":
			handleDisplayMetricAsTextJson(svc, w, r)
		default:
			http.Error(w, "", http.StatusBadRequest)
			return
		}
	})
}

func handleDisplayMetricAsTextPlain(svc service.StorageService, w http.ResponseWriter, r *http.Request) {
	raw := getRawDataFromContext(r.Context())

	if notFilled(raw.Name, raw.Kind) {
		http.NotFound(w, nil)
		return
	}
	//TODO error handling. define response by error
	metric, err := svc.Get(raw.Name, raw.Kind)
	if err != nil && (errors.Is(err, service.ErrMetricIsNotExist) || errors.Is(err, service.ErrUnknownKind)) {
		http.NotFound(w, nil)
		return
	} else if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(metric.Value()))

	w.Header().Add("Content-Type", "text/plain]; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func handleDisplayMetricAsTextJson(svc service.StorageService, w http.ResponseWriter, r *http.Request) {
	//TODO now only check logic. need to pass tests
	content, err := io.ReadAll(r.Body)
	r.Body.Close()

	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	input := struct {
		Id      string  `json:"id"`
		Type    string  `json:"type"`
		Gauge   float64 `json:"value,omitempty"`
		Counter int64   `json:"delta,omitempty"`
	}{}

	err = json.Unmarshal(content, &input)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	metric, err := svc.Get(input.Id, input.Type)
	if err != nil && (errors.Is(err, service.ErrMetricIsNotExist) || errors.Is(err, service.ErrUnknownKind)) {
		http.NotFound(w, nil)
		return
	} else if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	result, err := metric.ToJson()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	content, err = json.Marshal(result)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Write(content)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)

}

func getRawDataFromContext(ctx context.Context) RawData {
	return RawData{
		Kind:  chi.URLParamFromCtx(ctx, "type"),
		Name:  chi.URLParamFromCtx(ctx, "name"),
		Value: chi.URLParamFromCtx(ctx, "value"),
	}
}

func getStatusCode(err error) int {
	if errorBadRequest(err) {
		return http.StatusBadRequest
	} else if errorNotFound(err) {
		return http.StatusNotFound
	}
	return http.StatusOK
}

func errorBadRequest(err error) bool {
	return errors.Is(err, service.ErrEmptyKind) ||
		errors.Is(err, service.ErrUnknownKind) ||
		errors.Is(err, service.ErrInvalidValue) ||
		errors.Is(err, service.ErrEmptyValue)
}

func errorNotFound(err error) bool {
	return err == service.ErrEmptyName
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

func filled(v ...string) bool {
	for _, v := range v {
		if v == "" {
			return false
		}
	}
	return true
}

func notFilled(v ...string) bool {
	return !filled(v...)
}
