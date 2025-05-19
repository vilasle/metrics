package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vilasle/metrics/internal/metric"
)

type JSONRequestMaker struct {
	addr          *url.URL
	contentWriter JSONWriter
}

func NewJSONRequestMaker(addr string, writer JSONWriter) (*JSONRequestMaker, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	maker := &JSONRequestMaker{
		addr:          u,
		contentWriter: writer,
	}

	return maker, nil
}

func (maker *JSONRequestMaker) Make(objects ...metric.Metric) (*http.Request, error) {

	var err error
	if len(objects) == 1 {
		err = maker.contentWriter.Write(objects[0])
	} else {
		err = maker.contentWriter.Write(objects)
	}

	if err != nil {
		return nil, err
	}

	rd := bytes.NewReader(maker.contentWriter.Bytes())

	req, err := http.NewRequest(http.MethodPost, maker.addr.String(), rd)
	if err != nil {
		return nil, err
	}

	for k, v := range maker.contentWriter.headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Accept-Encoding", "gzip") //TODO why????

	return req, nil
}

type TextRequestMaker struct {
	addr *url.URL
}

func NewTextRequestMaker(addr string) (*TextRequestMaker, error) {
	maker := &TextRequestMaker{}

	u, err := url.Parse(addr)

	maker.addr = u

	return maker, err
}

func (maker *TextRequestMaker) Make(objects ...metric.Metric) (*http.Request, error) {
	if len(objects) < 1 {
		return nil, fmt.Errorf("objects does not have metrics")
	}

	if len(objects) > 1 {
		return nil, fmt.Errorf("this maker does not support work with multiple metrics")
	}

	metric := objects[0]

	addr := maker.addr.JoinPath(metric.Type(), metric.Name(), metric.Value()).String()

	req, err := http.NewRequest(http.MethodPost, addr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/plain")

	return req, nil
}
