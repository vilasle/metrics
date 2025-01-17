package postgresql

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
)

type PostgresqlMetricRepository struct {
	db     *pgxpool.Pool
	repeat []time.Duration
}

func NewRepository(db *pgxpool.Pool) (*PostgresqlMetricRepository, error) {
	r := &PostgresqlMetricRepository{
		db:     db,
		repeat: []time.Duration{time.Second * 1, time.Second * 3, time.Second * 5},
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
	if _, err := r.db.Exec(ctx, createTableTxt()); err != nil {
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
