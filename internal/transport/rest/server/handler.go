package rest

import (
	"context"
	"errors"
	"fmt"
	"html/template"
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

func UpdateMetric(svc service.StorageService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := cleanUselessData(r.Context())

		err := svc.Save(
			metric.NewRawMetric(getName(d), getKind(d), getValue(d)),
		)

		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(getStatusCode(err))
	})
}

func DisplayAllMetrics(svc service.StorageService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//FIXME refactor it
		//this handler catch url like /updater or /values and need make up how it control
		if r.RequestURI != "/" {
			http.NotFound(w, nil)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		metrics, err := svc.AllMetrics()
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		data := viewData{Metrics: []struct {
			Name string
			Link string
		}{}}
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
		d := cleanUselessData(r.Context())

		k, n := getKind(d), getName(d)

		if k == "" || n == "" {
			http.NotFound(w, nil)
			return
		}
		//TODO error handling. define response by error
		metric, err := svc.Get(n, k)
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
	})
}

func cleanUselessData(ctx context.Context) []string {
	r := make([]string, 0)

	v := ctx.Value("type")
	_ = v

	r = appendIfIsFilled(r, chi.URLParamFromCtx(ctx, "type"))
	r = appendIfIsFilled(r, chi.URLParamFromCtx(ctx, "name"))
	r = appendIfIsFilled(r, chi.URLParamFromCtx(ctx, "value"))

	return r
}

func appendIfIsFilled(r []string, v string) []string {
	if v != "" {
		r = append(r, v)
	}
	return r
}

func getKind(data []string) string {
	if len(data) > 0 {
		return data[0]
	}
	return ""
}

func getName(data []string) string {
	if len(data) > 1 {
		return data[1]
	}
	return ""
}

func getValue(data []string) string {
	if len(data) > 2 {
		return data[2]
	}
	return ""
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
