package rest

import (
	"errors"
	"net/http"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

//Response is the interface for wrapping http.ResponseWriter and post-processing before write response
type Response interface {
	write(w http.ResponseWriter)
}

type simpleResponse struct {
	content []byte
	err     error
}

func (r simpleResponse) write(w http.ResponseWriter) {
	w.WriteHeader(getStatusCode(r.err))

	w.Write(r.content)
}

type jsonResponse struct {
	sp simpleResponse
}

func newJSONResponse(content []byte, err error) Response {
	return jsonResponse{sp: simpleResponse{content: content, err: err}}
}

func (r jsonResponse) write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	r.sp.write(w)
}

type textResponse struct {
	sp simpleResponse
}

func newTextResponse(content []byte, err error) Response {
	return textResponse{sp: simpleResponse{content: content, err: err}}
}

func (r textResponse) write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/plain")
	r.sp.write(w)
}

type htmlResponse struct {
	sp simpleResponse
}

func newHTMLResponse(content []byte, err error) Response {
	return htmlResponse{sp: simpleResponse{content: content, err: err}}
}

func (r htmlResponse) write(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html")
	r.sp.write(w)
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
