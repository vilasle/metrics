package rest

import (
	"errors"
	"net/http"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

//TODO add godoc
type Response interface {
	Write(w http.ResponseWriter)
}

type simpleResponse struct {
	content []byte
	err     error
}

//TODO add godoc
func (r simpleResponse) Write(w http.ResponseWriter) {
	w.WriteHeader(getStatusCode(r.err))

	w.Write(r.content)
}

//TODO add godoc
type JSONResponse struct {
	sp simpleResponse
}

//TODO add godoc
func NewJSONResponse(content []byte, err error) Response {
	return JSONResponse{sp: simpleResponse{content: content, err: err}}
}

//TODO add godoc
func (r JSONResponse) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	r.sp.Write(w)
}

//TODO add godoc
type TextResponse struct {
	sp simpleResponse
}

//TODO add godoc
func NewTextResponse(content []byte, err error) Response {
	return TextResponse{sp: simpleResponse{content: content, err: err}}
}

//TODO add godoc
func (r TextResponse) Write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/plain")
	r.sp.Write(w)
}

//TODO add godoc
type HTMLResponse struct {
	sp simpleResponse
}

//TODO add godoc
func NewHTMLResponse(content []byte, err error) Response {
	return HTMLResponse{sp: simpleResponse{content: content, err: err}}
}

//TODO add godoc
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
		service.ErrEmptyValue,
		ErrEmptyRequestBody,
		metric.ErrConvertingRawValue,
		metric.ErrEmptyValue,
		metric.ErrUnknownMetricType,
		ErrInvalidHashSum,
	)
}
func errorNotFound(err error) bool {
	return errIs(err,
		service.ErrEmptyName,
		service.ErrMetricIsNotExist,
		service.ErrUnknownKind,
		ErrForbiddenResource,
		ErrEmptyRequiredFields,
		metric.ErrEmptyName,
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
