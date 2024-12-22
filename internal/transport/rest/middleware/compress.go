package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var defaultCompressibleContentTypes = []string{
	"text/html",
	"text/css",
	"text/plain",
	"text/javascript",
	"application/javascript",
	"application/x-javascript",
	"application/json",
	"application/atom+xml",
	"application/rss+xml",
	"image/svg+xml",
}

func Compress(types ...string) func(next http.Handler) http.Handler {
	c := newCompressor(types...)
	return c.handler
}

type writerResetter interface {
	io.Writer
	Reset(w io.Writer)
}

type compressor struct {
	types        map[string]struct{}
	encodersPool *sync.Pool
}

func newCompressor(types ...string) compressor {
	c := compressor{
		types: make(map[string]struct{}),
		encodersPool: &sync.Pool{
			New: func() interface{} {
				return newGzip()
			},
		},
	}

	if len(types) == 0 {
		types = defaultCompressibleContentTypes
	}

	for _, t := range types {
		c.types[t] = struct{}{}
	}

	return c
}

func (c *compressor) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoder, encoding, cleanup := c.selectEncoder(r.Header, w)

		cw := &compressedResponse{
			ResponseWriter: w,
			w:              w,
			contentTypes:   c.types,
			encoding:       encoding,
			compressible:   false,
		}
		if encoder != nil {
			cw.w = encoder
		}
		defer cleanup()
		defer cw.Close()

		next.ServeHTTP(cw, r)
	})
}

func (c *compressor) selectEncoder(h http.Header, w io.Writer) (io.Writer, string, func()) {
	header := h.Get("Accept-Encoding")

	accepted := strings.Contains(header, "gzip")
	if !accepted {
		return nil, "", func() {}
	}

	encoder := c.encodersPool.Get().(writerResetter)
	cleanup := func() {
		c.encodersPool.Put(encoder)
	}
	encoder.Reset(w)
	return encoder, "gzip", cleanup
}

func newGzip() io.Writer {
	gw, err := gzip.NewWriterLevel(io.Discard, gzip.BestCompression)
	if err != nil {
		return nil
	}
	return gw
}

type compressedResponse struct {
	headerIsWrote bool
	http.ResponseWriter
	w            io.Writer
	contentTypes map[string]struct{}
	encoding     string
	compressible bool
}

func (cw *compressedResponse) isCompressible() bool {
	contentType := cw.Header().Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = contentType[0:idx]
	}

	if _, ok := cw.contentTypes[contentType]; ok {
		return true
	}
	return false
}

func (cw *compressedResponse) WriteHeader(code int) {
	if cw.headerIsWrote {
		cw.ResponseWriter.WriteHeader(code)
		return
	}
	cw.headerIsWrote = true
	defer cw.ResponseWriter.WriteHeader(code)

	if cw.Header().Get("Content-Encoding") != "" {
		return
	}

	if !cw.isCompressible() {
		cw.compressible = false
		return
	}

	if cw.encoding != "" {
		cw.compressible = true
		cw.Header().Set("Content-Encoding", cw.encoding)
		cw.Header().Del("Content-Length")
	}
}

func (cw *compressedResponse) Write(p []byte) (int, error) {
	if !cw.headerIsWrote {
		cw.WriteHeader(http.StatusOK)
	}

	return cw.getWriter().Write(p)
}

func (cw *compressedResponse) getWriter() io.Writer {
	if cw.compressible {
		return cw.w
	}
	return cw.ResponseWriter
}

func (cw *compressedResponse) Close() error {
	if c, ok := cw.getWriter().(io.WriteCloser); ok {
		return c.Close()
	}
	return fmt.Errorf("response writer does not implement io.WriteCloser")
}
