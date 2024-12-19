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

type simpleResponse struct {
	content []byte
	err     error
}

func (r simpleResponse) Write(w http.ResponseWriter) {
	w.WriteHeader(getStatusCode(r.err))

	w.Write(r.content)
}

type JSONResponse struct {
	sp simpleResponse
}

func NewJSONResponse(content []byte, err error) Response {
	return JSONResponse{sp: simpleResponse{content: content, err: err}}
}

func (r JSONResponse) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	r.sp.Write(w)
}

type TextResponse struct {
	sp simpleResponse
}

func NewTextResponse(content []byte, err error) Response {
	return TextResponse{sp: simpleResponse{content: content, err: err}}
}

func (r TextResponse) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/plain")
	r.sp.Write(w)
}

type HTMLResponse struct {
	sp simpleResponse
}

func NewHTMLResponse(content []byte, err error) Response {
	return HTMLResponse{sp: simpleResponse{content: content, err: err}}
}

func (r HTMLResponse) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html")
	r.sp.Write(w)
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
	return errIs(err, 
		service.ErrEmptyKind,
		service.ErrUnknownKind,
		service.ErrInvalidValue,
		service.ErrEmptyValue,
		ErrEmptyRequestBody,
		metric.ErrInvalidMetric,
		metric.ErrNotFilledValue,
	)
}
func errorNotFound(err error) bool {
	return errIs(err, 
		service.ErrEmptyName,
		service.ErrMetricIsNotExist, 
		service.ErrUnknownKind,
		ErrForbiddenResource,
		ErrEmptyRequiredFields,
	)
}

func errorUnsupportedContent(err error) bool {
	return errIs(err, ErrUnknownContentType)
}

func errIs(err error, errs ...error) bool {
	for _, e := range errs {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}
