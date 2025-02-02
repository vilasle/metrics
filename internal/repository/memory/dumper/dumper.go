package dumper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vilasle/metrics/internal/logger"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
	"go.uber.org/zap"
)

const (
	gaugeID   = "0"
	counterID = "1"
)

type Config struct {
	Timeout time.Duration
	Restore bool
	Storage repository.MetricRepository
	Stream  *FileStream
}

type dumpedMetric struct {
	metric.Metric
}

var ErrWrongDumpedLine = fmt.Errorf("wrong dumped line")

func (d dumpedMetric) dumpedContent() []byte {
	var (
		kind        = 0
		name, value = d.Name(), d.Value()
	)

	if d.Type() == metric.TypeCounter {
		kind = 1
	}

	c := fmt.Sprintf("%d;%s;%s\n", kind, name, value)
	return []byte(c)
}

type FileDumper struct {
	fs       *FileStream
	storage  repository.MetricRepository
	syncSave bool
	srvMx    *sync.Mutex
}

func NewFileDumper(ctx context.Context, config Config) (*FileDumper, error) {
	d := &FileDumper{
		storage: config.Storage,
		fs:      config.Stream,
		srvMx:   &sync.Mutex{},
	}
	/*
		on during work on sync mode we will add only new lines to file
		example
			0;gauge1;123
			0;gauge1;124
			0;gauge1;126
			1;counter1;126
			1;counter1;126
			1;counter1;126
		for counter such situation is ok, for gauge not is.

		In general if the last launch worked on sync mode we would have all history transactions.
		When we restore repository from file we will have unique values for gauge and historical data for counter
		and after dumping got file which will match real situation on repository
	*/

	if config.Restore {
		if err := d.restore(); err != nil {
			return nil, err
		}

		if err := d.DumpAll(); err != nil {
			return nil, err
		}
	} else {
		if err := d.fs.Clear(); err != nil {
			return nil, err
		}
	}

	if config.Timeout == 0 {
		d.syncSave = true
	} else {
		go d.dumpOnBackground(ctx, config.Timeout)
	}

	return d, nil
}

func (d *FileDumper) Save(ctx context.Context, entity ...metric.Metric) error {
	d.srvMx.Lock()
	defer d.srvMx.Unlock()
	if err := d.storage.Save(ctx, entity...); err != nil {
		return err
	}
	if !d.syncSave {
		return nil
	}

	errs := make([]error, 0)
	for _, m := range entity {
		dm := dumpedMetric{m}
		c := dm.dumpedContent()
		if _, err := d.fs.Write(c); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (d *FileDumper) DumpAll() error {
	d.srvMx.Lock()
	defer d.srvMx.Unlock()

	s, err := d.all(context.Background())
	if err != nil {
		return err
	}

	buf := bytes.Buffer{}
	for _, m := range s {
		dm := dumpedMetric{m}
		c := dm.dumpedContent()
		if _, err := buf.Write(c); err != nil {
			return err
		}
	}

	logger.Debugw(
		"content on dump before dumping",
		zap.String("content", buf.String()))

	_, err = d.fs.Rewrite(buf.Bytes())
	return err
}

func (d *FileDumper) Get(ctx context.Context, metricType string, filterName ...string) ([]metric.Metric, error) {
	return d.storage.Get(ctx, metricType, filterName...)
}

func (d *FileDumper) Ping(ctx context.Context) error {
	return d.storage.Ping(ctx)
}

func (d *FileDumper) Close() {
	d.storage.Close()
}

func (d *FileDumper) dumpOnBackground(ctx context.Context, timeout time.Duration) {
	ticker := time.NewTicker(timeout)
	for {
		select {
		case <-ticker.C:
			d.DumpAll()
		case <-ctx.Done():
			logger.Debug("got signal. need to dump all metrics")
			d.DumpAll()
			logger.Debug("dumping finished")
			d.fs.Close()
			return
		}
	}
}

func (d *FileDumper) restore() error {
	all, err := d.fs.ScanAll()
	if err != nil {
		return err
	}

	logger.Debugw(
		"content on dump before restore",
		zap.Any("dump", all))

	errs := make([]error, 0)

	rawGauge := make(map[string]metric.Metric)
	rawCounter := make([]metric.Metric, 0)

	ctx := context.Background()

	for i, b := range all {
		raw := strings.Split(b, ";")

		if len(raw) != 3 {
			errs = append(errs, errors.Join(ErrWrongDumpedLine, fmt.Errorf("wrong line. offset=%d;content=%s", i, b)))
			continue
		}
		name, value := raw[1], raw[2]
		if strings.HasPrefix(b, gaugeID) {
			m, err := metric.ParseMetric(name, value, metric.TypeGauge)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			rawGauge[name] = m
		} else if strings.HasPrefix(b, counterID) {
			m, err := metric.ParseMetric(name, value, metric.TypeCounter)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			rawCounter = append(rawCounter, m)
		}
	}

	for _, g := range rawGauge {
		if err := d.storage.Save(ctx, g); err != nil {
			errs = append(errs, err)
		}
	}

	for _, c := range rawCounter {
		if err := d.storage.Save(ctx, c); err != nil {
			errs = append(errs, err)
		}
	}
	_all, err := d.all(ctx)
	if err != nil {
		return err
	}
	logger.Debugf("after restoring there are %d metrics", len(_all))

	for _, m := range _all {
		logger.Debugw("metric",
			zap.String("name", m.Name()),
			zap.String("type", m.Type()),
			zap.String("value", m.Value()),
		)
	}

	return errors.Join(errs...)
}

func (d *FileDumper) all(ctx context.Context) ([]metric.Metric, error) {
	gauges, err := d.storage.Get(ctx, metric.TypeGauge)
	if err != nil {
		return nil, err
	}

	counters, err := d.storage.Get(ctx, metric.TypeCounter)
	if err != nil {
		return nil, err
	}
	return append(gauges, counters...), nil
}
