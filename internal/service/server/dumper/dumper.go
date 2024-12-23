package dumper

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/service"
)

const (
	gaugeId   = "0"
	counterId = "1"
)

type dumpedMetric struct {
	metric.RawMetric
}

var ErrWrongDumpedLine = fmt.Errorf("wrong dumped line")

func (d dumpedMetric) dumpedContent() []byte {
	var (
		kind        = 0
		name, value = d.Name, d.Value
	)

	if d.Kind == "counter" {
		kind = 1
	}

	c := fmt.Sprintf("%d;%s;%s\n", kind, name, value)
	return []byte(c)
}

type FileDumper struct {
	fs       *FileStream
	svc      service.StorageService
	syncSave bool
	srvMx    *sync.Mutex
}

func NewFileDumper(
	svc service.StorageService,
	fs *FileStream,
	restore bool,
	timeout time.Duration) (*FileDumper, error) {

	d := &FileDumper{
		svc:   svc,
		fs:    fs,
		srvMx: &sync.Mutex{},
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

	if restore {
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

	if timeout == 0 {
		d.syncSave = true
	} else {
		go d.dumpOnBackground(timeout)
	}

	return d, nil
}

func (d *FileDumper) Save(m metric.RawMetric) error {
	d.srvMx.Lock()
	defer d.srvMx.Unlock()
	if err := d.svc.Save(m); err != nil {
		return err
	}
	if !d.syncSave {
		return nil
	}
	//TODO raw metric change in should be interface for wrapping internal struct
	dm := dumpedMetric{m}

	c := dm.dumpedContent()
	_, err := d.fs.Write(c)
	return err
}

func (d *FileDumper) DumpAll() error {
	d.srvMx.Lock()
	defer d.srvMx.Unlock()
	s, err := d.svc.AllMetricsAsIs()
	if err != nil {
		return err
	}

	buf := bytes.Buffer{}
	for _, m := range s {
		dm := dumpedMetric{
			metric.NewRawMetric(m.Name(), m.Type(), m.Value()),
		}
		c := dm.dumpedContent()
		if _, err := buf.Write(c); err != nil {
			return err
		}
	}

	_, err = d.fs.Rewrite(buf.Bytes())
	return err
}

func (d *FileDumper) Get(name string, kind string) (metric.Metric, error) {
	return d.svc.Get(name, kind)
}

func (d *FileDumper) AllMetrics() ([]metric.Metric, error) {
	return d.svc.AllMetrics()
}

func (d *FileDumper) AllMetricsAsIs() ([]metric.Metric, error) {
	return d.svc.AllMetricsAsIs()
}

func (d *FileDumper) dumpOnBackground(timeout time.Duration) {
	ticker := time.NewTicker(timeout)
	for {
		<-ticker.C
		d.DumpAll()
		//TODO cancel function
	}
}

func (d *FileDumper) restore() error {
	all, err := d.fs.ScanAll()
	if err != nil {
		return err
	}

	errs := make([]error, 0)

	rawGauge := make(map[string]metric.RawMetric)
	rawCounter := make([]metric.RawMetric, 0)

	for i, b := range all {
		raw := strings.Split(b, ";")

		if len(raw) != 3 {
			errs = append(errs, errors.Join(ErrWrongDumpedLine, fmt.Errorf("wrong line. offset=%d;content=%s", i, b)))
			continue
		}
		name, value := raw[1], raw[2]
		if strings.HasPrefix(b, gaugeId) {
			rawGauge[name] = metric.NewRawMetric(name, "gauge", value)
		} else if strings.HasPrefix(b, counterId) {
			rawCounter = append(rawCounter, metric.NewRawMetric(name, "counter", value))
		}
	}

	for _, g := range rawGauge {
		if err := d.svc.Save(g); err != nil {
			errs = append(errs, err)
		}
	}

	for _, c := range rawCounter {
		if err := d.svc.Save(c); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
