package json

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service/agent/sender/rest"
)

type HTTPJsonSender struct {
	*url.URL
	client http.Client
}

func NewHTTPJsonSender(addr string) (HTTPJsonSender, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return HTTPJsonSender{}, err
	}
	return HTTPJsonSender{URL: u, client: http.Client{Timeout: time.Second * 5}}, nil
}

func (s HTTPJsonSender) Send(value metric.Metric) error {
	u := *s.URL
	content, err := value.ToJSON()
	if err != nil {
		return err
	}

	rd := bytes.NewReader(content)

	req, err := http.NewRequest(http.MethodPost, u.String(), rd)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	switch statusCode {
	case http.StatusNotFound:
		err = errors.Join(rest.ErrWrongMetricName, fmt.Errorf("status code %d", statusCode))
	case http.StatusBadRequest:
		err = errors.Join(rest.ErrWrongMetricTypeOrValue, fmt.Errorf("status code %d", statusCode))
	}

	return err
}
