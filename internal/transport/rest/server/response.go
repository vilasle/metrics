package rest

import (
	"errors"
	"net/http"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

type Response interface {
	Write(w http.ResponseWriter)
}

type JsonResponse struct {
	content []byte
	err     error
}

func NewJsonResponse(content []byte, err error) Response {
	return JsonResponse{content: content, err: err}
}

func (r JsonResponse) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	w.Write(r.content)
	w.WriteHeader(getStatusCode(r.err))
}

type TextResponse struct {
	content []byte
	err     error
}

func NewTextResponse(content []byte, err error) Response {
	return TextResponse{content: content, err: err}
}

func (r TextResponse) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/plain")

	w.WriteHeader(getStatusCode(r.err))
	w.Write(r.content)
}

type HtmlResponse struct {
	content []byte
	err     error
}

func NewHtmlResponse(content []byte, err error) Response {
	return HtmlResponse{content: content, err: err}
}

func (r HtmlResponse) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html")

	w.WriteHeader(getStatusCode(r.err))
	w.Write(r.content)
}

func getStatusCode(err error) int {
	if errorBadRequest(err) {
		return http.StatusBadRequest
	} else if errorNotFound(err) {
		return http.StatusNotFound
	} else if errorUnsupportedContent(err) {
		return http.StatusUnsupportedMediaType
	} else if err != nil {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func errorBadRequest(err error) bool {
	return errors.Is(err, service.ErrEmptyKind) ||
		errors.Is(err, service.ErrUnknownKind) ||
		errors.Is(err, service.ErrInvalidValue) ||
		errors.Is(err, service.ErrEmptyValue) ||
		errors.Is(err, ErrEmptyRequestBody) ||
		errors.Is(err, metric.ErrInvalidMetricType) ||
		errors.Is(err, metric.ErrInvalidMetric)
}

func errorNotFound(err error) bool {
	return errors.Is(err, service.ErrEmptyName) ||
		errors.Is(err, ErrForbiddenResource)
}

func errorUnsupportedContent(err error) bool {
	return errors.Is(err, ErrUnknownContentType)
}
