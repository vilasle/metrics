package rest

import (
	"net/http"

	"github.com/vilasle/metrics/internal/service"
)

type HandlerWithResponse func(w http.ResponseWriter, r *http.Request) Response

// ServeHTTP implements http.Handler interface.
func (fn HandlerWithResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := fn(w, r)
	response.write(w)
}

// UpdateMetric is handler for adding/updating metric.
// Accept POST requests.
// Can accept Content-Type [text/plain, application/json]
// if request Content-Type is text/plain, then url must be in format /<type>/<name>/<value>. Response body will by empty
// if request Content-Type is application/json
// for gauge metric body must be in format :
// {"type": "gauge", "id" : "metric_id", "value": "metric_value"}
// for counter metric body must be in format :
// {"type": "gauge", "id" : "metric_id", "delta": metric_value}
func UpdateMetric(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return updateMetric(svc, r)
	}
}

// UpdateMetric is handler for adding/updating metrics.
// Accept POST requests.
// Can accept Content-Type [application/json]
// Accept json array and pass data to service.MetricService
// Body body must be in format :
// [
//
//		{"type": "gauge", "id" : "metric_id", "value": "metric_value"},
//	 {"type": "counter", "id" : "metric_id", "delta": metric_value}
//
// ]
func BatchUpdate(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return updateMetrics(svc, r)
	}
}

// DisplayAllMetrics is handler for displaying all metrics.
// Accept GET requests.
// Return HTML page with list of all metrics.
func DisplayAllMetrics(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return showAllMetrics(svc, r)
	}
}

// DisplayMetric is handler for displaying specific metric.
// Accept POST requests for json body and GET requests for empty body
// Can accept Content-Type [text/plain, application/json]
// if request Content-Type is text/plain, then url must be in format /<type>/<name>/. Response body will content value of metric
// if request Content-Type is application/json body must be in format : {"type": "gauge", "id" : "metric_id"}
// Response body will content json string with value of metric like this:
//
//		{"type": "gauge", "id" : "metric_id", "value": "metric_value"} - for gauge metric,
//	 {"type": "counter", "id" : "metric_id", "delta": metric_value} - for counter metric
func DisplayMetric(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return showSpecificMetric(svc, r)
	}
}

// Ping is handler for checking service health.
// Accept GET requests.
// Return 200 OK if service is healthy.
func Ping(svc service.MetricService) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return newTextResponse(emptyBody(), svc.Ping(r.Context()))
	}
}
