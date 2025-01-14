package rest

import (
	"net/http"

	"github.com/vilasle/metrics/internal/service"
)

type HandlerWithResponse func(w http.ResponseWriter, r *http.Request) Response

func (fn HandlerWithResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := fn(w, r)
	response.Write(w)
}

func UpdateMetric(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return updateMetric(svc, r)
	}
}

func DisplayAllMetrics(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return showAllMetrics(svc, r)
	}
}

func DisplayMetric(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return showSpecificMetric(svc, r)
	}
}
