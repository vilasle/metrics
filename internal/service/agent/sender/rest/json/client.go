package json

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"sync"
	"time"
)

type httpClient struct {
	client         http.Client
	useCompression bool
	encodersPool   *sync.Pool
}

func (h httpClient) NewRequest(method, url string, body []byte) (*http.Request, error) {
	if h.useCompression {
		encoder := h.encodersPool.Get().(gzip.Writer)
		_, err := encoder.Write(body)
		if err != nil {
			panic(err)
		}
		err = encoder.Close()
		if err != nil {
			panic(err)
		}
		
		
		req, err := http.NewRequest(method, url, bytes.NewReader(newBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Encoding", "gzip")
		return req, nil
	}

	return http.NewRequest(method, url, bytes.NewReader(body))
}

func newClient(useCompression bool) httpClient {
	return httpClient{
		client:         http.Client{Timeout: time.Second * 5},
		useCompression: useCompression,
		encodersPool: &sync.Pool{
			New: func() interface{} {
				return encoderGzip(io.Discard, gzip.BestCompression)
			},
		},
	}
}

func (h httpClient) Do(req *http.Request) (*http.Response, error) {
	return h.client.Do(req)
}

func encoderGzip(w io.Writer, level int) io.Writer {
	gw, err := gzip.NewWriterLevel(w, level)
	if err != nil {
		return nil
	}
	return gw
}

func compress(body []byte) ([]byte, error) {

	return body, nil
}
