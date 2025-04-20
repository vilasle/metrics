package json

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"crypto/hmac"
	"crypto/sha256"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service/agent/sender/rest"
)

//TODO add godoc
type HTTPJsonSender struct {
	*url.URL
	httpClient
	hashSumKey string
	req        chan metric.Metric
	resp       chan error
	rateLimit  int
}

//TODO add godoc
func NewHTTPJsonSender(addr string, hashSumKey string, rateLimit int) (HTTPJsonSender, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return HTTPJsonSender{}, err
	}
	s := HTTPJsonSender{
		URL:        u,
		httpClient: newClient(false),
		hashSumKey: hashSumKey,
		req:        make(chan metric.Metric, rateLimit),
		resp:       make(chan error, rateLimit),
		rateLimit:  rateLimit,
	}

	s.runWorkers(rateLimit)

	return s, nil
}

//TODO add godoc
func (s HTTPJsonSender) Send(value metric.Metric) error {
	u := *s.URL
	content, err := prepareBodyForReport(value)
	if err != nil {
		return err
	}

	req, err := s.newRequest(http.MethodPost, u.String(), content)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	if err := s.addHashSumHeader(req, &content); err != nil {
		//if could not create hash-sum, but it does not break main logic
		//that's why we continue and allow to the server decide take this report or no
		logger.Error("can not create hash-sum", "err", err)
	}

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

//TODO add godoc
func (s HTTPJsonSender) SendWithLimit(value ...metric.Metric) error {
	limit := s.rateLimit
	errs := make([]error, 0)

	for _, v := range value {
		s.req <- v
		limit--

		if limit > 0 {
			continue
		}

		for i := 0; i < s.rateLimit; i++ {
			errs = append(errs, <-s.resp)
		}
		limit = s.rateLimit
	}

	return errors.Join(errs...)
}

//TODO add godoc
func (s HTTPJsonSender) SendBatch(values ...metric.Metric) error {
	u := *s.URL
	content, err := prepareBatchBodyForReport(values...)
	if err != nil {
		return err
	}

	req, err := s.newRequest(http.MethodPost, u.String(), content)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	if err := s.addHashSumHeader(req, &content); err != nil {
		//if could not create hash-sum, but it does not break main logic
		//that's why we continue and allow to the server decide take this report or no
		logger.Error("can not create hash-sum", "err", err)
	}

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

func (s HTTPJsonSender) runWorkers(workersQty int) {
	rate := workersQty
	if rate < 1 {
		rate = 1
	}

	for i := 0; i < rate; i++ {
		go s.runWorker()
	}
}

func (s HTTPJsonSender) runWorker() {
	for value := range s.req {
		s.resp <- s.Send(value)
	}
}

func (s HTTPJsonSender) addHashSumHeader(req *http.Request, pC *[]byte) error {
	content := *pC

	if s.hashSumKey == "" {
		return nil
	}
	w := hmac.New(sha256.New, []byte(s.hashSumKey))
	if _, err := w.Write(content); err != nil {
		return err
	}

	srcHash := w.Sum(nil)
	hash := base64.URLEncoding.EncodeToString(srcHash)

	req.Header.Add("HashSHA256", hash)
	logger.Debug("request src hash-sum", "val", srcHash)
	logger.Debug("request base64 hash-sum", "val", hash)

	return nil
}

func prepareBodyForReport(value metric.Metric) ([]byte, error) {
	return json.Marshal(value)
}

func prepareBatchBodyForReport(value ...metric.Metric) ([]byte, error) {
	return json.Marshal(value)
}
