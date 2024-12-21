package rest

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/vilasle/metrics/internal/compress"
	"go.uber.org/zap"
)

func allowedMethods(method ...string) func(h http.Handler) http.Handler {
	allowedMethods := make(map[string]struct{}, len(method))
	for _, m := range method {
		allowedMethods[strings.ToUpper(strings.TrimSpace(m))] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if len(allowedMethods) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			if _, ok := allowedMethods[r.Method]; ok {
				next.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return http.HandlerFunc(fn)
	}
}

// copy past from chi middleware but chi exec this if content-length is more than 0, but I want exec it every request
func allowedContentType(contentTypes ...string) func(h http.Handler) http.Handler {
	allowedContentTypes := make(map[string]struct{}, len(contentTypes))
	for _, ctype := range contentTypes {
		allowedContentTypes[strings.TrimSpace(strings.ToLower(ctype))] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if len(allowedContentTypes) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			s := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
			if i := strings.Index(s, ";"); i > -1 {
				s = s[0:i]
			}

			if _, ok := allowedContentTypes[s]; ok {
				next.ServeHTTP(w, r)
				return
			}

			w.WriteHeader(http.StatusUnsupportedMediaType)
		}
		return http.HandlerFunc(fn)
	}
}

// TODO hide zap logger under interface Logger
func WithLogger(logger *zap.SugaredLogger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				logger.Infow(
					"handle request",
					zap.String("uri", r.RequestURI),
					zap.String("method", r.Method),
					zap.Int("code", ww.Status()),
					zap.Duration("delay", time.Since(start)),
					zap.Int("size", ww.BytesWritten()),
				)
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}

func WithCompress(types ...string) func(h http.Handler) http.Handler {
	poll := &sync.Pool{
		New: func() interface{} {
			return compress.NewCompressor(gzip.BestCompression)
		},
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			compressAccept := r.Header.Get("Accept-Encoding")
			if !strings.Contains(compressAccept, "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			cw := poll.Get().(compress.CompressorWriter)
			defer poll.Put(cw)

			crw := &compressedResponseWriter{
				ResponseWriter: w,
				w:              cw,
				types:          types,
			}
			defer crw.Close()

			w.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(crw, r)
		}
		return http.HandlerFunc(fn)
	}

}

type compressedResponseWriter struct {
	http.ResponseWriter
	w     compress.CompressorWriter
	types []string
}

func (cw *compressedResponseWriter) Write(b []byte) (int, error) {
	if cw.w == nil {
		cw.writeAsIs(b)
	}

	if len(cw.types) == 0 {
		return cw.w.Write(b)
	}

	for _, t := range cw.types {
		if strings.Contains(cw.Header().Get("Content-Type"), t) {
			return cw.w.Write(b)
		}
	}

	return cw.writeAsIs(b)
}

func (cw *compressedResponseWriter) writeAsIs(b []byte) (int, error) {
	cw.ResponseWriter.Header().Del("Content-Encoding")
	return cw.ResponseWriter.Write(b)
}

func (cw *compressedResponseWriter) Close() error {
	if cw.w == nil {
		return nil
	}
	_, err := cw.ResponseWriter.Write(cw.w.Bytes())
	return err
}
