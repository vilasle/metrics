package postgresql

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vilasle/metrics/internal/metric"
	"github.com/vilasle/metrics/internal/repository"
)

const (
	gaugeInx   = 1
	counterInx = 2
)

type PostgresqlMetricRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) (*PostgresqlMetricRepository, error) {
	r := &PostgresqlMetricRepository{db: db}
	ctx := context.Background()
	//TODO use ctx with timeout
	err := r.Ping(ctx)
	if err != nil {
		return nil, err
	}
	err = r.initMetadata(ctx)
	return r, err
}

func (r *PostgresqlMetricRepository) Save(entity metric.Metric) error {
	return r.getSaver(entity.Type()).save(entity)
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

	if _, err := r.db.Exec(ctx, createIndexesTxt()); err != nil {
		return errors.Join(repository.ErrInitializeMetadata, err)
	}
	return nil
}

func createTableTxt() string {
	return `
	CREATE TABLE IF NOT EXISTS metrics (
    	"type" SMALLINT NOT NULL,
    	"name" VARCHAR(100) NOT NULL,
    	"value" DOUBLE PRECISION NULL,
    	"delta" BIGINT NULL,
    	"created_at" TIMESTAMP NOT NULL,
    	
		UNIQUE("type", "name"),
    	
		CONSTRAINT mapping_type
    	-- 1 = gauge, gauge fields is "value"; 2 = counter, counter field is "delta"
    	CHECK ( ("type" BETWEEN 1 AND 2) AND ( 
				("type" = 1 AND "value" IS NOT NULL AND "delta" IS NULL ) OR 
				("type" = 2 AND "delta" IS NOT NULL AND "value" IS NULL)
    	    )
    	)
	);`
}

func createIndexesTxt() string {
	return `
	CREATE INDEX IF NOT EXISTS metrics_created_at_idx ON metrics ("type", "name");
	`
}
