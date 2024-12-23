package dumper

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vilasle/metrics/internal/model"
	"github.com/vilasle/metrics/internal/repository"
)

type CounterDumper struct {
	repository.MetricRepository[model.Counter]
	syncSave     bool
	timeout      time.Duration
	fs           *FileStream
	repositoryMx *sync.Mutex
}

func NewCounterDumper(
	repository repository.MetricRepository[model.Counter],
	syncSave bool,
	timeout time.Duration,
	restore bool,
	fs *FileStream) *CounterDumper {

	d := &CounterDumper{
		MetricRepository: repository,
		syncSave:         syncSave,
		timeout:          timeout,
		fs:               fs,
		repositoryMx:     &sync.Mutex{},
	}

	if restore {
		d.Restore()
		d.Dump()
	}

	if timeout > 0 {
		go d.DumpOnBackground()
	}

	return d
}

func (d *CounterDumper) Save(name string, value model.Counter) error {
	d.repositoryMx.Lock()
	defer d.repositoryMx.Unlock()
	if !d.syncSave {
		return d.MetricRepository.Save(name, value)
	}

	return d.saveWithDump(name, value)
}

// restore metrics
func (d *CounterDumper) Restore() error {
	d.repositoryMx.Lock()
	defer d.repositoryMx.Unlock()

	all, err := d.fs.ScanAll()
	if err != nil {
		return err
	}

	for _, b := range all {
		if !strings.HasPrefix(b, "1") {
			continue
		}

		raw := strings.Split(b, ";")

		if len(raw) != 3 {
			continue
		}

		name := raw[1]
		value, err := strconv.ParseInt(raw[2], 10, 64)
		if err != nil {
			continue
		}

		d.MetricRepository.Save(name, model.Counter(value))
	}
	return nil
}

// save metrics
func (d *CounterDumper) Dump() error {
	d.repositoryMx.Lock()

	all, err := d.MetricRepository.AllAsIs()
	if err != nil {
		d.repositoryMx.Unlock()
		return err
	}
	d.repositoryMx.Unlock()

	allContent := bytes.Buffer{}
	for name, values := range all {
		for i := range values {
			allContent.Write(d.getMetricValue(name, values[i]))
		}
	}

	if _, err = d.fs.Rewrite(allContent.Bytes()); err == nil {
		return d.fs.Flush()
	}
	return err
}

func (d *CounterDumper) saveWithDump(name string, value model.Counter) error {
	err := d.MetricRepository.Save(name, value)
	if err != nil {
		return err
	}

	b := d.getMetricValue(name, value)

	if _, err = d.fs.Write(b); err == nil {
		return d.fs.Flush()
	}
	return err
}

func (d *CounterDumper) getMetricValue(name string, value model.Counter) []byte {
	return []byte(fmt.Sprintf("1;%s;%d\n", name, value))
}

func (d *CounterDumper) DumpOnBackground() {
	ticker := time.NewTicker(d.timeout)
	for {
		<-ticker.C
		d.Dump()
		//TODO cancel function
	}
}
