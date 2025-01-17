package postgresql

import (
	"context"

	"github.com/vilasle/metrics/internal/metric"
)

type saver interface {
	save(metric.Metric) error
}

type unknownSaver struct{}

func (s unknownSaver) save(metric.Metric) error {
	return metric.ErrUnknownMetricType
}

type gaugeSaver struct {
	db repeater
}

func (s gaugeSaver) save(m metric.Metric) error {
	return s.db.Exec(context.TODO(), s.saveTxt(), m.Name(), m.Float64())
}

func (s gaugeSaver) saveTxt() string {
	return `
	INSERT INTO gauges ("id", "value")
	VALUES ($1, $2) 
	ON CONFLICT ("id") DO UPDATE SET "value" = EXCLUDED."value";
	`
}

type counterSaver struct {
	db repeater
}

func (s counterSaver) save(m metric.Metric) error {
	return s.db.Exec(context.TODO(), s.saveTxt(), m.Name(), m.Int64())
}

func (s counterSaver) saveTxt() string {
	return `
	INSERT INTO counters ("id", "value", "created_at")
	VALUES ($1, $2, now())
	`
}
