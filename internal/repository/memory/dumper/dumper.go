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

type initOpt func(context.Context, *FileDumper) error

//TODO add godoc
type Config struct {
	Timeout time.Duration
	Restore bool
	Storage repository.MetricRepository
	SerialWriter
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

//TODO add godoc
type FileDumper struct {
	fs       SerialWriter
	storage  repository.MetricRepository
	syncSave bool
	srvMx    *sync.Mutex
}

//TODO add godoc
/*
On during work on sync mode we will add only new lines to file
example:

	0;gauge1;123
	0;gauge1;124
	0;gauge1;126
	1;counter1;126
	1;counter1;126
	1;counter1;126

For counter such situation is ok, for gauge not is.

In general if the last launch worked on sync mode we would have all history transactions.
When we restore repository from file we will have unique values for gauge and historical data for counter
and after dumping got file which will match real situation on repository
*/
func NewFileDumper(ctx context.Context, config Config) (*FileDumper, error) {
	d := &FileDumper{
		storage:  config.Storage,
		fs:       config.SerialWriter,
		srvMx:    &sync.Mutex{},
		syncSave: config.Timeout == 0,
	}

	if err := d.prepareToRun(ctx, d.getInitOpts(config)...); err != nil {
		return nil, err
	}

	d.runByMode(ctx, config.Timeout)

	return d, nil
}

//TODO add godoc
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

//TODO add godoc
func (d *FileDumper) DumpAll(ctx context.Context) error {
	d.srvMx.Lock()
	defer d.srvMx.Unlock()

	s, err := d.all(ctx)
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

	logger.Debugw("content on dump before dumping", zap.String("content", buf.String()))

	_, err = d.fs.Rewrite(buf.Bytes())
	return err
}

//TODO add godoc
func (d *FileDumper) Get(ctx context.Context, metricType string, filterName ...string) ([]metric.Metric, error) {
	return d.storage.Get(ctx, metricType, filterName...)
}

//TODO add godoc
func (d *FileDumper) Ping(ctx context.Context) error {
	return d.storage.Ping(ctx)
}

//TODO add godoc
func (d *FileDumper) Close() {
	d.storage.Close()
}

func (d *FileDumper) restore(ctx context.Context) error {
	all, err := d.fs.ScanAll()
	if err != nil {
		return err
	}

	logger.Debugw("content on dump before restore", zap.Any("dump", all))

	errs := make([]error, 0)

	rawGauge := make(map[string]metric.Metric)
	rawCounter := make([]metric.Metric, 0)

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
	qty := len(rawGauge) + len(rawCounter)
	for _, g := range rawGauge {
		if err := d.storage.Save(ctx, g); err != nil {
			errs = append(errs, err)
			qty--
		}
	}

	for _, c := range rawCounter {
		if err := d.storage.Save(ctx, c); err != nil {
			errs = append(errs, err)
			qty--
		}
	}

	logger.Debugf("after restoring there are %d metrics", qty)

	return errors.Join(errs...)
}

func (d *FileDumper) getInitOpts(config Config) []initOpt {
	if config.Restore {
		return []initOpt{withRestore}
	}
	return []initOpt{withClear}
}

func (d *FileDumper) prepareToRun(ctx context.Context, opts ...initOpt) error {
	for _, opt := range opts {
		if err := opt(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

func (d *FileDumper) runByMode(ctx context.Context, timeout time.Duration) {
	if !d.syncSave {
		go d.dumpOnBackground(ctx, timeout)
	}
}

func (d *FileDumper) dumpOnBackground(ctx context.Context, timeout time.Duration) {
	ticker := time.NewTicker(timeout)
	for {
		select {
		case <-ticker.C:
			d.DumpAll(ctx)
		case <-ctx.Done():
			logger.Debug("got signal. need to dump all metrics")
			d.stop(ctx)
			return
		}
	}
}

func (d *FileDumper) stop(ctx context.Context) {
	if err := d.DumpAll(ctx); err != nil {
		logger.Error("error on dump all metrics", zap.Error(err))
	}
	defer d.fs.Close()
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

func withRestore(ctx context.Context, d *FileDumper) error {
	if err := d.restore(ctx); err != nil {
		return err
	}

	if err := d.DumpAll(ctx); err != nil {
		return err
	}
	return nil
}

func withClear(ctx context.Context, d *FileDumper) error {
	return d.fs.Clear()
}
