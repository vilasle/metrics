package rest

import (
	"net/http"

	"github.com/vilasle/metrics/internal/service"
	"go.uber.org/zap"
)

type HandlerWithResponse func(w http.ResponseWriter, r *http.Request) Response

func (fn HandlerWithResponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	response := fn(w, r)
	response.Write(w)
}

func UpdateMetric(svc service.StorageService, logger *zap.SugaredLogger) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return updateMetric(svc, r, logger)
	}
}

func DisplayAllMetrics(svc service.StorageService, logger *zap.SugaredLogger) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return showAllMetrics(svc, r)
	}
}

func DisplayMetric(svc service.StorageService, logger *zap.SugaredLogger) HandlerWithResponse {
	return func(w http.ResponseWriter, r *http.Request) Response {
		return showSpecificMetric(svc, r, logger)
	}
}
