package rest

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/vilasle/metrics/internal/metric"
)

type HTTPSender struct {
	*url.URL
	client http.Client
}

func NewHTTPSender(addr string) (HTTPSender, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return HTTPSender{}, err
	}
	return HTTPSender{URL: u, client: http.Client{Timeout: time.Second * 5}}, nil
}

func (s HTTPSender) Send(value metric.Metric) error {
	u := *s.URL
	addr := u.JoinPath(value.Type(), value.Name(), value.Value()).String()

	req, err := http.NewRequest(http.MethodPost, addr, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	switch statusCode {
	case http.StatusNotFound:
		err = errors.Join(ErrWrongMetricName, fmt.Errorf("status code %d", statusCode))
	case http.StatusBadRequest:
		err = errors.Join(ErrWrongMetricTypeOrValue, fmt.Errorf("status code %d", statusCode))
	}

	return err
}
