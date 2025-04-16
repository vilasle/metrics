package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
)

type PostgresqlMetricRepository struct {
	db repeater
}

func NewRepository(db *sql.DB) (*PostgresqlMetricRepository, error) {
	r := &PostgresqlMetricRepository{
		db: repeater{
			db:          db,
			repeatSteps: []time.Duration{time.Second * 1, time.Second * 3, time.Second * 5},
		},
	}

	ctx := context.Background()
	//TODO use ctx with timeout
	err := r.Ping(ctx)
	if err != nil {
		return nil, err
	}
	err = r.initMetadata(ctx)
	return r, err
}

func (r *PostgresqlMetricRepository) Save(ctx context.Context, entity ...metric.Metric) error {
	switch len(entity) {
	case 0:
		return repository.ErrEmptySetOfMetric
	case 1:
		e := entity[0]
		return r.getSaver(e.Type()).save(ctx, e)
	default:
		//TODO wrap it
		return r.saveAll(ctx, entity...)
	}
}

func (r *PostgresqlMetricRepository) getSaver(metricType string) saver {
	switch metricType {
	case metric.TypeGauge:
		return &gaugeSaver{db: r.db}
	case metric.TypeCounter:
		return &counterSaver{db: r.db}
	default:
		return &unknownSaver{}
	}
}

func (r *PostgresqlMetricRepository) saveAll(ctx context.Context, entity ...metric.Metric) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	errs := make([]error, 0)

	for _, e := range entity {
		if err := r.getSaver(e.Type()).save(ctx, e); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return tx.Commit()
}

func (r *PostgresqlMetricRepository) Get(ctx context.Context, metricType string, filterName ...string) ([]metric.Metric, error) {
	return r.getGetter(metricType).get(ctx, filterName...)
}

func (r *PostgresqlMetricRepository) getGetter(metricType string) getter {
	switch metricType {
	case metric.TypeGauge:
		return &gaugeGetter{db: r.db}
	case metric.TypeCounter:
		return &counterGetter{db: r.db}
	default:
		return &unknownGetter{}
	}
}

func (r *PostgresqlMetricRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

func (r *PostgresqlMetricRepository) Close() {
	r.db.Close()
}

func (r *PostgresqlMetricRepository) initMetadata(ctx context.Context) error {
	if err := r.db.Exec(ctx, createTableTxt()); err != nil {
		return errors.Join(repository.ErrInitializeMetadata, err)
	}
	return nil
}

func createTableTxt() string {
	return `
	CREATE TABLE IF NOT EXISTS gauges (
    	"value" DOUBLE PRECISION NOT NULL,
    	"id" VARCHAR(100) NOT NULL PRIMARY KEY
	);

	CREATE TABLE IF NOT EXISTS counters (
    	"value" BIGINT NOT NULL,
    	"id" VARCHAR(100) NOT NULL,
    	"created_at" TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS counter_name_idx ON counters ("id");
	`
}
