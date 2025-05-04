package text

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service/agent/sender/rest"
)

// HTTPTextSender struct for preparing http request. Metric passes as url of specific format
type HTTPTextSender struct {
	*url.URL
	client http.Client
}

// NewHTTPTextSender creates new HTTPTextSender or returns error if addr is not valid
func NewHTTPTextSender(addr string) (HTTPTextSender, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return HTTPTextSender{}, err
	}
	return HTTPTextSender{URL: u, client: http.Client{Timeout: time.Second * 5}}, nil
}

// Send prepares request and send metric to server
func (s HTTPTextSender) Send(value metric.Metric) error {
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

		err = errors.Join(rest.ErrWrongMetricName, fmt.Errorf("status code %d", statusCode))
	case http.StatusBadRequest:
		err = errors.Join(rest.ErrWrongMetricTypeOrValue, fmt.Errorf("status code %d", statusCode))
	}

	return err
}
