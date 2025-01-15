package postgresql

import (
	"context"
	"database/sql"

	"github.com/vilasle/metrics/internal/metric"
)

type PostgresqlMetricRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *PostgresqlMetricRepository {
	return &PostgresqlMetricRepository{db: db}
}

func (r *PostgresqlMetricRepository) Save(entity metric.Metric) error {
	//TODO implement it
	panic("not implemented")
}

func (r *PostgresqlMetricRepository) Get(metricType string, filterName ...string) ([]metric.Metric, error) {
	//TODO implement it
	panic("not implemented")
}

func (r *PostgresqlMetricRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
