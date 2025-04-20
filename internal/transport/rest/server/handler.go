package rest

import (
	"net/http"

	"github.com/vilasle/metrics/internal/service"
)

type HandlerWithResponse func(w http.ResponseWriter, r *http.Request) Response

//TODO add godoc
func (fn HandlerWithResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := fn(w, r)
	response.Write(w)
}

//TODO add godoc
func UpdateMetric(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return updateMetric(svc, r)
	}
}

//TODO add godoc
func BatchUpdate(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return updateMetrics(svc, r)
	}
}

//TODO add godoc
func DisplayAllMetrics(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return showAllMetrics(svc, r)
	}
}

//TODO add godoc
func DisplayMetric(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return showSpecificMetric(svc, r)
	}
}

//TODO add godoc
func Ping(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return NewTextResponse(emptyBody(), svc.Ping(r.Context()))
	}
}
