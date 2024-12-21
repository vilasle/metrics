package json

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service/agent/sender/rest"
)

type HTTPJsonSender struct {
	*url.URL
	httpClient
}

func NewHTTPJsonSender(addr string) (HTTPJsonSender, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return HTTPJsonSender{}, err
	}
	return HTTPJsonSender{URL: u, httpClient: newClient(true)}, nil
}

func (s HTTPJsonSender) Send(value metric.Metric) error {
	u := *s.URL
	content, err := value.ToJSON()
	if err != nil {
		return err
	}

	req, err := s.NewRequest(http.MethodPost, u.String(), content)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	switch statusCode {
	case http.StatusNotFound:
		err = errors.Join(
			rest.ErrWrongMetricName,
			fmt.Errorf("status code %d. metric = %s",
				statusCode, string(content)))
	case http.StatusBadRequest:
		err = errors.Join(
			rest.ErrWrongMetricTypeOrValue,
			fmt.Errorf("status code %d. metric = %s",
				statusCode, string(content)))
	}

	return err
}
