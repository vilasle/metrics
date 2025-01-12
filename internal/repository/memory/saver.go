package memory

import (
	"sync"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
)

type saver interface {
	save(metric.Metric) error
}

type gaugeSaver struct {
	storage gaugeStorage
	mx      *sync.Mutex
}

func (s gaugeSaver) save(entity metric.Metric) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.storage[entity.Name()] = entity

	return nil
}

type counterSaver struct {
	storage counterStorage
	mx      *sync.Mutex
}

func (s counterSaver) save(entity metric.Metric) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	name := entity.Name()
	if _, ok := s.storage[name]; !ok {
		s.storage[name] = make([]metric.Metric, 0, 1)
	}

	s.storage[name] = append(s.storage[name], entity)

	return nil
}

type unknownSaver struct{}

func (s unknownSaver) save(entity metric.Metric) error {
	return repository.ErrUnknownMetricType
}
