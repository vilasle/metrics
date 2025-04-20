package json

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/vilasle/metrics/internal/compress"
)

type httpClient struct {
	client         http.Client
	useCompression bool
	encodersPool   *sync.Pool
}

func (h httpClient) newRequest(method, url string, body []byte) (*http.Request, error) {
	if body == nil {
		return http.NewRequest(method, url, nil)
	}

	var (
		rd      io.Reader
		headers = make(map[string]string)
		req     *http.Request
		err     error
	)

	if h.useCompression {
		wrt := h.encodersPool.Get().(compress.CompressorWriter)

		_, err := wrt.Write(body)
		if err != nil {
			return nil, err
		}
		rd = bytes.NewReader(wrt.Bytes())
		headers["Content-Encoding"] = "gzip"

		wrt.Reset()
		h.encodersPool.Put(wrt)
	} else {
		rd = bytes.NewReader(body)
	}

	if req, err = http.NewRequest(method, url, rd); err == nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	return req, err
}

func newClient(useCompression bool) httpClient {
	return httpClient{
		client:         http.Client{Timeout: time.Second * 5},
		useCompression: useCompression,
		encodersPool: &sync.Pool{
			New: func() interface{} {
				return compress.NewCompressor(gzip.BestCompression)
			},
		},
	}
}