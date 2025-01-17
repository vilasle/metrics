package postgresql

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
)

type repeater struct {
	db     *pgxpool.Pool
	repeat []time.Duration
}

func (r repeater) Exec(ctx context.Context, sql string, args ...interface{}) (err error) {
	for _, d := range r.repeat {
		if _, err = r.db.Exec(ctx, sql, args...); err == nil {
			return nil
		}
		time.Sleep(d)
	}
	return err
}

func (r repeater) Query(ctx context.Context, sql string, args ...interface{}) (rows pgx.Rows, err error) {
	for _, d := range r.repeat {
		if rows, err = r.db.Query(ctx, sql, args...); err == nil {
			return rows, nil
		}
		time.Sleep(d)
	}
	return nil, err
}

func (r repeater) Ping(ctx context.Context) (err error) {
	for _, d := range r.repeat {
		if err = r.db.Ping(ctx); err == nil {
			return nil
		}
		time.Sleep(d)
	}
	return err
}

func (r repeater) Close() {
	r.db.Close()
}

func (r repeater) Begin(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

type PostgresqlMetricRepository struct {
	db repeater
}

func NewRepository(db *pgxpool.Pool) (*PostgresqlMetricRepository, error) {
	r := &PostgresqlMetricRepository{
		db: repeater{
			db:     db,
			repeat: []time.Duration{time.Second * 1, time.Second * 3, time.Second * 5},
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

func (r *PostgresqlMetricRepository) Save(entity ...metric.Metric) error {
	switch len(entity) {
	case 0:
		return repository.ErrEmptySetOfMetric
	case 1:
		e := entity[0]
		return r.getSaver(e.Type()).save(e)
	default:
		//TODO wrap it
		return r.saveAll(entity...)
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

func (r *PostgresqlMetricRepository) saveAll(entity ...metric.Metric) error {
	tx, err := r.db.Begin(context.TODO())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	errs := make([]error, 0)

	for _, e := range entity {
		if err := r.getSaver(e.Type()).save(e); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return tx.Commit(context.Background())
}

func (r *PostgresqlMetricRepository) Get(metricType string, filterName ...string) ([]metric.Metric, error) {
	return r.getGetter(metricType).get(filterName...)
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
