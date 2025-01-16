package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
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
	db *pgxpool.Pool
}

func (s gaugeSaver) save(m metric.Metric) error {
	_, err := s.db.Exec(context.TODO(), s.saveTxt(), gaugeInx, m.Name(), m.Float64())
	return err
}

func (s gaugeSaver) saveTxt() string {
	return `
	INSERT INTO metrics ("type", "name", "value", "created_at")
	VALUES ($1, $2, $3, now())
	ON CONFLICT ("type", "name") DO UPDATE 
	SET "value" = EXCLUDED."value", "created_at" = now();
	`
}

type counterSaver struct {
	db *pgxpool.Pool
}

func (s counterSaver) save(m metric.Metric) error {
	_, err := s.db.Exec(context.TODO(), s.saveTxt(), counterInx, m.Name(), m.Int64())
	return err
}

func (s counterSaver) saveTxt() string {
	return `
	INSERT INTO metrics ("type", "name", "delta", "created_at")
	VALUES ($1, $2, $3, now())
	ON CONFLICT ("type", "name") DO UPDATE 
	SET "delta" = EXCLUDED."delta", "created_at" = now();
	`
}
