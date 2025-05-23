package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
)

type RequestMaker interface {
	Make(objects ...metric.Metric) (*http.Request, error)
}

type SenderOption func(*HTTPSender)

type HTTPSender struct {
	maker  RequestMaker
	client http.Client
	//background sending
	rateLimit int
	reqCh     chan metric.Metric
	respCh    chan error
}

func WithRateLimit(limit int) SenderOption {
	return func(e *HTTPSender) {
		e.rateLimit = limit
		e.reqCh = make(chan metric.Metric, limit)
		e.respCh = make(chan error, limit)
	}
}

func NewHTTPSender(rm RequestMaker, opts ...SenderOption) *HTTPSender {
	s := &HTTPSender{
		maker:  rm,
		client: http.Client{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *HTTPSender) Start(ctx context.Context, wg *sync.WaitGroup) {
	go s.startWorkers(ctx, wg)
}

func (s *HTTPSender) Send(objects ...metric.Metric) error {
	if s.rateLimit > 0 {
		return s.sendAsync(objects...)
	}
	return s.sendSync(objects...)
}

func (s *HTTPSender) sendSync(objects ...metric.Metric) error {
	req, err := s.maker.Make(objects...)
	if err != nil {
		return err
	}

	if err := s.send(req); err != nil {
		return errors.Join(err, fmt.Errorf("failed metrics: %s", objects))
	}
	return nil
}

func (s *HTTPSender) sendAsync(objects ...metric.Metric) error {
	limit := s.rateLimit
	errs := make([]error, 0)

	for _, v := range objects {
		s.reqCh <- v
		limit--

		if limit > 0 {
			continue
		}

		for i := 0; i < s.rateLimit; i++ {
			errs = append(errs, <-s.respCh)
		}
		limit = s.rateLimit
	}

	return errors.Join(errs...)
}

func (s *HTTPSender) send(req *http.Request) error {
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	switch statusCode {
	case http.StatusNotFound:
		err = errors.Join(
			ErrWrongMetricName,
			fmt.Errorf("status code %d", statusCode))
	case http.StatusBadRequest:
		err = errors.Join(
			ErrWrongMetricTypeOrValue,
			fmt.Errorf("status code %d", statusCode))
	}

	return err

}

func (s *HTTPSender) startWorkers(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(s.rateLimit)

	for qty := s.rateLimit; qty > 0; qty-- {
		go s.background(ctx, wg)
	}
}

func (s *HTTPSender) background(ctx context.Context, wg *sync.WaitGroup) {
	logger.Info("run worker")
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			for r := range s.reqCh {
				s.respCh <- s.sendSync(r)
			}
			return
		case m := <-s.reqCh:
			logger.Info("got metrics", "metric", m)
			s.respCh <- s.sendSync(m)
		}
	}
}
