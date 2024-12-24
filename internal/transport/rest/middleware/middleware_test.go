package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vilasle/metrics/internal/logger"
)

func Test_allowedMethods(t *testing.T) {
	type args struct {
		allowedMethods []string
		handler        http.Handler
		method         string
		path           string
	}
	tests := []struct {
		name       string
		args       args
		statusCode int
	}{
		{
			name: "method is allowed, expected 200 OK",
			args: args{
				allowedMethods: []string{http.MethodGet, http.MethodPost},
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method: http.MethodGet,
				path:   "/",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "method is not allowed, expected 405 Method Not Allowed",
			args: args{
				allowedMethods: []string{http.MethodPost, http.MethodPut},
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method: http.MethodGet,
				path:   "/",
			},
			statusCode: http.StatusMethodNotAllowed,
		},
		{
			name: "allowed any methods, expected 200 OK",
			args: args{
				allowedMethods: []string{},
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method: http.MethodGet,
				path:   "/",
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := AllowedMethods(tt.args.allowedMethods...)
			middleware := fn(tt.args.handler)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)

			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

func Test_allowedContentType(t *testing.T) {
	type args struct {
		contentTypes []string
		contentType  []string
		handler      http.Handler
		method       string
		path         string
	}
	tests := []struct {
		name       string
		args       args
		statusCode int
	}{
		{
			name: "contents allowed content-types, expected 200 OK",
			args: args{
				contentTypes: []string{"text/plain", "application/json"},
				contentType:  []string{"text/plain"},
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method: http.MethodGet,
				path:   "/",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "does not content allowed content-types, expected 415 Unsupported Media Type",
			args: args{
				contentTypes: []string{"text/plain", "application/json"},
				contentType:  []string{"application/xml"},
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method: http.MethodGet,
				path:   "/",
			},
			statusCode: http.StatusUnsupportedMediaType,
		},
		{
			name: "allowed any content-types, expected 200 OK",
			args: args{
				contentTypes: []string{},
				contentType:  []string{"application/xml"},
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method: http.MethodGet,
				path:   "/",
			},
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := AllowedContentType(tt.args.contentTypes...)
			middleware := fn(tt.args.handler)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			for _, v := range tt.args.contentType {
				req.Header.Add("Content-Type", v)
			}
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.statusCode, rr.Code)
		})
	}
}

type writerMock struct {
	content []byte
}

func (l *writerMock) Write(content []byte) (int, error) {
	l.content = append(l.content, content...)
	return len(content), nil
}

func (l writerMock) Sync() error {
	return nil
}

func TestWithLogger(t *testing.T) {

	wrt := &writerMock{
		content: make([]byte, 0),
	}

	logger.Init(wrt, true)

	type args struct {
		handler http.Handler
		method  string
		path    string
	}
	tests := []struct {
		name string
		args args
		exp  string
	}{
		{
			name: "Code: Ok",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				method: http.MethodGet,
				path:   "/",
			},
			exp: `"level":"info".+"ts":\d+.\d+,"msg":.+,"uri":.+,"method":.+,"code":200,"delay":\d+.\d+,"size":\d+`,
		},
		{
			name: "Code: Not Found",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}),
				method: http.MethodGet,
				path:   "/",
			},
			exp: `"level":"info".+"ts":\d+.\d+,"msg":.+,"uri":.+,"method":.+,"code":404,"delay":\d+.\d+,"size":\d+`,
		},
		{
			name: "Code: Internal Server Error",
			args: args{
				handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}),
				method: http.MethodGet,
				path:   "/",
			},
			exp: `"level":"info".+"ts":\d+.\d+,"msg":.+,"uri":.+,"method":.+,"code":500,"delay":\d+.\d+,"size":\d+`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := WithLogger()
			middleware := fn(tt.args.handler)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(tt.args.method, tt.args.path, nil)
			middleware.ServeHTTP(rr, req)

			content := string(wrt.content)

			exp := regexp.MustCompile(tt.exp)
			assert.Regexp(t, exp, content)
		})
	}
}
