package postgresql

import (
	"context"

	"github.com/vilasle/metrics/internal/metric"
)

type saver interface {
	save(context.Context, metric.Metric) error
}

type unknownSaver struct{}

func (s unknownSaver) save(context.Context, metric.Metric) error {
	return metric.ErrUnknownMetricType
}

type gaugeSaver struct {
	db repeater
}

func (s gaugeSaver) save(ctx context.Context, m metric.Metric) error {
	return s.db.exec(ctx, s.saveTxt(), m.Name(), m.Float64())
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

func (s counterSaver) save(ctx context.Context, m metric.Metric) error {
	return s.db.exec(ctx, s.saveTxt(), m.Name(), m.Int64())
}

func (s counterSaver) saveTxt() string {
	return `
	INSERT INTO counters ("id", "value", "created_at")
	VALUES ($1, $2, now())
	`
}
